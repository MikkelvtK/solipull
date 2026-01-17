package cli

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/urfave/cli/v3"
	"log/slog"
)

type CLI struct {
	cmd        *cli.Command
	solService *service.SolicitationService

	syncRep *syncReporter
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
			c.solicitations(),
		},
	}

	return c
}

func (c *CLI) Run(ctx context.Context, args []string) error {
	return c.cmd.Run(ctx, args)
}

func (c *CLI) solicitations() *cli.Command {
	return &cli.Command{
		Name:  "solicitations",
		Usage: "Solicitation tool",
		Commands: []*cli.Command{
			c.sync(),
		},
	}
}
