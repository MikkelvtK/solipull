package creleases

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"regexp"
	"strings"
	"time"
)

type listParser struct {
	reUrl *regexp.Regexp
	q     *queue.Queue
}

func NewListParser(months, publishers []string, q *queue.Queue) scraper.ParsingStrategy[*colly.XMLElement] {
	return &listParser{
		reUrl: regexp.MustCompile(generateUrlRegex(months, publishers)),
		q:     q,
	}
}

func (p *listParser) Selector() string {
	return "//loc"
}

func (p *listParser) Parse(e *colly.XMLElement) {
	if !p.reUrl.MatchString(e.Text) {
		return
	}

	if err := p.q.AddURL(e.Text); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("URL found:", e.Text)
}

func (p *listParser) Bind(c *colly.Collector) {
	c.OnXML(p.Selector(), p.Parse)
}

func generateUrlRegex(months []string, publishers []string) string {
	m := strings.Join(months, "|")
	p := strings.Join(publishers, "|")
	y := time.Now().Year()
	return fmt.Sprintf("(?i)(%s)-(%s)-(%d|%d)-solicitations", p, m, y, y+1)
}

type detailParser struct {
	c *cache.Cache
	e scraper.ComicBookExtractor
}

func NewDetailParser(c *cache.Cache, e scraper.ComicBookExtractor) scraper.ParsingStrategy[*colly.HTMLElement] {
	return &detailParser{c, e}
}

func (p *detailParser) Selector() string {
	return "div.wp-block-columns"
}

func (p *detailParser) Parse(e *colly.HTMLElement) {
	cb := models.ComicBook{}
	cb.Publisher = p.e.Publisher(e.Request.URL.String())

	e.DOM.Children().Find("p").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			cb.Title = p.e.Title(s.Text())
			cb.Issue = p.e.Issue(s.Text())
		case 1:
			cb.Pages = p.e.Pages(s.Text())
			cb.Price = p.e.Price(s.Text())
			cb.Creators = p.e.Creators(scraper.Wrap(s))
		case 2:
			cb.ReleaseDate = p.e.ReleaseDate(s.Text())
		}
	})

	if err := p.c.Put(cb); err != nil {
		log.Println(err)
	}
}

func (p *detailParser) Bind(c *colly.Collector) {
	c.OnHTML(p.Selector(), p.Parse)
}
