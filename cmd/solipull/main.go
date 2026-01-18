package main

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/app"
	"github.com/MikkelvtK/solipull/internal/cli"
	"github.com/MikkelvtK/solipull/internal/models"
	"log/slog"
	"os"
)

func main() {
	a := app.NewApplication([]string{"march", "february"}, []string{"dc", "marvel", "image"})

	cmd := cli.New(a.Serv, &models.AppMetrics{}, slog.Default())
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error running cli: %v\n", err.Error())
		os.Exit(1)
	}
}
