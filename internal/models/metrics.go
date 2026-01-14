package models

import "log/slog"

type AppMetrics struct {
}

type ErrorObserver interface {
	OnError(level slog.Level, args ...string)
}
