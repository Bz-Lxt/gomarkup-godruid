package pool

import (
	"fmt"
	"time"

	"godruid/internal/errorsx"
)

// Config is validated at pool construction. Zero values are not used as defaults
// except through DefaultConfig / Options.
type Config struct {
	MaxIdle            int
	MaxActive          int
	MaxWaitTimeout     time.Duration
	IdleTTL            time.Duration
	HealthInterval     time.Duration
	HealthTimeout      time.Duration
	DialTimeout        time.Duration
	ReconnectAttempts  int
	ReconnectBaseDelay time.Duration
	ReconnectMaxDelay  time.Duration
	SnapshotInterval   time.Duration
	MetricsRetention   time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxIdle:            8,
		MaxActive:          32,
		MaxWaitTimeout:     3 * time.Second,
		IdleTTL:            45 * time.Second,
		HealthInterval:     5 * time.Second,
		HealthTimeout:      time.Second,
		DialTimeout:        3 * time.Second,
		ReconnectAttempts:  5,
		ReconnectBaseDelay: 100 * time.Millisecond,
		ReconnectMaxDelay:  2 * time.Second,
		SnapshotInterval:   time.Second,
		MetricsRetention:   15 * time.Minute,
	}
}

func (c Config) Validate() error {
	if c.MaxActive <= 0 {
		return fieldErr("MaxActive", "must be > 0")
	}
	if c.MaxIdle < 0 {
		return fieldErr("MaxIdle", "must be >= 0")
	}
	if c.MaxIdle > c.MaxActive {
		return fieldErr("MaxIdle", "must be <= MaxActive")
	}
	for _, item := range []struct {
		name string
		d    time.Duration
	}{
		{"MaxWaitTimeout", c.MaxWaitTimeout},
		{"IdleTTL", c.IdleTTL},
		{"HealthInterval", c.HealthInterval},
		{"HealthTimeout", c.HealthTimeout},
		{"DialTimeout", c.DialTimeout},
		{"ReconnectBaseDelay", c.ReconnectBaseDelay},
		{"ReconnectMaxDelay", c.ReconnectMaxDelay},
		{"SnapshotInterval", c.SnapshotInterval},
		{"MetricsRetention", c.MetricsRetention},
	} {
		if item.d <= 0 {
			return fieldErr(item.name, "must be > 0")
		}
	}
	if c.ReconnectAttempts < 1 {
		return fieldErr("ReconnectAttempts", "must be >= 1")
	}
	if c.ReconnectMaxDelay < c.ReconnectBaseDelay {
		return fieldErr("ReconnectMaxDelay", "must be >= ReconnectBaseDelay")
	}
	return nil
}

func fieldErr(field, msg string) error {
	return errorsx.Invalid(field, fmt.Sprintf("%s %s", field, msg))
}
