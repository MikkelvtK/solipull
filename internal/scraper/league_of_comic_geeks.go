package scraper

import (
	"fmt"
	"github.com/MikkelvtK/pul/internal/cache"
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"log"
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

type SolicitationParser struct {
	domain string
	months []string
}

func NewSolicitationParser(domain string, months []string) *SolicitationParser {
	return &SolicitationParser{
		domain: domain,
		months: months,
	}
}

func (s *SolicitationParser) Selector() string {
	return ".card-solicitation"
}

func (s *SolicitationParser) Parse(e *colly.HTMLElement) {
	a, err := extractLinkAltData(e.ChildAttr("img", "alt"))
	if err != nil {
		log.Println(err)
		return
	}

	if a.year < time.Now().Year() {
		return
	}

	if slices.Contains(s.months, a.month) {
		l := e.ChildAttr("a", "href")

		e.Request.Ctx.Put("publisher", a.publisher)
		if err = e.Request.Visit(strings.Join([]string{s.domain, l}, "")); err != nil {
			log.Println(err)
			return
		}
	}
}

type IssueParser struct {
	cache *cache.Cache[string, models.ComicBook]
}

func NewIssueParser(c *cache.Cache[string, models.ComicBook]) *IssueParser {
	return &IssueParser{
		cache: c,
	}
}

func (i IssueParser) Selector() string {
	return ".issue"
}

func (i IssueParser) Parse(e *colly.HTMLElement) {
	cb := models.ComicBook{
		Creators: make(map[string][]string),
	}

	err := extractSynopsisData(&cb, e.ChildText(".synopsis + div"))
	if err != nil {
		log.Println(err)
		return
	}

	err = extractSummary(&cb, e.ChildTexts(".comic-summary .copy-really-large, "+
		".comic-summary .copy-really-small"))
	if err != nil {
		log.Println(err)
		return
	}

	e.ForEach(".creators .row", func(i int, ec *colly.HTMLElement) {
		err = extractCreators(&cb, ec.ChildTexts(".copy-really-small, .copy-small"))
		if err != nil {
			log.Println(err)
			return
		}
	})

	cb.Publisher = e.Request.Ctx.Get("publisher")
	i.cache.Put(cb.Publisher, cb)
}

func NewLeagueOfComicGeeksScraper(months []string, publishers []string) (*StandardScraper, error) {
	const baseEndpoint = "https://leagueofcomicgeeks.com"

	c, err := newDefaultCollector(baseEndpoint)
	if err != nil {
		return nil, err
	}

	r := cache.NewCache[string, models.ComicBook]()
	u := make([]string, 0)

	s := NewSolicitationParser(baseEndpoint, months)
	i := NewIssueParser(r)

	for _, p := range publishers {
		u = append(u, baseEndpoint+"/solicitations/"+p)
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("starting scraping for %s\n", r.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Printf("finished scraping %s\n", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("scraping failed: %s with error %s\n", r.Request.URL, err.Error())
	})

	return &StandardScraper{
			collector:   c,
			urls:        u,
			strategies:  []ParsingStrategy{s, i},
			resultCache: r,
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
