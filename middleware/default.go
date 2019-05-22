package middleware

import (
	"context"
	"net/http"

	"github.com/episub/spawn/store"
	"github.com/episub/spawn/validate"
	"github.com/episub/spawn/vars"
)

// DefaultMW Sets up items needed for most requests
// - Adds a data object to the context, used for passing data through to OPA requests
// - Sets validation context
func DefaultMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), vars.SharedData, store.NewDataStore())
		ctx = validate.SetContext(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
