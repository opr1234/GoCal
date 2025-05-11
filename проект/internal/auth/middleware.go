package auth

import (
    "context"
    "fmt"
    "net/http"
    "strings"
)

func Middleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                respondError(w, http.StatusUnauthorized, "Authorization header required")
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                respondError(w, http.StatusUnauthorized, "Invalid token format")
                return
            }

            userID, err := ParseToken(parts[1], secret)
            if err != nil {
                status := http.StatusUnauthorized
                if err == ErrTokenExpired {
                    status = http.StatusForbidden
                }
                respondError(w, status, fmt.Sprintf("Token validation failed: %v", err))
                return
            }

            ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        }
    }
}

func respondError(w http.ResponseWriter, code int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}

type contextKey string
const ContextKeyUserID = contextKey("user_id")