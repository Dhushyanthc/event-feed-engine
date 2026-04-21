package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(redisClient *redis.Client, limit int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {

	return func(next http.HandlerFunc) http.HandlerFunc {

		return func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			// Identify user (fallback to IP if not authenticated)
			userID, ok := r.Context().Value(UserIDKey).(int64)

			var key string
			if ok {
				key = "rate:user:" + strconv.FormatInt(userID, 10)
			} else {
				key = "rate:ip:" + r.RemoteAddr
			}

			// increment counter
			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			// set expiry only on first request
			if count == 1 {
				redisClient.Expire(ctx, key, window)
			}

			if count > int64(limit) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}
