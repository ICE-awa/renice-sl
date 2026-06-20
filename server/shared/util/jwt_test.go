package util

import (
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/shared/config"
)

func jwtTestConfig() *config.JwtConfig {
	return &config.JwtConfig{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		AccessExpires:  time.Hour,
		RefreshExpires: 24 * time.Hour,
	}
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	t.Parallel()

	cfg := jwtTestConfig()
	token, err := GenerateAccessToken(cfg, 42, "ice")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := ParseAccessToken(cfg, token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "ice" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestGenerateAndParseRefreshToken(t *testing.T) {
	t.Parallel()

	cfg := jwtTestConfig()
	token, err := GenerateRefreshToken(cfg, 42, "ice")
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	claims, err := ParseRefreshToken(cfg, token)
	if err != nil {
		t.Fatalf("ParseRefreshToken returned error: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "ice" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestParseTokenRejectsWrongTokenType(t *testing.T) {
	t.Parallel()

	cfg := jwtTestConfig()
	refreshToken, err := GenerateRefreshToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAccessToken(cfg, refreshToken); err == nil {
		t.Fatal("refresh token should not parse as access token")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	t.Parallel()

	cfg := jwtTestConfig()
	token, err := GenerateAccessToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}

	other := jwtTestConfig()
	other.AccessSecret = "other-secret"
	if _, err := ParseAccessToken(other, token); err == nil {
		t.Fatal("token signed with another secret should be rejected")
	}
}
