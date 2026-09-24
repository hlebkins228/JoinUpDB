package jwt

import (
	"JoinUp/internal/settings"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int
	Role   string
	jwt.RegisteredClaims
}

type JwtManager struct {
	secret string
}

func NewJwtManager(secret string) JwtManager {
	return JwtManager{secret: secret}
}

func (m *JwtManager) NewToken(userID int, role string) (string, error) {
	claims := Claims{
		Role:             role,
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(settings.JwtTokenExpDuration))}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.secret))
}

func (m *JwtManager) ValidateToken(tokenStr string) (*Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims,
		func(token *jwt.Token) (any, error) {
			// Защита от подмены alg
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secret), nil
		})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &claims, nil
}

func (m *JwtManager) ValidateTokenFromHeader(header http.Header) (*Claims, error) {
	s := header.Get(settings.AuthHeader)
	if s == "" {
		return nil, fmt.Errorf("no authorization header")
	}
	tokenStr := strings.TrimPrefix(s, "Bearer ")
	if s == tokenStr {
		return nil, fmt.Errorf("invalid authorization header")
	}

	return m.ValidateToken(tokenStr)
}
