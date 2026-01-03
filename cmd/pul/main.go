package main

import (
	"github.com/MikkelvtK/pul/internal/scraper"
)

func main() {
	s := scraper.NewLeagueOfComicGeeksScraper([]string{"dc", "marvel", "image"})
	s.Scrape()
}
