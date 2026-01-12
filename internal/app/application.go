package app

import (
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"time"
)

type Application struct {
	Scraper scraper.Scraper
	Cache   *cache.Cache
}

func NewApplication(months, publishers []string) *Application {
	c := cache.NewCache()
	e := scraper.NewComicReleasesExtractor(months, publishers)
	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		log.Fatal(err)
	}

	col, err := scraper.NewCollector(service.Domain, 5)
	if err != nil {
		log.Fatal(err)
	}

	navCollector := col.Clone()
	solCollector := col.Clone()
	s := scraper.NewComicReleasesScraper(navCollector, solCollector, q, e)

	return &Application{
		Scraper: s,
		Cache:   c,
	}
}
