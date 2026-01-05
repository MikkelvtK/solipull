package scraper

import (
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
)

type StandardScraper struct {
	scraper *colly.Collector
	urls    []string
	results chan models.ComicBook
	errs    chan error
}

func (s *StandardScraper) Scrape() ([]models.ComicBook, error) {
	for _, url := range s.urls {
		if err := s.scraper.Visit(url); err != nil {
			s.errs <- err
		}
	}

	go func() {
		s.scraper.Wait()
		close(s.results)
		close(s.errs)
	}()

	c := make([]models.ComicBook, 0)
	for r := range s.results {
		c = append(c, r)
	}

	if err := <-s.errs; err != nil {
		return nil, err
	}

	return c, nil
}
