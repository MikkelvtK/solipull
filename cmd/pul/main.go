package main

import (
	"github.com/MikkelvtK/pul/internal/scraper"
)

func main() {
	s := scraper.NewLeagueOfComicGeeksScraper([]string{"march"}, []string{"dc", "marvel"})
	s.Scrape()
}
