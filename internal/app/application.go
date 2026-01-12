package app

import (
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
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
	e := scraper.NewComicReleasesExtractor()
	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		log.Fatal(err)
	}

	listCollector, err := scraper.NewDefaultCollector(scraper.Domain, nil)
	if err != nil {
		log.Fatal(err)
	}
	listParser := scraper.NewListParser(months, publishers, q)
	listParser.Bind(listCollector)

	detailCollector, err := scraper.NewDefaultCollector(scraper.Domain,
		&colly.LimitRule{DomainGlob: scraper.Domain, Parallelism: 5, RandomDelay: 5 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	detailParser := scraper.NewDetailParser(c, e)
	detailParser.Bind(detailCollector)

	s := scraper.NewComicReleasesScraper(listCollector, detailCollector, q)

	return &Application{
		Scraper: s,
		Cache:   c,
	}
}
