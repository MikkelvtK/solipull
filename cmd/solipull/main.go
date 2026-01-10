package main

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/database"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"log"
	"os"
)

func main() {
	cfgPath, _ := os.UserConfigDir()

	db, err := database.InitDB(cfgPath+"/solipull/solipull.db", "sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := cache.NewCache()
	s, err := scraper.NewComicReleasesScraper(c, []string{"march"}, []string{"dc"})
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}

	cbs, err := c.GetAll()

	fmt.Println(cbs)
}
