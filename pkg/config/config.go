package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Auth     AuthConfig
	Email    EmailConfig
}

type ServerConfig struct {
	Host string
	Port int
	// Mode is "local" (single-inspector, SQLite) or "cloud" (multi-tenant, PostgreSQL).
	Mode string
}

type DatabaseConfig struct {
	// Driver is "sqlite" or "postgres".
	Driver string
	DSN    string
}

type StorageConfig struct {
	// Driver is "local" or "s3".
	Driver    string
	LocalPath string
	S3Bucket  string
	S3Region  string
}

type AuthConfig struct {
	JWTSecret     string
	TokenTTLHours int
}

type EmailConfig struct {
	// Driver is "smtp", "sendgrid", or "queue_only".
	// queue_only persists deliveries for later sending — suitable for offline mode.
	Driver   string
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: env("SERVER_HOST", "0.0.0.0"),
			Port: envInt("SERVER_PORT", 8080),
			Mode: env("SERVER_MODE", "local"),
		},
		Database: DatabaseConfig{
			Driver: env("DATABASE_DRIVER", "sqlite"),
			DSN:    env("DATABASE_DSN", "./juno.db"),
		},
		Storage: StorageConfig{
			Driver:    env("STORAGE_DRIVER", "local"),
			LocalPath: env("STORAGE_LOCAL_PATH", "./data/photos"),
			S3Bucket:  env("STORAGE_S3_BUCKET", ""),
			S3Region:  env("STORAGE_S3_REGION", ""),
		},
		Auth: AuthConfig{
			JWTSecret:     env("JWT_SECRET", ""),
			TokenTTLHours: envInt("JWT_TTL_HOURS", 24),
		},
		Email: EmailConfig{
			Driver:   env("EMAIL_DRIVER", "queue_only"),
			SMTPHost: env("SMTP_HOST", ""),
			SMTPPort: envInt("SMTP_PORT", 587),
			SMTPUser: env("SMTP_USER", ""),
			SMTPPass: env("SMTP_PASS", ""),
		},
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	if c.Server.Mode != "local" && c.Server.Mode != "cloud" {
		return fmt.Errorf("SERVER_MODE must be 'local' or 'cloud', got %q", c.Server.Mode)
	}
	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("DATABASE_DRIVER must be 'sqlite' or 'postgres', got %q", c.Database.Driver)
	}
	if c.Auth.JWTSecret == "" {
		if c.Server.Mode == "cloud" {
			return fmt.Errorf("JWT_SECRET is required in cloud mode")
		}
		// Permissive default for local dev; not safe for production.
		c.Auth.JWTSecret = "dev-secret-change-me"
	}
	return nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
