package creleases

import (
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

const (
	Domain = "comicreleases.com"
)

type comicReleasesScraper struct {
	listCollector   *colly.Collector
	detailCollector *colly.Collector
	queue           *queue.Queue
	url             string
}

func (s *comicReleasesScraper) Run() error {
	if err := s.listCollector.Visit(s.url); err != nil {
		return err
	}

	s.listCollector.Wait()

	if err := s.queue.Run(s.detailCollector); err != nil {
		return err
	}

	s.detailCollector.Wait()
	return nil
}

type OptionsFunc func(*comicReleasesScraper)

func NewComicReleasesScraper(list, detail *colly.Collector, q *queue.Queue) scraper.Scraper {
	return &comicReleasesScraper{
		listCollector:   list,
		detailCollector: detail,
		queue:           q,
		url:             "https://www." + Domain + "/sitemap.xml",
	}
}
