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
		s.wg.Add(1)
		go s.worker(url, errs)
	}

	go func() {
		s.wg.Wait()
		close(s.results)
		close(errs)
	}()

	c := make([]models.ComicBook, 0)
	for r := range s.results {
		c = append(c, r)
	}

	if err := <-errs; err != nil {
		return nil, err
	}

	return c, nil
}

func (s *StandardScraper) worker(job string, errs chan<- error) {
	defer s.wg.Done()

	if err := s.scraper.Visit(job); err != nil {
		errs <- err
	}
}
