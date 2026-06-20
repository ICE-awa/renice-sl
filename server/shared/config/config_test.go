package config

import "testing"

func TestDatabaseConfig_DSN(t *testing.T) {
	t.Parallel()

	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "ice",
		Password: "secret",
		DBName:   "renice",
		SSLMode:  "disable",
	}

	got := cfg.DSN()
	want := "postgres://ice:secret@localhost:5432/renice?sslmode=disable"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRedisConfig_ADDR(t *testing.T) {
	t.Parallel()

	cfg := &RedisConfig{Host: "redis", Port: 6379}

	got := cfg.ADDR()
	want := "redis:6379"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
