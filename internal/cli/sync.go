package cli

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
)

func (c *CLI) sync() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Synchronize local database with the latest comic book publisher solicitations.",
		Description: "Scrapes Comic Releases sitemap and solicitation pages to identify new comic book releases. " +
			"Discovered titles are parsed for data and inserted into the local SQLite database. This process ensures " +
			"your available titles are up to date for collection and pull-list management.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			publishers, err := getPublishersUserInput(cmd)
			if err != nil {
				return err
			}

			months, err := getMonthsUserInput(cmd)
			if err != nil {
				return err
			}

			rep := newSyncReporter(c.metrics, c.logger)
			if err = c.solService.Sync(ctx, rep, months, publishers); err != nil {
				return err
			}

			return rep.reportResults()
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "publisher",
				Aliases: []string{"p"},
				Usage:   "Publishers to sync",
			},
			&cli.StringSliceFlag{
				Name:    "month",
				Aliases: []string{"m"},
				Usage:   "Months to sync",
			},
		},
	}
}

type syncReporter struct {
	pb *progressbar.ProgressBar

	metrics *models.AppMetrics
	logger  *slog.Logger
}

func (s *syncReporter) OnError(ctx context.Context, level slog.Level, msg string, args ...any) {
	s.metrics.ErrorsFound.Add(1)
	//s.logger.Log(ctx, level, msg, args)
}

func (s *syncReporter) OnUrlFound(n int) {
	s.metrics.PagesFound.Add(int32(n))
}

func (s *syncReporter) OnNavigationComplete() {
	if s.metrics.PagesFound.Load() == 0 {
		s.OnError(nil, slog.LevelError, "no pages found")
		return
	}

	fmt.Printf("✔ Found: %d pages to scrape\n\n", s.metrics.PagesFound.Load())

	s.pb = progressbar.NewOptions(int(s.metrics.PagesFound.Load()),
		progressbar.OptionSetDescription("➔ Pages scraped:"),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionShowCount(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    "=",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func (s *syncReporter) reportResults() error {
	if s.pb != nil {
		if err := s.pb.Finish(); err != nil {
			return err
		}
	}

	fmt.Println("✅  Sync complete!")
	fmt.Printf("   Scraped: %d new comics\n\n", s.metrics.ComicBooksFound.Load())

	if s.metrics.ComicBooksFound.Load() > 0 {
		fmt.Printf("⚠️ Finished with %d extraction warnings.\n   "+
			"Run 'solipull logs' to view detailed diagnostics.\n", s.metrics.ErrorsFound.Load())
	}
	return nil
}

func (s *syncReporter) OnComicBookScraped(n int) {
	s.metrics.ComicBooksFound.Add(int32(n))
}

func (s *syncReporter) OnScrapingComplete() {
	if s.pb == nil {
		s.OnError(nil, slog.LevelError, "nothing to scrape")
		return
	}

	if err := s.pb.Add(1); err != nil {
		fmt.Println("Error on scraping:", err)
	}
}

func newSyncReporter(metrics *models.AppMetrics, logger *slog.Logger) *syncReporter {
	fmt.Println("➔ Finding solicitation pages to scrape...")

	return &syncReporter{
		metrics: metrics,
		logger:  logger,
	}
}
