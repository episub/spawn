package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/episub/spawn/v2/static"
	"github.com/episub/spawn/v2/store"
	"github.com/episub/spawn/v2/validate"
	"github.com/episub/spawn/v2/vars"
	opentracing "github.com/opentracing/opentracing-go"
)

// DefaultMW Sets up items needed for most requests
func DefaultMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := DefaultContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// DefaultContext Sets the default context needed for most requests
// - Adds a data object to the context, used for passing data through to OPA requests
// - Sets validation context
func DefaultContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, vars.SharedData, store.NewDataStore())
	ctx = validate.SetContext(ctx)

	return ctx
}

// BodyLimitMW Limits body size to the provided bytes
func BodyLimitMW(size int64) func(http.Handler) http.Handler {
	bodyLimit := size
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bodyLimit)
			next.ServeHTTP(w, r)
		})
	}
}

// DeliverFile Sends a file, including information about its content type
// and an etag based on a hash of the bytes
func DeliverFile(
	ctx context.Context,
	file static.File,
	w http.ResponseWriter,
	r *http.Request,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "deliverFile")
	defer span.Finish()

	etag, err := file.ETag()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	w.Header().Add("ETag", "\""+etag+"\"")

	name, err := file.Name()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	http.ServeContent(w, r, name, time.Time{}, file)

	return nil
}
