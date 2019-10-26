package middleware

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"net/http"

	"github.com/episub/spawn/store"
	"github.com/episub/spawn/validate"
	"github.com/episub/spawn/vars"
	"github.com/h2non/filetype"
	opentracing "github.com/opentracing/opentracing-go"
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

// DeliverFile Sends a file, including information about its content type
// and an etag based on a hash of the bytes
func DeliverFile(
	ctx context.Context,
	bytes []byte,
	w http.ResponseWriter,
	r *http.Request,
) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "deliverFile")
	defer span.Finish()

	kind, err := filetype.Match(bytes)

	if err != nil {
		log.WithField("error", err).Error("Could not determine file type")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(kind.Extension) > 0 {
		w.Header().Add("Content-Type", kind.Extension)
	}
	etagRaw := md5.Sum(bytes)
	etag := base64.StdEncoding.EncodeToString(etagRaw[:])

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Write(bytes)
}
