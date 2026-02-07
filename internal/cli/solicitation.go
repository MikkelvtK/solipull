package cli

import (
	"github.com/urfave/cli/v3"
)

func (c *CLI) solicitation() *cli.Command {
	return &cli.Command{
		Name:  "solicitation",
		Usage: "Solicitation tool",
		Commands: []*cli.Command{
			c.sync(),
			c.view(),
		},
	}
}
