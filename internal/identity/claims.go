package identity

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID string `json:"user_id"`

	ClientIP string `json:"client_ip,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Version  int64  `json:"version,omitempty"`

	jwt.RegisteredClaims
}
