package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log"
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
	return nil
}

func newDefaultCollector(domain string) (*colly.Collector, error) {
	// TODO: Add random user agent capabilities

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
	)

	err := c.Limit(&colly.LimitRule{DomainGlob: domain, Parallelism: 1, RandomDelay: 5 * time.Second})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func registerDefaultCallbacks(c *colly.Collector) *colly.Collector {
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("starting scraping for %s\n", r.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Printf("finished scraping %s\n", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("scraping failed: %s with error %s\n", r.Request.URL, err.Error())
	})

	return c
}
