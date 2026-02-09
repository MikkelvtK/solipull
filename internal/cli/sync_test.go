package cli

import (
	"bytes"
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
)

func captureStdout(f func(), t *testing.T) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	if err := w.Close(); err != nil {
		t.Errorf("Error closing stdout: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Error reading stdout: %v", err)
	}

	os.Stdout = oldStdout
	return buf.String()
}

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
		{
			name: "accepts nil parameters",
			args: args{
				metrics: nil,
				logger:  nil,
			},
			want: &syncReporter{},
		},
		{
			name: "accepts AppMetrics",
			args: args{
				metrics: &models.AppMetrics{},
				logger:  nil,
			},
			want: &syncReporter{metrics: &models.AppMetrics{}},
		},
		{
			name: "accepts Logger",
			args: args{
				metrics: nil,
				logger:  &slog.Logger{},
			},
			want: &syncReporter{logger: &slog.Logger{}},
		},
		{
			name: "accepts both",
			args: args{
				metrics: &models.AppMetrics{},
				logger:  &slog.Logger{},
			},
			want: &syncReporter{metrics: &models.AppMetrics{}, logger: &slog.Logger{}},
		},
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
		metrics *models.AppMetrics
	}
	type args struct {
		n int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		numCalls int
		want     int
	}{
		{
			name: "default test case",
			fields: fields{
				metrics: &models.AppMetrics{},
			},
			args: args{
				n: 1,
			},
			numCalls: 1,
			want:     1,
		},
		{
			name: "persists calls",
			fields: fields{
				metrics: &models.AppMetrics{},
			},
			args: args{
				n: 1,
			},
			numCalls: 5,
			want:     5,
		},
		{
			name: "accepts bigger numbers",
			fields: fields{
				metrics: &models.AppMetrics{},
			},
			args: args{
				n: 101,
			},
			numCalls: 5,
			want:     505,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				metrics: tt.fields.metrics,
			}
			for i := 0; i < tt.numCalls; i++ {
				s.OnComicBookScraped(tt.args.n)
			}
			if int(s.metrics.ComicBooksFound.Load()) != tt.want {
				t.Errorf("OnComicBookScraped got = %v, want = %v", s.metrics.ComicBooksFound.Load(), tt.want)
			}
		})
	}
}

func Test_syncReporter_OnError(t *testing.T) {
	type fields struct {
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
		{
			name: "default test case",
			fields: fields{
				metrics: &models.AppMetrics{},
				logger:  nil,
			},
			args: args{
				ctx:   context.Background(),
				level: slog.LevelInfo,
				msg:   "test",
				args:  []any{"test"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				metrics: tt.fields.metrics,
				logger:  tt.fields.logger,
			}
			s.OnError(tt.args.ctx, tt.args.level, tt.args.msg, tt.args.args...)
		})
	}
}

func Test_syncReporter_OnNavigationComplete(t *testing.T) {
	type fields struct {
		metrics *models.AppMetrics
	}
	tests := []struct {
		name   string
		fields fields
		want   string
		wantPb bool
	}{
		{
			name: "no pages found",
			fields: fields{
				metrics: &models.AppMetrics{},
			},
			want:   "No pages found",
			wantPb: false,
		},
		{
			name: "pages found",
			fields: fields{
				metrics: func() *models.AppMetrics {
					m := &models.AppMetrics{}
					m.PagesFound.Add(5)
					return m
				}(),
			},
			want:   "✔ Found: 5 pages to scrape",
			wantPb: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				metrics: tt.fields.metrics,
			}
			output := captureStdout(s.OnNavigationComplete, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("OnNavigationComplete got = %v, want = %v", output, tt.want)
				return
			}
			if tt.wantPb && s.pb == nil {
				t.Errorf("OnNavigationComplete wantPb = %v got %v", tt.wantPb, s.pb)
			}
		})
	}
}

func Test_syncReporter_OnScrapingComplete(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "default test case",
			fields: fields{
				pb: progressbar.NewOptions(5, progressbar.OptionSetVisibility(false)),
			},
		},
		{
			name: "pb is nil",
			fields: fields{
				pb:      nil,
				metrics: &models.AppMetrics{},
			},
		},
		{
			name: "error pb",
			fields: fields{
				pb: progressbar.NewOptions(0,
					progressbar.OptionSetVisibility(false)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
			}
			s.OnScrapingComplete()
		})
	}
}

func Test_syncReporter_OnUrlFound(t *testing.T) {
	type fields struct {
		metrics *models.AppMetrics
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "default test case",
			fields: fields{
				metrics: &models.AppMetrics{},
			},
			args: args{
				n: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				metrics: tt.fields.metrics,
			}
			s.OnUrlFound(tt.args.n)
		})
	}
}

func Test_syncReporter_reportResults(t *testing.T) {
	type fields struct {
		pb      *progressbar.ProgressBar
		metrics *models.AppMetrics
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "default test case",
			fields: fields{
				pb:      progressbar.NewOptions(5, progressbar.OptionSetVisibility(false)),
				metrics: &models.AppMetrics{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncReporter{
				pb:      tt.fields.pb,
				metrics: tt.fields.metrics,
			}
			if err := s.reportResults(); (err != nil) != tt.wantErr {
				t.Errorf("reportResults() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
