package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"
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

			err = c.solService.Sync(ctx, rep, months, publishers)
			if err != nil {
				return err
			}

			return rep.reportResults()
		},
	}
}

func getPublishersUserInput(cmd *cli.Command) ([]string, error) {
	var input []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select your publishers").
				Options(
					huh.NewOption("DC", "dc"),
					huh.NewOption("Marvel", "marvel"),
					huh.NewOption("Image", "image"),
				).
				Value(&input),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(input) == 0 {
		return nil, errors.New("no publishers to scrape provided")
	}
	return input, nil
}

func getMonthsUserInput(cmd *cli.Command) ([]string, error) {
	months := []time.Month{time.January, time.February, time.March, time.April, time.May, time.June, time.July,
		time.August, time.September, time.October, time.November, time.December}

	monthOptions := slices.Collect(func(yield func(huh.Option[string]) bool) {
		for _, month := range months {
			s := month.String()

			if !yield(huh.NewOption(s, strings.ToLower(s))) {
				return
			}
		}
	})

	var input []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select your months").
				Options(monthOptions...).
				Value(&input),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	if len(input) == 0 {
		return nil, errors.New("no months to scrape provided")
	}
	return input, nil
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
