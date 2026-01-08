package scraper

import (
	"errors"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"regexp"
	"slices"
	"strings"
	"time"
)

type crComicRegex struct {
	url         *regexp.Regexp
	publisher   *regexp.Regexp
	pages       *regexp.Regexp
	price       *regexp.Regexp
	creators    *regexp.Regexp
	releaseDate *regexp.Regexp
}

func newCrComicRegex(months []string, publishers []string) *crComicRegex {
	return &crComicRegex{
		url:         regexp.MustCompile(generateUrlRegex(months, publishers)),
		publisher:   regexp.MustCompile("(?i)/(\\w+)-[a-zA-Z]+-\\d{4}-solicitations"),
		pages:       regexp.MustCompile("(?i)(\\d+)\\s*(?:pages?|pgs?\\.?)"),
		price:       regexp.MustCompile("\\$(\\d+\\.\\d{2})"),
		releaseDate: regexp.MustCompile("(?i)(\\d{1,2}/\\d{1,2}/\\d{2,4})"),
	}
}

type crListParser struct {
	reg *crComicRegex
	q   *queue.Queue
}

func newCrListParser(reg *crComicRegex, q *queue.Queue) *crListParser {
	return &crListParser{reg, q}
}

func (p *crListParser) Selector() string {
	return "//loc"
}

func (p *crListParser) Parse(e *colly.XMLElement) {
	if !p.reg.url.MatchString(e.Text) {
		return
	}

	r, err := e.Request.New("GET", e.Text, nil)
	if err != nil {
		log.Println(err)
		return
	}

	g := p.reg.publisher.FindStringSubmatch(e.Text)
	if len(g) > 1 {
		r.Ctx.Put("publisher", g[1])
	}

	if err := p.q.AddRequest(r); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("URL found:", e.Text)
}

type crDetailParser struct {
	c *cache.Cache
	e *extractor
}

func newCrDetailParser(c *cache.Cache, e *extractor) *crDetailParser {
	return &crDetailParser{c, e}
}

func (p *crDetailParser) Selector() string {
	return "div.wp-block-columns"
}

func (p *crDetailParser) Parse(e *colly.HTMLElement) {
	cb := models.ComicBook{}
	cb.Publisher = e.Request.Ctx.Get("publisher")

	e.DOM.Children().Find("p").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			cb.Title = p.e.extract(s.Text(), crTitle)
			cb.Issue = p.e.extract(s.Text(), crIssue)
		case 1:
			cb.Pages = p.e.extract(s.Text(), crPages)
			cb.Price = p.e.extract(s.Text(), crPrice)
			cb.Creators = p.e.extractCreators(s)
		case 2:
			cb.ReleaseDate = crParseTime(s.Text(), p.e)
		}
	})

	if err := p.c.Put(cb); err != nil {
		log.Println(err)
	}
}

func NewComicReleasesScraper(cache *cache.Cache, months []string, publishers []string) (*Scraper, error) {
	const domain = "comicreleases.com"

	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		return nil, err
	}

	c, err := newDefaultCollector(domain)
	if err != nil {
		return nil, err
	}

	r := newCrComicRegex(months, publishers)
	e := newExtractor(r)

	listCollector := c.Clone()
	listParser := newCrListParser(r, q)
	listCollector.OnXML(listParser.Selector(), listParser.Parse)

	limit := &colly.LimitRule{DomainGlob: domain, Parallelism: 5, RandomDelay: 5 * time.Second}
	detailCollector := c.Clone()
	detailCollector.CacheDir = fmt.Sprintf("./%s_cache", strings.ReplaceAll(domain, "https:", ""))
	if err = detailCollector.Limit(limit); err != nil {
		return nil, err
	}

	detailParser := newCrDetailParser(cache, e)
	detailCollector.OnHTML(detailParser.Selector(), detailParser.Parse)

	return &Scraper{
		listCollector:   listCollector,
		detailCollector: detailCollector,
		queue:           q,
		url:             "https://" + domain + "/sitemap.xml",
	}, nil
}

func generateUrlRegex(months []string, publishers []string) string {
	m := strings.Join(months, "|")
	p := strings.Join(publishers, "|")
	y := time.Now().Year()
	return fmt.Sprintf("(?i)(%s)-(%s)-(%d|%d)-solicitations", p, m, y, y+1)
}

type extractor struct {
	regex *crComicRegex
}

func newExtractor(regex *crComicRegex) *extractor {
	return &extractor{regex}
}

func (e *extractor) extract(s string, extractFunc func(string, *crComicRegex) (string, error)) string {
	result, err := extractFunc(s, e.regex)
	if err != nil {
		log.Println(err)
		return ""
	}

	return result
}

func crTitle(s string, _ *crComicRegex) (string, error) {
	split := strings.Split(s, "#")
	if len(split) == 0 {
		return "", errors.New("no title found")
	}

	return cases.Title(language.English).String(split[0]), nil
}

func crIssue(s string, _ *crComicRegex) (string, error) {
	split := strings.Split(s, "#")
	if len(split) < 2 {
		return "", nil
	}

	return cases.Title(language.English).String(split[1]), nil
}

func crPages(s string, regex *crComicRegex) (string, error) {
	return regex.pages.FindString(s), nil
}

func crPrice(s string, regex *crComicRegex) (string, error) {
	return regex.price.FindString(s), nil
}

func crReleaseDate(s string, regex *crComicRegex) (string, error) {
	d := regex.releaseDate.FindString(s)
	if len(d) == 0 {
		return "", errors.New("no release date found")
	}

	return d, nil
}

func crParseTime(s string, e *extractor) time.Time {
	d := e.extract(s, crReleaseDate)
	if d == "" {
		return time.Time{}
	}

	t, err := time.Parse("1/2/06", d)
	if err != nil {
		log.Println(err)
		return time.Time{}
	}

	return t
}

func (e *extractor) extractCreators(s *goquery.Selection) map[string][]string {
	results := make(map[string][]string)

	s.Contents().Each(func(j int, t *goquery.Selection) {
		if goquery.NodeName(t) == "br" {
			return
		}

		roles := []string{"writer", "artist", "cover artist"}
		for _, role := range roles {
			v := strings.ToLower(strings.TrimSpace(t.Text()))

			if !strings.HasPrefix(v, role) {
				continue
			}

			split := strings.Split(v, ":")
			names := strings.Split(split[1], ",")
			names = slices.Collect(func(yield func(string) bool) {
				for _, name := range names {
					namesNoAnd := strings.ReplaceAll(name, "and", "")
					namesNoSpace := strings.TrimSpace(namesNoAnd)
					yield(cases.Title(language.English).String(namesNoSpace))
				}
			})

			results[role] = names
		}
	})

	return results
}
