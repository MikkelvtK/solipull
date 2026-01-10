package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"regexp"
	"strings"
	"time"
)

type Scraper interface {
	Run() error
}

type ParsingStrategy[T any] interface {
	Selector() string
	Parse(e T)
	Bind(c *colly.Collector)
}

func NewDefaultCollector(domain string, limit *colly.LimitRule) (*colly.Collector, error) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
	)

	domainSplit := strings.Split(domain, ".")
	if len(domainSplit) != 2 {
		return nil, fmt.Errorf("invalid domain: %s", domain)
	}

	regStr := fmt.Sprintf("^(https?://)?([\\w\\d-]+\\.)?%s\\.%s(/.*)?$", domainSplit[0], domainSplit[1])
	c.URLFilters = []*regexp.Regexp{
		regexp.MustCompile(regStr),
	}

	c.IgnoreRobotsTxt = false

	limitToUse := limit
	if limitToUse == nil {
		limitToUse = &colly.LimitRule{DomainGlob: "*" + domain + "*", Parallelism: 1, RandomDelay: 5 * time.Second}
	}
	err := c.Limit(limitToUse)
	if err != nil {
		return nil, err
	}

	return c, nil
}
