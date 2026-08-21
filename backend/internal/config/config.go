package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr        string
	Env             string
	LogLevel        string
	DatabaseDSN     string
	JWTSecret       string
	JWTExpireHours  int
	UploadDir       string
	UploadMaxBytes  int64
	BlevePath       string
	SearchAnalyzer  string
	LockHeartbeatS  int
	LockTimeoutS    int
}

func Load() Config {
	return Config{
		HTTPAddr:       env("GOWIKI_HTTP_ADDR", ":8080"),
		Env:            env("GOWIKI_ENV", "development"),
		LogLevel:       env("GOWIKI_LOG_LEVEL", "info"),
		DatabaseDSN:    env("GOWIKI_DATABASE_DSN", "postgres://gowiki:gowiki@127.0.0.1:27123/gowiki?sslmode=disable"),
		JWTSecret:      env("GOWIKI_JWT_SECRET", "gowiki-dev-secret-change-me"),
		JWTExpireHours: envInt("GOWIKI_JWT_EXPIRE_HOURS", 24),
		UploadDir:      env("GOWIKI_UPLOAD_DIR", "./data/uploads"),
		UploadMaxBytes: int64(envInt("GOWIKI_UPLOAD_MAX_MB", 8)) * 1024 * 1024,
		BlevePath:      env("GOWIKI_BLEVE_PATH", "./data/bleve"),
		SearchAnalyzer: strings.ToLower(env("GOWIKI_SEARCH_ANALYZER", "gse")),
		LockHeartbeatS: envInt("GOWIKI_LOCK_HEARTBEAT_S", 60),
		LockTimeoutS:   envInt("GOWIKI_LOCK_TIMEOUT_S", 180),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
