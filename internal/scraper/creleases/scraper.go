package creleases

import (
	"errors"
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

func NewComicReleasesScraper(options ...OptionsFunc) (scraper.Scraper, error) {
	s := &comicReleasesScraper{}

	for _, opt := range options {
		opt(s)
	}

	if s.listCollector == nil {
		return nil, errors.New("a listCollector must be provided to the scraper")
	}

	if s.detailCollector == nil {
		return nil, errors.New("a detailCollector must be provided to the scraper")
	}

	if s.queue == nil {
		q, err := queue.New(1, &queue.InMemoryQueueStorage{MaxSize: 10_000})
		if err != nil {
			return nil, err
		}

		s.queue = q
	}

	return s, nil
}
