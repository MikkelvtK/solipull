package main

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/app"
	"github.com/MikkelvtK/solipull/internal/cli"
	"os"
)

func main() {
	a := app.NewApplication([]string{"march"}, []string{"dc"})

	cmd := cli.New(a.Serv)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
