package main

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/app"
)

func main() {
	a := app.NewApplication([]string{"february"}, []string{"marvel"})
	a.Serv.Sync(context.Background())
}
