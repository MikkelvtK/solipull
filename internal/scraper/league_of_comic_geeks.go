package scraper

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"slices"
	"strconv"
	"strings"
	"time"
)

type altData struct {
	month     string
	year      int
	publisher string
}

func NewLeagueOfComicGeeksScraper(months []string, publishers []string) (*StandardScraper, error) {
	const baseEndpoint = "https://leagueofcomicgeeks.com"

	listScraper := colly.NewCollector(
		colly.AllowedDomains("leagueofcomicgeeks.com"),
		colly.Async(true),
		colly.MaxDepth(2),
		colly.CacheDir("./league_of_comic_geeks_cache"),
	)

	err := listScraper.Limit(&colly.LimitRule{DomainGlob: "*leagueofcomicgeeks*", Parallelism: 2, RandomDelay: time.Second})
	if err != nil {
		return nil, err
	}

	detailScraper := listScraper.Clone()
	results := make(chan models.ComicBook, 50)
	urls := make([]string, 0)
	errs := make(chan error)

	for _, p := range publishers {
		urls = append(urls, baseEndpoint+"/solicitations/"+p)
	}

	// TODO: use fake user agents
	listScraper.OnRequest(func(r *colly.Request) {
		fmt.Println("searching", r.URL)
	})

	listScraper.OnHTML(".card-solicitation", func(e *colly.HTMLElement) {
		a, err := extractLinkAltData(e.ChildAttr("img", "alt"))
		if err != nil {
			errs <- err
			return
		}

		if a.year < time.Now().Year() {
			return
		}

		if slices.Contains(months, a.month) {
			l := e.ChildAttr("a", "href")

			if err = detailScraper.Visit(strings.Join([]string{baseEndpoint, l}, "")); err != nil {
				errs <- err
				return
			}

			detailScraper.Wait()
		}
	})

	detailScraper.OnRequest(func(r *colly.Request) {
		fmt.Printf("starting scraping for %s\n", r.URL)
	})

	detailScraper.OnHTML(".issue", func(e *colly.HTMLElement) {
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

		fmt.Println(cb)

		results <- cb
	})

	detailScraper.OnScraped(func(r *colly.Response) {
		fmt.Printf("finished scraping %s\n", r.Request.URL)
	})

	fmt.Println("Scraper initialized")
	return &StandardScraper{
			scraper: listScraper,
			urls:    urls,
			results: results,
			errs:    errs,
		},
		nil
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
