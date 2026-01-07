package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log"
	"regexp"
	"strings"
	"time"
)

var regex *regexp.Regexp

type comicReleasesListParser struct {
	q *queue.Queue
}

func newComicReleasesListParser(q *queue.Queue) *comicReleasesListParser {
	return &comicReleasesListParser{q}
}

func (p *comicReleasesListParser) Selector() string {
	return "//loc"
}

func (p *comicReleasesListParser) Parse(e *colly.XMLElement) {
	if !regex.MatchString(e.Text) {
		return
	}

	if err := p.q.AddURL(e.Text); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("URL found:", e.Text)
}

func NewComicReleasesScraper(months []string, publishers []string) (*Scraper, error) {
	const domain = "comicreleases.com"

	regex = regexp.MustCompile(generateRegex(months, publishers))

	q, err := queue.New(2, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		return nil, err
	}

	c, err := newDefaultCollector(domain)
	if err != nil {
		return nil, err
	}

	listCollector := c.Clone()

	listParser := newComicReleasesListParser(q)

	listCollector.OnXML(listParser.Selector(), listParser.Parse)

	return &Scraper{
		listCollector:   listCollector,
		detailCollector: nil,
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
