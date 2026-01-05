package main

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/scraper"
	"log"
)

func main() {
	s, _ := scraper.NewLeagueOfComicGeeksScraper([]string{"march"}, []string{"dc", "marvel"})
	cb, err := s.Run()
	if err != nil {
		log.Println(err)
	}

	r, _ := cb.GetAll()

	fmt.Println(r)
	fmt.Println(len(r["dc"]) + len(r["marvel"]))
}
