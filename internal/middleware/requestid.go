package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

const requestIDHeader = "X-Request-ID"

// requestIDMiddleware is the implementation of the request ID middleware.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = generateID()
		}

		ctx := r.Context()
		ctx = setRequestID(ctx, id)

		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// setRequestID stores the request ID in the context.
func setRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// generateID produces a random UUID v4 string.
func generateID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	// Set version 4 and variant bits per RFC 4122.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
