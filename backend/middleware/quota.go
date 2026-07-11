package middleware

import (
	"encoding/json"
	"net/http"

	"voice-chat-api-go/auth"
	"voice-chat-api-go/logger"
	"voice-chat-api-go/quota"
)

// RequireQuota consumes one unit of the given quota kind after auth.
func RequireQuota(kind quota.Kind) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, ok := auth.UIDFromContext(r.Context())
			if !ok || uid == "" {
				http.Error(w, `{"error":"unauthorized","message":"missing uid"}`, http.StatusUnauthorized)
				return
			}
			if err := quota.Consume(r.Context(), uid, kind); err != nil {
				if quota.IsExceeded(err) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"error":   "quota_exceeded",
						"message": "daily limit reached",
						"kind":    string(kind),
					})
					return
				}
				logger.Error("quota consume failed: %v", err)
				http.Error(w, `{"error":"quota_unavailable","message":"failed to check quota"}`, http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
