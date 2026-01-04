package main

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/scraper"
)

func main() {
	s := scraper.NewLeagueOfComicGeeksScraper([]string{"march"}, []string{"dc"})
	cb, _ := s.Scrape()

	fmt.Println(cb)
}
