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
)

var (
	allowedPublishers = []string{"dc", "marvel", "image"}
	allowedMonths     = []string{"january", "february", "march", "april", "may", "june", "july", "august", "september",
		"october", "november", "december"}
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

func getPublishersUserInput(cmd *cli.Command) ([]string, error) {
	raw := cmd.StringSlice("publisher")

	if len(raw) > 0 {
		publishers, err := parseStringSliceFlag("publisher", raw, allowedPublishers)
		if err != nil {
			return nil, err
		}

		return publishers, nil
	}

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
				Value(&input).
				Validate(func(v []string) error {
					if len(v) == 0 {
						return errors.New("please select at least one publisher")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}
	return input, nil
}

func getMonthsUserInput(cmd *cli.Command) ([]string, error) {
	raw := cmd.StringSlice("month")

	if len(raw) > 0 {
		months, err := parseStringSliceFlag("month", raw, allowedMonths)
		if err != nil {
			return nil, err
		}

		return months, nil
	}

	monthOptions := slices.Collect(func(yield func(huh.Option[string]) bool) {
		for _, month := range allowedMonths {
			if !yield(huh.NewOption(month, strings.ToLower(month))) {
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

func parseStringSliceFlag(flagName string, input, allowedValues []string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("invalid %s input", flagName)
	}

	result := make([]string, 0, len(input))

	for _, item := range input {
		parts := strings.Split(item, ",")

		for _, p := range parts {
			lp := strings.ToLower(strings.TrimSpace(p))

			if !slices.Contains(allowedValues, lp) {
				return nil, fmt.Errorf("invalid %s specified: %s", flagName, p)
			}

			if slices.Contains(result, lp) {
				return nil, fmt.Errorf("duplicate %s specified: %s", flagName, p)
			}

			result = append(result, lp)
		}
	}

	return result, nil
}
