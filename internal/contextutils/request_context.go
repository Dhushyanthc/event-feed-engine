package contextutils

import "net/http"

type contextKey string

const RequestIDKey contextKey = "request_id"

func GetRequestID(r *http.Request) string {
	id, ok := r.Context().Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return id
}