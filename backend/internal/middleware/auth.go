package middleware

import (
	"context"
	"net/http"

	"github.com/SanthoshCheemala/FLARE/backend/internal/auth"
)

func Auth(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Expect "Bearer <token>"
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]
			claims, err := authSvc.ValidateAccessToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			userCtx := &auth.UserContext{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			}

			ctx := auth.SetUserContext(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := auth.GetUserContext(r.Context())
			if userCtx == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if userCtx.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			for _, role := range roles {
				if userCtx.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func GetUser(ctx context.Context) *auth.UserContext {
	return auth.GetUserContext(ctx)
}
