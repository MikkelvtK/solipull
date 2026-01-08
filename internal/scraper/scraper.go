package scraper

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"time"
)

type ParsingStrategy[T any] interface {
	Selector() string
	Parse(e T)
}

type Scraper struct {
	listCollector   *colly.Collector
	detailCollector *colly.Collector
	queue           *queue.Queue
	url             string
}

func (s *Scraper) Run() error {
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

func newDefaultCollector(domain string) (*colly.Collector, error) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
	)

	c.IgnoreRobotsTxt = false

	err := c.Limit(&colly.LimitRule{DomainGlob: domain, Parallelism: 1, RandomDelay: 5 * time.Second})
	if err != nil {
		return nil, err
	}

	return c, nil
}
