package scraper

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"slices"
	"strconv"
	"strings"
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
		fmt.Println(url)
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
	const baseEndpoint = "https://leagueofcomicgeeks.com"

	wg := new(sync.WaitGroup)
	scraper := colly.NewCollector()
	results := make(chan models.ComicBook, 50)
	urls := make([]string, 0)

	for _, p := range publishers {
		urls = append(urls, baseEndpoint+"/solicitations/"+p)
	}

	scraper.OnHTML(".card-solicitation", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a", "href")
		if strings.Contains(link, "march") && strings.Contains(link, "2026") {
			fmt.Printf("Found march link: %s\n", link)
			e.Request.Visit(strings.Join([]string{baseEndpoint, link}, ""))
		}
	})

	scraper.OnHTML(".issue", func(e *colly.HTMLElement) {
		cb := models.ComicBook{
			Creators: make(map[string][]string),
		}

		data := slices.Collect(func(yield func(string) bool) {
			for _, x := range strings.Split(e.ChildText(".synopsis + div"), "\u00a0") {
				if strings.Contains(x, "·") {
					continue
				}

				if strings.TrimSpace(x) == "" {
					continue
				}

				if !yield(strings.TrimSpace(x)) {
					return
				}
			}
		})

		cb.Type = data[0]
		cb.Pages, _ = strconv.Atoi(strings.Split(data[1], " ")[0])
		cb.Price = data[2]

		t := strings.Split(e.ChildText(".comic-summary a"), "#")
		cb.Title = strings.TrimSpace(t[0])
		if len(t) > 1 {
			n, err := strconv.Atoi(t[1])
			if err != nil {
				fmt.Println(err)
			}

			cb.Issue = n
		}

		e.ForEach(".creators .row", func(i int, ec *colly.HTMLElement) {
			c := ec.ChildTexts(".copy-really-small, .copy-small")
			if len(c) != 2 {
				// TODO: logging
				return
			}

			if _, ok := cb.Creators[c[0]]; !ok {
				cb.Creators[c[0]] = make([]string, 0)
			}

			cb.Creators[c[0]] = append(cb.Creators[c[0]], c[1])
		})

		fmt.Println(cb)
		results <- cb
	})

	return &StandardScraper{
		wg:      wg,
		scraper: scraper,
		urls:    urls,
		results: results,
	}
}
