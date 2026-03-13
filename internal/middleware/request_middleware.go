package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/Dhushyanthc/event-feed-engine/internal/contextutils"
)

func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		requestID := uuid.New().String()

		ctx = context.WithValue(ctx, contextutils.RequestIDKey, requestID)

		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		next(w, r)
	}
}