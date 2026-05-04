package util

import (
	"errors"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(cfg *config.JwtConfig, userID int64, username string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessExpires)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "access",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GenerateRefreshToken(cfg *config.JwtConfig, userID int64, username string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshExpires)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "refresh",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(cfg.RefreshSecret))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func ParseAccessToken(cfg *config.JwtConfig, token string) (*Claims, error) {
	return parseToken(cfg.AccessSecret, token, "access")
}

func ParseRefreshToken(cfg *config.JwtConfig, token string) (*Claims, error) {
	return parseToken(cfg.RefreshSecret, token, "refresh")
}

func parseToken(secret, tokenStr, expectedSubject string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	sub, _ := claims.GetSubject()
	if sub != expectedSubject {
		return nil, errors.New("token type mismatch")
	}

	return claims, nil
}
