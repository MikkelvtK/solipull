package main

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/app"
)

func main() {
	a := app.NewApplication([]string{"february", "march"}, []string{"dc", "marvel"})
	a.Serv.Sync(context.Background())
}
