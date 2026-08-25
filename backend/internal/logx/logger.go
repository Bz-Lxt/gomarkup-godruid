package logx

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu      sync.RWMutex
	current *slog.Logger
	level   = new(slog.LevelVar)
)

func init() {
	level.Set(slog.LevelInfo)
	current = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func Init(lvl string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}
	l := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
	mu.Lock()
	current = l
	mu.Unlock()
	return l
}

func L() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		return slog.Default()
	}
	return current
}

func Level() slog.Level {
	return level.Level()
}
