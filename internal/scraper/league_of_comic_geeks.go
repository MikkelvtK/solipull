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

func NewLeagueOfComicGeeksScraper(months []string, publishers []string) *StandardScraper {
	const baseEndpoint = "https://leagueofcomicgeeks.com"

	wg := new(sync.WaitGroup)
	scraper := colly.NewCollector()
	results := make(chan models.ComicBook, 50)
	urls := make([]string, 0)
	errs := make(chan error, 0)

	for _, p := range publishers {
		urls = append(urls, baseEndpoint+"/solicitations/"+p)
	}

	scraper.OnHTML(".card-solicitation", func(e *colly.HTMLElement) {
		a, err := extractLinkAltData(e.ChildAttr("img", "alt"))
		if err != nil {
			errs <- err
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

				e.Request.Ctx.Put("publisher", a.publisher)
				if err = e.Request.Visit(strings.Join([]string{baseEndpoint, l}, "")); err != nil {
					errs <- err
					return
				}
			}
		}()
	})

	scraper.OnHTML(".issue", func(e *colly.HTMLElement) {
		cb := models.ComicBook{
			Creators: make(map[string][]string),
		}

		err := extractSynopsisData(&cb, e.ChildText(".synopsis + div"))
		if err != nil {
			errs <- err
			return
		}

		err = extractSummary(&cb, e.ChildTexts(".comic-summary .copy-really-large, "+
			".comic-summary .copy-really-small"))
		if err != nil {
			errs <- err
			return
		}

		e.ForEach(".creators .row", func(i int, ec *colly.HTMLElement) {
			err = extractCreators(&cb, ec.ChildTexts(".copy-really-small, .copy-small"))
			if err != nil {
				errs <- err
				return
			}
		})

		cb.Publisher = e.Request.Ctx.Get("publisher")

		results <- cb
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

	if len(data) <= 3 {
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

func extractSummary(cb *models.ComicBook, s []string) error {
	titleRaw := s[1]
	titleSplit := strings.Split(titleRaw, "#")
	cb.Title = strings.TrimSpace(titleSplit[0])
	if len(titleSplit) > 1 {
		n, err := strconv.Atoi(titleSplit[1])
		if err != nil {
			cb.Title = titleRaw
		} else {
			cb.Issue = n
		}
	}

	for _, l := range []string{"Jan 2nd, 2006", "Jan 2st, 2006", "Jan 2th, 2006"} {
		d, err := time.Parse(l, s[0])
		if err == nil {
			cb.ReleaseDate = d
			break
		}
	}

	return nil
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
