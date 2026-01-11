package app

import (
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/MikkelvtK/solipull/internal/scraper/creleases"
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
	e := creleases.NewComicReleasesExtractor()
	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		log.Fatal(err)
	}

	listCollector, err := scraper.NewDefaultCollector(creleases.Domain, nil)
	if err != nil {
		log.Fatal(err)
	}
	listParser := creleases.NewListParser(months, publishers, q)
	listParser.Bind(listCollector)

	detailCollector, err := scraper.NewDefaultCollector(creleases.Domain,
		&colly.LimitRule{DomainGlob: creleases.Domain, Parallelism: 5, RandomDelay: 5 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	detailParser := creleases.NewDetailParser(c, e)
	detailParser.Bind(detailCollector)

	s := creleases.NewComicReleasesScraper(listCollector, detailCollector, q)

	return &Application{
		Scraper: s,
		Cache:   c,
	}
}
