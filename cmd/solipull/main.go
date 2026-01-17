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
	a := app.NewApplication([]string{"march"}, []string{"dc", "marvel"})

	cmd := cli.New(a.Serv, &models.AppMetrics{}, slog.Default())
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
