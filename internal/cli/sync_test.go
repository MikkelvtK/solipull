package cli

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
	"log/slog"
	"reflect"
	"testing"
)

func TestCLI_sync(t *testing.T) {
	tests := []struct {
		name string
		want *cli.Command
	}{
		{
			name: "returns a cli.command",
			want: &cli.Command{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLI{}
			if got := c.sync(); got == nil {
				t.Errorf("sync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newSyncReporter(t *testing.T) {
	type args struct {
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *syncReporter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSyncReporter(tt.args.metrics, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSyncReporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncReporter_OnComicBookScraped(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnComicBookScraped(tt.args.n)
		})
	}
}

func Test_syncReporter_OnError(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	type args struct {
		ctx   context.Context
		level slog.Level
		msg   string
		args  []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnError(tt.args.ctx, tt.args.level, tt.args.msg, tt.args.args...)
		})
	}
}

func Test_syncReporter_OnNavigationComplete(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnNavigationComplete()
		})
	}
}

func Test_syncReporter_OnScrapingComplete(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnScrapingComplete()
		})
	}
}

func Test_syncReporter_OnUrlFound(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnUrlFound(tt.args.n)
		})
	}
}

func Test_syncReporter_reportResults(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
		logger  *slog.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			if err := s.reportResults(); (err != nil) != tt.wantErr {
				t.Errorf("reportResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
