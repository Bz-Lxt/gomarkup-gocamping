package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu sync.RWMutex
	L  *slog.Logger
)

func Init(level string, env string) {
	lv := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		if env != "production" {
			lv = slog.LevelDebug
		}
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: lv}
	var w io.Writer = os.Stdout
	h := slog.NewJSONHandler(w, opts)
	mu.Lock()
	L = slog.New(h)
	mu.Unlock()
}

func Get() *slog.Logger {
	mu.RLock()
	l := L
	mu.RUnlock()
	if l == nil {
		Init("info", "development")
		return Get()
	}
	return l
}

func Info(msg string, args ...any) {
	if l := Get(); l != nil {
		l.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if l := Get(); l != nil {
		l.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if l := Get(); l != nil {
		l.Error(msg, args...)
	}
}

func Debug(msg string, args ...any) {
	if l := Get(); l != nil {
		l.Debug(msg, args...)
	}
}
