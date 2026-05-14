package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Nats     NatsConfig     `mapstructure:"nats"`
	Jwt      JwtConfig      `mapstructure:"jwt"`
	Link     LinkConfig     `mapstructure:"link"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type NatsConfig struct {
	Host string `mapstructure:"host"`
}

type JwtConfig struct {
	AccessSecret   string        `mapstructure:"access_token"`
	RefreshSecret  string        `mapstructure:"refresh_token"`
	AccessExpires  time.Duration `mapstructure:"access_expires"`
	RefreshExpires time.Duration `mapstructure:"refresh_expires"`
}

type LinkConfig struct {
	Expires     time.Duration `mapstructure:"expires"`
	NullExpires time.Duration `mapstructure:"null_expires"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

func (d *RedisConfig) ADDR() string {
	return fmt.Sprintf(
		"%s:%d",
		d.Host, d.Port,
	)
}

func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("server.mode", "SERVER_MODE")
	_ = v.BindEnv("database.host", "DATABASE_HOST")
	_ = v.BindEnv("database.port", "DATABASE_PORT")
	_ = v.BindEnv("database.user", "DATABASE_USER")
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = v.BindEnv("database.dbname", "DATABASE_DBNAME")
	_ = v.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	_ = v.BindEnv("nats.host", "NATS_HOST")
	_ = v.BindEnv("jwt.access_secret", "JWT_ACCESS_SECRET")
	_ = v.BindEnv("jwt.refresh_secret", "JWT_REFRESH_SECRET")
	_ = v.BindEnv("jwt.access_expires", "JWT_ACCESS_EXPIRES")
	_ = v.BindEnv("jwt.refresh_expires", "JWT_REFRESH_EXPIRES")
	_ = v.BindEnv("link.expires", "LINK_EXPIRES")
	_ = v.BindEnv("link.null_expires", "LINK_NULL_EXPIRES")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}
