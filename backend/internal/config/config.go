package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env            string
	HTTPAddr       string
	DatabaseURL    string
	RedisAddr      string
	JWTSecret      string
	TileProvider   string
	DEMProvider    string
	DEMHTTPURL     string
	GPSProvider    string
	NotifyProvider string
	NotifyHTTPURL  string
	CORSOrigins    []string
	LogLevel       string
	JWTExpire      time.Duration
	MapboxToken    string
}

func Load() Config {
	c := Config{
		Env:            getenv("APP_ENV", "development"),
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:    getenv("DATABASE_URL", "postgres://gocamping:gocamping@127.0.0.1:28314/gocamping?sslmode=disable"),
		RedisAddr:      getenv("REDIS_ADDR", "127.0.0.1:28315"),
		JWTSecret:      getenv("JWT_SECRET", "gocamping-dev-secret-change-me"),
		TileProvider:   getenv("TILE_PROVIDER", "local"),
		DEMProvider:    getenv("DEM_PROVIDER", "synthetic"),
		DEMHTTPURL:     getenv("DEM_HTTP_URL", "https://api.open-elevation.com/api/v1/lookup"),
		GPSProvider:    getenv("GPS_PROVIDER", "simulator"),
		NotifyProvider: getenv("NOTIFY_PROVIDER", "mock"),
		NotifyHTTPURL:  getenv("NOTIFY_HTTP_URL", ""),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		JWTExpire:      time.Duration(getenvInt("JWT_EXPIRE_HOURS", 72)) * time.Hour,
		MapboxToken:    getenv("MAPBOX_TOKEN", ""),
	}
	raw := getenv("CORS_ORIGINS", "http://localhost:28311,http://localhost:28312")
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			c.CORSOrigins = append(c.CORSOrigins, p)
		}
	}
	return c
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
