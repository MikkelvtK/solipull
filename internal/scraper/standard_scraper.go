package scraper

import (
	"fmt"
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
	fmt.Printf("Starting worker for %s\n", job)
	if err := s.scraper.Visit(job); err != nil {
		errs <- err
	}
	fmt.Printf("Finished worker for %s\n", job)
}

func NewLeagueOfComicGeeksScraper(publishers []string) *StandardScraper {
	const baseEndpoint = "https://leagueofcomicgeeks.com/solicitations/"

	wg := new(sync.WaitGroup)
	scraper := colly.NewCollector()
	results := make(chan models.ComicBook, 50)
	urls := make([]string, 0)

	for _, p := range publishers {
		urls = append(urls, baseEndpoint+p)
	}

	return &StandardScraper{
		wg:      wg,
		scraper: scraper,
		urls:    urls,
		results: results,
	}
}
