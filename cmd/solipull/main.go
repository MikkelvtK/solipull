package main

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/app"
)

func main() {
	a := app.NewApplication([]string{"march"}, []string{"dc"})
	a.Serv.Sync(context.Background())
}
