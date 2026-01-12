package models

import "sync/atomic"

type RunStats struct {
	ErrorCount *atomic.Int32
}

func NewRunStats() *RunStats {
	return &RunStats{ErrorCount: &atomic.Int32{}}
}
