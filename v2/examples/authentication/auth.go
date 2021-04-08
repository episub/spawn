package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/episub/spawn/middleware"
	"github.com/episub/spawn/security"
	"github.com/example/todo/gnorm/public/session"
	"github.com/example/todo/gnorm/public/user"
	"github.com/example/todo/loader"
	"github.com/gofrs/uuid"
)

// Session Auth session
type Session struct {
	session.Row
}

// GetID Returns Session ID
func (s Session) GetID() string {
	return s.SessionID.String()
}

// GetExpiry Returns session expiry
func (s Session) GetExpiry() time.Time {
	return s.Expires
}

// GetUser Returns user this session is for
func (s Session) GetUser(ctx context.Context) (middleware.User, error) {
	user, err := loader.Loader.GetUser(ctx, s.UserID)
	return User{Row: user}, err
}

// Destroy Destroys this session
func (s Session) Destroy(ctx context.Context) error {
	_, err := loader.Loader.DeleteSession(ctx, s.SessionID.String())

	return err
}

// User Auth session
type User struct {
	user.Row
}

// GetID Returns User ID.
func (u User) GetID() string {
	return fmt.Sprintf("%d", u.UserID)
}

// GetInactive Returns inactive status
func (u User) GetInactive() bool {
	return false
}

// authenticateUser Checks whether a login with username and password is
// permitted
func authenticateUser(ctx context.Context, username string, password string) (middleware.User, error) {
	if len(username) == 0 || len(password) == 0 {
		return User{}, fmt.Errorf("Must provide both username and password")
	}

	// Usernames should be stored in database in lower case so that we don't
	// differentiate between coolcat and CoolCat
	username = strings.ToLower(username)

	user, err := loader.Loader.OneUser(ctx, []sq.Sqlizer{sq.Eq{user.UsernameCol: username}}, nil)

	if err != nil {
		return nil, err
	}

	return User{Row: user}, security.AuthenticateUser(ctx, []byte{}, []byte(user.Password), []byte(password))
}

// createSession Creates a new session for user
func createSession(ctx context.Context, user middleware.User) (sessionID string, expiry time.Time, err error) {
	uid, _ := strconv.Atoi(user.GetID())

	expiry = time.Now().Add(time.Hour * 7 * 24)

	var sessionUUID uuid.UUID
	sessionUUID, err = loader.Loader.CreateSession(ctx, uid, expiry)

	sessionID = sessionUUID.String()

	return
}

// getSession Fetches session with the given id
func getSession(ctx context.Context, id string) (middleware.Session, error) {
	sessionUUID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	session, err := loader.Loader.GetSession(ctx, sessionUUID)

	return Session{Row: session}, err
}
