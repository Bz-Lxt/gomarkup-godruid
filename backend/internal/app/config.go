package app

import (
	"os"
	"strconv"
	"strings"
	"time"

	"godruid/internal/pool"
)

type Settings struct {
	Listen    string
	Demo      bool
	Connector string
	LogLevel  string
	Target    string
	Pool      pool.Config
}

func LoadSettings() Settings {
	cfg := pool.DefaultConfig()
	s := Settings{
		Listen:    env("GODRUID_LISTEN", ":8080"),
		Demo:      envBool("GODRUID_DEMO", true),
		Connector: strings.ToLower(env("GODRUID_CONNECTOR", "fake")),
		LogLevel:  env("GODRUID_LOG_LEVEL", "info"),
		Target:    env("GODRUID_TARGET", "127.0.0.1:6379"),
		Pool:      cfg,
	}
	if v, ok := envInt("GODRUID_MAX_IDLE"); ok {
		s.Pool.MaxIdle = v
	}
	if v, ok := envInt("GODRUID_MAX_ACTIVE"); ok {
		s.Pool.MaxActive = v
	}
	if v, ok := envDur("GODRUID_MAX_WAIT"); ok {
		s.Pool.MaxWaitTimeout = v
	}
	if v, ok := envDur("GODRUID_IDLE_TTL"); ok {
		s.Pool.IdleTTL = v
	}
	if v, ok := envDur("GODRUID_HEALTH_INTERVAL"); ok {
		s.Pool.HealthInterval = v
	}
	return s
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envInt(key string) (int, bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

func envDur(key string) (time.Duration, bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0, false
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, false
	}
	return d, true
}
