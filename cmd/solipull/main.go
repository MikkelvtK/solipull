package main

import (
	"github.com/MikkelvtK/solipull/internal/scraper"
	"log"
)

func main() {
	s, err := scraper.NewComicReleasesScraper([]string{"february", "march"}, []string{"dc", "marvel", "image"})
	if err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}
