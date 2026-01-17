package models

import (
	"context"
	"log/slog"
	"sync/atomic"
)

type AppMetrics struct {
	ErrorsFound     atomic.Int32
	ComicBooksFound atomic.Int32
	PagesFound      atomic.Int32
}

type ErrorObserver interface {
	OnError(ctx context.Context, level slog.Level, msg string, args ...any)
}
