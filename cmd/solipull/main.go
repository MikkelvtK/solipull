package main

import (
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"log"
)

func main() {
	c := cache.NewCache()

	s, err := scraper.NewComicReleasesScraper(c, []string{"march"}, []string{"dc"})
	if err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}
