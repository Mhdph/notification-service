package identity

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	secret []byte
}

func NewMiddleware(
	secret string,
) *Middleware {
	return &Middleware{
		secret: []byte(secret),
	}
}

func (m *Middleware) Handle(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			tokenString :=
				strings.TrimSpace(
					r.Header.Get("x-apikey"),
				)

			userID :=
				strings.TrimSpace(
					r.Header.Get("user-id"),
				)

			appID :=
				strings.TrimSpace(
					r.Header.Get("app-id"),
				)

			if tokenString == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"missing x-apikey header",
				)
				return
			}

			if userID == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"missing user-id header",
				)
				return
			}

			if appID == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"missing app-id header",
				)
				return
			}

			claims := &Claims{}

			token, err := jwt.ParseWithClaims(
				tokenString,
				claims,
				func(
					token *jwt.Token,
				) (any, error) {
					return m.secret, nil
				},
				jwt.WithValidMethods(
					[]string{
						jwt.SigningMethodHS256.Alg(),
					},
				),
				jwt.WithExpirationRequired(),
			)

			if err != nil || !token.Valid {
				writeError(
					w,
					http.StatusUnauthorized,
					"invalid or expired token",
				)
				return
			}

			if claims.UserID == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"token does not contain user_id",
				)
				return
			}

			if claims.UserID != userID {
				writeError(
					w,
					http.StatusUnauthorized,
					"user-id does not match token",
				)
				return
			}

			requestIdentity := Identity{
				UserID: claims.UserID,
				AppID:  appID,
			}

			ctx := WithContext(
				r.Context(),
				requestIdentity,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"error": message,
		},
	)
}
