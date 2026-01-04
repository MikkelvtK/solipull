package scraper

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type altData struct {
	month     string
	year      int
	publisher string
}

type synopsisData struct {
	bookType string
	price    string
	pages    int
}

// TODO: add proper error handling
func NewLeagueOfComicGeeksScraper(months []string, publishers []string) *StandardScraper {
	const baseEndpoint = "https://leagueofcomicgeeks.com"

	wg := new(sync.WaitGroup)
	scraper := colly.NewCollector()
	results := make(chan models.ComicBook, 50)
	urls := make([]string, 0)

	for _, p := range publishers {
		urls = append(urls, baseEndpoint+"/solicitations/"+p)
	}

	scraper.OnHTML(".card-solicitation", func(e *colly.HTMLElement) {
		a, err := extractLinkAltData(e.ChildAttr("img", "alt"))
		if err != nil {
			return
		}

		if a.year < time.Now().Year() {
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if slices.Contains(months, a.month) {
				l := e.ChildAttr("a", "href")
				if err = e.Request.Visit(strings.Join([]string{baseEndpoint, l}, "")); err != nil {
					return
				}
			}
		}()
	})

	scraper.OnHTML(".issue", func(e *colly.HTMLElement) {
		cb := models.ComicBook{
			Creators: make(map[string][]string),
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			s := e.ChildText(".synopsis + div")
			err := extractSynopsisData(&cb, s)
			if err != nil {
				return
			}

			t := e.ChildText(".comic-summary a")
			extractSummary(&cb, t)

			e.ForEach(".creators .row", func(i int, ec *colly.HTMLElement) {
				c := ec.ChildTexts(".copy-really-small, .copy-small")
				err = extractCreators(&cb, c)
				if err != nil {
					return
				}
			})

			results <- cb
		}()
	})

	return &StandardScraper{
		wg:      wg,
		scraper: scraper,
		urls:    urls,
		results: results,
	}
}

func extractLinkAltData(alt string) (*altData, error) {
	s := strings.Split(alt, " ")
	if len(s) != 5 {
		return nil, fmt.Errorf("invalid alt data: %s", alt)
	}

	y, err := strconv.Atoi(s[3])
	if err != nil {
		return nil, err
	}

	a := &altData{
		month:     strings.ToLower(s[2]),
		year:      y,
		publisher: strings.ToLower(s[0]),
	}

	return a, nil
}

func extractSynopsisData(cb *models.ComicBook, s string) error {
	data := slices.Collect(func(yield func(string) bool) {
		for _, x := range strings.Split(s, "\u00a0") {
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

	if len(data) != 5 {
		return fmt.Errorf("invalid synopsis: %s", s)
	}

	pRaw := strings.Split(data[1], " ")
	if len(pRaw) != 2 {
		return fmt.Errorf("invalid book pages: %s", pRaw)
	}

	pNum, err := strconv.Atoi(pRaw[0])
	if err != nil {
		return err
	}

	cb.Price = data[2]
	cb.Type = strings.ToLower(data[0])
	cb.Pages = pNum
	return nil
}

func extractSummary(cb *models.ComicBook, s string) {
	t := strings.Split(s, "#")
	cb.Title = strings.TrimSpace(t[0])
	if len(t) > 1 {
		n, err := strconv.Atoi(t[1])
		if err != nil {
			cb.Title = s
		} else {
			cb.Issue = n
		}
	}
}

func extractCreators(cb *models.ComicBook, s []string) error {
	if len(s) != 2 {
		return fmt.Errorf("invalid creators: %s", s)
	}

	if _, ok := cb.Creators[s[0]]; !ok {
		cb.Creators[s[0]] = make([]string, 0)
	}

	cb.Creators[s[0]] = append(cb.Creators[s[0]], s[1])
	return nil
}
