package main

import (
	"github.com/MikkelvtK/pul/internal/scraper"
)

func main() {
	s := scraper.NewLeagueOfComicGeeksScraper()
	s.Scrape("Jan", []string{"DC", "Marvel", "Image"})
}
