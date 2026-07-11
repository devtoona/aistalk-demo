package middleware

import (
	"net/http"
	"strings"

	"voice-chat-api-go/auth"
	"voice-chat-api-go/logger"
)

// RequireFirebaseAuth rejects requests without a valid Firebase ID token.
// Token sources:
//  1. Authorization: Bearer <token>
//  2. access_token query (for EventSource / SSE)
func RequireFirebaseAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.Disabled() {
			next.ServeHTTP(w, r.WithContext(auth.WithUID(r.Context(), "local-dev")))
			return
		}

		token := bearerToken(r)
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("access_token"))
		}
		if token == "" {
			http.Error(w, `{"error":"unauthorized","message":"missing firebase id token"}`, http.StatusUnauthorized)
			return
		}

		uid, err := auth.VerifyIDToken(r.Context(), token)
		if err != nil {
			logger.Warn("firebase token verify failed: %v", err)
			http.Error(w, `{"error":"unauthorized","message":"invalid firebase id token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithUID(r.Context(), uid)))
	})
}

func bearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
