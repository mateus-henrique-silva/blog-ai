package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	AppPort         string
	AppSecret       string // 32-byte hex, used for AES-256-GCM session encryption
	IPHashSecret    string // used for SHA-256 IP hashing in analytics
	DBPath          string
	UploadDir       string
	UploadMaxMB     int64
	SessionDuration time.Duration
	RateLimitLogin  int           // max login attempts per window
	RateLimitWindow time.Duration // rolling window duration
	CSPMode         string        // "strict" or "lenient"
}

func Load() *Config {
	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnv("APP_PORT", "3000"),
		AppSecret:       getEnv("APP_SECRET", ""),
		IPHashSecret:    getEnv("IP_HASH_SECRET", ""),
		DBPath:          getEnv("DB_PATH", "./data/blog.db"),
		UploadDir:       getEnv("UPLOAD_DIR", "./web/static/uploads"),
		UploadMaxMB:     int64(getEnvInt("UPLOAD_MAX_MB", 20)),
		SessionDuration: getEnvDuration("SESSION_DURATION", 24*time.Hour),
		RateLimitLogin:  getEnvInt("RATE_LIMIT_LOGIN", 5),
		RateLimitWindow: getEnvDuration("RATE_LIMIT_WINDOW", 15*time.Minute),
		CSPMode:         getEnv("CSP_MODE", "lenient"),
	}

	if cfg.AppEnv == "production" {
		if cfg.AppSecret == "" {
			log.Fatal("APP_SECRET must be set in production")
		}
		if cfg.IPHashSecret == "" {
			log.Fatal("IP_HASH_SECRET must be set in production")
		}
	} else {
		if cfg.AppSecret == "" {
			log.Println("WARNING: APP_SECRET not set — using insecure development default")
			cfg.AppSecret = "dev-secret-key-32-bytes-xxxxxxxx"
		}
		if cfg.IPHashSecret == "" {
			log.Println("WARNING: IP_HASH_SECRET not set — using insecure development default")
			cfg.IPHashSecret = "dev-ip-hash-secret-placeholder"
		}
	}

	return cfg
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
