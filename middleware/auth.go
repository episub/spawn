package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/gqlerror"
)

var log = logrus.New()

// SetLogger Provide custom logrus logger
func SetLogger(l *logrus.Logger) {
	log = l
}

// Auth Struct used for managing authorisation and authentication matters
type Auth struct {
	AuthenticateUser func(ctx context.Context, username string, password string) (User, error)
	CookieName       string
	CreateSession    func(context.Context, User) (string, time.Time, error)
	Debug            bool
	GetSession       func(context.Context, string) (Session, error)
}

// User Generic user interface used by functions, allowing projects to provide
// their own
type User interface {
	GetID() string
	GetInactive() bool
}

// Session interface allowing projects to provide their own
type Session interface {
	// Destroy will destroy the currently active session if one exists.  Should
	// not return error if there is no session
	GetUser(context.Context) (User, error)
	Destroy(context.Context) error
	GetExpiry() time.Time
	GetID() string
}

// NewAuth Returns a new Auth struct
func NewAuth(
	cookieName string,
	authenticateUser func(ctx context.Context, username string, password string) (User, error),
	createSession func(context.Context, User) (string, time.Time, error),
	getSession func(context.Context, string) (Session, error),
	debug bool,
) Auth {
	return Auth{
		CookieName:       cookieName,
		AuthenticateUser: authenticateUser,
		CreateSession:    createSession,
		GetSession:       getSession,
		Debug:            debug,
	}
}

// SessionMW Manages cookies, and puts the user and session in the context if appropriate, and returns unauthorised if session is expired or user is inactive
func (a Auth) SessionMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(a.CookieName)
		if err == nil && c != nil {
			// Cookie found:
			session, err := a.GetSession(r.Context(), c.Value)
			if err != nil {
				log.WithFields(logrus.Fields{"error": err, "session": c.Value}).Warning("Failed to fetch session from database")
				a.SetUnauthorised(w, r)
				return
			}

			// Invalidate if expired:
			if time.Now().After(session.GetExpiry()) {
				log.WithField("session", session.GetID()).Info("Session expired")
				a.SetUnauthorised(w, r)
				return
			}

			// Invalidate if user inactive
			user, err := session.GetUser(r.Context())
			if err != nil {
				log.WithFields(logrus.Fields{"error": err}).Warning("Could not fetch session user")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// If inactive, don't use them:
			if user.GetInactive() {
				a.SetUnauthorised(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), "session", session)
			ctx = a.GetAuthenticationContext(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnforceAuthenticationMW Adds authentication related information to context and rejects request if unauthenticated
func (a Auth) EnforceAuthenticationMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fetch session from context:
		user, authenticated := a.CheckAuthenticated(w, r)

		if !authenticated {
			a.SetUnauthorised(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(a.GetAuthenticationContext(r.Context(), user)))
	})
}

// GetAuthenticationContext Sets the user and the user ID in context
func (a Auth) GetAuthenticationContext(ctx context.Context, user User) context.Context {
	if user != nil {
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "created_by", user.GetID())
	}
	return ctx
}

// CheckAuthenticated Returns true if user is authenticated (present in context)
func (a *Auth) CheckAuthenticated(w http.ResponseWriter, r *http.Request) (User, bool) {
	user, ok := r.Context().Value("user").(User)

	if !ok {
		return nil, false
	}

	return user, ok
}

// SetUnauthorised Used to present a standard unauthorised response
func (a Auth) SetUnauthorised(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	gerr := gqlerror.Error{}
	gerr.Message = "Invalid or expired session"
	gerr.Extensions = map[string]interface{}{
		"code": 401,
	}

	b, err := json.Marshal(gerr)
	if err != nil {
		log.WithField("error", err).Error("Could not json encode error message")
	}
	w.Write(b)
}

// AuthenticationHandler Authenticates user and returns jwt if valid
func (a Auth) AuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "authenticationHandler")
	defer span.Finish()

	// Destroy existing session on this client, if it exists, since sessions shouldn't be shared across machines:
	a.DestroySession(r)

	var err error
	var username, password string
	var invalidLoginMsg = "Invalid username or password"

	// Grab username and password
	if len(r.URL.Query()["username"]) > 0 {
		username = r.URL.Query()["username"][0]
	}

	if len(r.URL.Query()["password"]) > 0 {
		password = r.URL.Query()["password"][0]
	}

	// If no username or password provided, request is bad
	if len(username) == 0 || len(password) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, invalidLoginMsg)
		log.Info("Attempt to authenticate with empty username or password")
		return
	}

	user, err := a.AuthenticateUser(ctx, username, password)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, invalidLoginMsg)
		log.WithFields(logrus.Fields{"error": err, "username": username}).Info("Failed to validate user password")
		return
	}

	session, expiry, err := a.CreateSession(ctx, user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, invalidLoginMsg)
		log.WithField("error", err).Error("Failed to create session")
		return
	}

	var c http.Cookie
	c.Name = a.CookieName
	c.Value = session
	c.Expires = expiry
	c.HttpOnly = true
	if !a.Debug {
		c.Secure = true
	}
	http.SetCookie(w, &c)

	w.WriteHeader(http.StatusOK)
}

// LogoutHandler Authenticates user and returns jwt if valid
func (a Auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	a.DestroySession(r)

	c := &http.Cookie{
		Name:    a.CookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}

	http.SetCookie(w, c)

	w.WriteHeader(http.StatusOK)
	return
}

// DestroySession Destroys session if one exists
func (a Auth) DestroySession(r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "destroySession")
	defer span.Finish()

	c, err := r.Cookie(a.CookieName)
	if err == nil && c != nil {
		// Cookie found:
		session, err := a.GetSession(r.Context(), c.Value)

		// Ignore errors if session can't be loaded or destroyed.  We don't guarantee session destruction, but aim to do it when possible:
		if err != nil {
			return
		}

		err = session.Destroy(ctx)
		if err != nil {
			log.WithField("error", err).Error("Failed to destroy session")
		}
	}
}
