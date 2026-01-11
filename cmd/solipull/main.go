package main

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/app"
	"log"
)

func main() {
	a := app.NewApplication([]string{"march"}, []string{"dc"})
	err := a.Scraper.Run()
	if err != nil {
		log.Fatal(err)
	}

	cbs, err := a.Cache.GetAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cbs)
}
