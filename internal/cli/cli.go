package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
	"log/slog"
	"slices"
	"strings"
)

var (
	allowedPublishers = []string{"dc", "marvel", "image"}
	allowedMonths     = []string{"january", "february", "march", "april", "may", "june", "july", "august", "september",
		"october", "november", "december"}
)

type CLI struct {
	cmd        *cli.Command
	solService *service.SolicitationService

	form    *huh.Form
	metrics *models.AppMetrics
	logger  *slog.Logger
}

func New(s *service.SolicitationService, m *models.AppMetrics, l *slog.Logger) *CLI {
	c := &CLI{
		solService: s,
		metrics:    m,
		logger:     l,
	}

	c.cmd = &cli.Command{
		Name:  "solipull",
		Usage: "Solipull tool",
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Println("Hello, world!")
			return nil
		},
		Commands: []*cli.Command{
			c.solicitation(),
		},
	}

	return c
}

func (c *CLI) Run(ctx context.Context, args []string) error {
	return c.cmd.Run(ctx, args)
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
