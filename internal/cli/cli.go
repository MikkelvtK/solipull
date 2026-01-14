package cli

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/urfave/cli/v3"
)

type CLI struct {
	cmd        *cli.Command
	solService *service.SolicitationService
}

func New(s *service.SolicitationService) *CLI {
	c := &CLI{
		solService: s,
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

func (c *CLI) sync() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Synchronize local database with the latest comic book publisher solicitations.",
		Description: "Scrapes Comic Releases sitemap and solicitation pages to identify new comic book releases. " +
			"Discovered titles are parsed for data and inserted into the local SQLite database. This process ensures " +
			"your available titles are up to date for collection and pull-list management.",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return c.solService.Sync(ctx)
		},
	}
}
