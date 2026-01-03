package scraper

import (
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"sync"
)

type StandardScraper struct {
	wg      *sync.WaitGroup
	scraper *colly.Collector
	urls    []string
	results chan models.ComicBook
}

func (s *StandardScraper) Scrape() ([]models.ComicBook, error) {
	errs := make(chan error)

	for _, url := range s.urls {
		go s.worker(url, errs)
	}

	return nil, nil
}

func (s *StandardScraper) worker(job string, errs chan<- error) {
	defer s.wg.Done()
	if err := s.scraper.Visit(job); err != nil {
		errs <- err
	}
}

func NewLeagueOfComicGeeksScraper(months []string, publishers []string) *StandardScraper {

	return &StandardScraper{}
}
