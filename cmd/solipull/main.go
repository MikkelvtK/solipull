package main

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"log"
)

func main() {
	c := cache.NewCache()

	s, err := scraper.NewComicReleasesScraper(c, []string{"march"}, []string{"dc", "marvel"})
	if err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}

	cbs, _ := c.GetAll()
	fmt.Println(len(cbs))
	fmt.Println(cbs)
}
