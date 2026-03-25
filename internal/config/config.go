package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	JWT     JWTConfig
	SMTP    SMTPConfig
	Storage StorageConfig
}

type AppConfig struct {
	Env         string
	Port        string
	URL         string
	FrontendURL string
}

type DBConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type StorageConfig struct {
	Type      string
	LocalPath string
	S3Bucket  string
	S3Region  string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !strings.Contains(err.Error(), "no such file or directory") {
				return nil, fmt.Errorf("config read .env: %w", err)
			}
		}
	}

	setDefaults()

	cfg := &Config{}

	cfg.App = AppConfig{
		Env:         viper.GetString("APP_ENV"),
		Port:        viper.GetString("APP_PORT"),
		URL:         viper.GetString("APP_URL"),
		FrontendURL: viper.GetString("FRONTEND_URL"),
	}

	cfg.DB = DBConfig{
		URL:             viper.GetString("DATABASE_URL"),
		MaxConns:        viper.GetInt32("DB_MAX_CONNS"),
		MinConns:        viper.GetInt32("DB_MIN_CONNS"),
		MaxConnLifetime: viper.GetDuration("DB_MAX_CONN_LIFETIME"),
		MaxConnIdleTime: viper.GetDuration("DB_MAX_CONN_IDLE_TIME"),
	}

	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	cfg.JWT = JWTConfig{
		Secret:        viper.GetString("JWT_SECRET"),
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}

	cfg.SMTP = SMTPConfig{
		Host: viper.GetString("SMTP_HOST"),
		Port: viper.GetInt("SMTP_PORT"),
		User: viper.GetString("SMTP_USER"),
		Pass: viper.GetString("SMTP_PASS"),
		From: viper.GetString("SMTP_FROM"),
	}

	cfg.Storage = StorageConfig{
		Type:      viper.GetString("STORAGE_TYPE"),
		LocalPath: viper.GetString("STORAGE_LOCAL_PATH"),
		S3Bucket:  viper.GetString("STORAGE_S3_BUCKET"),
		S3Region:  viper.GetString("STORAGE_S3_REGION"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_URL", "http://localhost:8080")
	viper.SetDefault("FRONTEND_URL", "http://localhost:5173")

	viper.SetDefault("DB_MAX_CONNS", 25)
	viper.SetDefault("DB_MIN_CONNS", 5)
	viper.SetDefault("DB_MAX_CONN_LIFETIME", "1h")
	viper.SetDefault("DB_MAX_CONN_IDLE_TIME", "30m")

	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "720h")

	viper.SetDefault("SMTP_PORT", 587)

	viper.SetDefault("STORAGE_TYPE", "local")
	viper.SetDefault("STORAGE_LOCAL_PATH", "./storage/pdfs")
}

func (c *Config) validate() error {
	if c.DB.URL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("config: JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("config: JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
