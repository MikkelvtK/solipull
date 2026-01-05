package scraper

import (
	"github.com/MikkelvtK/pul/internal/cache"
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
)

type StandardScraper struct {
	collector   *colly.Collector
	strategies  []ParsingStrategy
	urls        []string
	resultCache *cache.Cache[string, models.ComicBook]
}

func (s *StandardScraper) Run() (*cache.Cache[string, models.ComicBook], error) {
	for _, strat := range s.strategies {
		s.collector.OnHTML(strat.Selector(), strat.Parse)
	}

	for _, u := range s.urls {
		if err := s.collector.Visit(u); err != nil {
			return nil, err
		}
	}

	s.collector.Wait()
	return s.resultCache, nil
}
