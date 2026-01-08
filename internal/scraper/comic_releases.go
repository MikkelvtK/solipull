package scraper

import (
	"errors"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"regexp"
	"strings"
	"time"
)

type crComicRegex struct {
	url   *regexp.Regexp
	pages *regexp.Regexp
	price *regexp.Regexp
}

func newCrComicRegex(months []string, publishers []string) *crComicRegex {
	return &crComicRegex{
		url:   regexp.MustCompile(generateRegex(months, publishers)),
		pages: regexp.MustCompile("(?i)(\\d+)\\s*(?:pages?|pgs?\\.?)"),
		price: regexp.MustCompile("\\$(\\d+\\.\\d{2})"),
	}
}

type crListParser struct {
	reg *regexp.Regexp
	q   *queue.Queue
}

func newCrListParser(reg *regexp.Regexp, q *queue.Queue) *crListParser {
	return &crListParser{reg, q}
}

func (p *crListParser) Selector() string {
	return "//loc"
}

func (p *crListParser) Parse(e *colly.XMLElement) {
	if !p.reg.MatchString(e.Text) {
		return
	}

	if err := p.q.AddURL(e.Text); err != nil {
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

func (c *crDetailParser) Selector() string {
	return "div.wp-block-columns"
}

func (c *crDetailParser) Parse(e *colly.HTMLElement) {
	data := e.ChildTexts("p")
	if len(data) < 3 {
		log.Println("Incorrect number of elements found:", len(data))
		return
	}

	cb := models.ComicBook{}
	cb.Title = c.e.extract(data[0], crTitle)
	cb.Issue = c.e.extract(data[0], crIssue)
	cb.Pages = c.e.extract(data[1], crPages)
	cb.Price = c.e.extract(data[1], crPrice)

	fmt.Println(cb)
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
	listParser := newCrListParser(r.url, q)
	listCollector.OnXML(listParser.Selector(), listParser.Parse)

	detailCollector := c.Clone()
	detailCollector.CacheDir = fmt.Sprintf("./%s_cache", strings.ReplaceAll(domain, "https:", ""))
	err = detailCollector.Limit(&colly.LimitRule{DomainGlob: domain, Parallelism: 5, RandomDelay: 5 * time.Second})
	if err != nil {
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

func generateRegex(months []string, publishers []string) string {
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

	return split[0], nil
}

func crIssue(s string, _ *crComicRegex) (string, error) {
	split := strings.Split(s, "#")
	if len(split) < 2 {
		return "", nil
	}

	return split[1], nil
}

func crPages(s string, regex *crComicRegex) (string, error) {
	return regex.pages.FindString(s), nil
}

func crPrice(s string, regex *crComicRegex) (string, error) {
	return regex.price.FindString(s), nil
}
