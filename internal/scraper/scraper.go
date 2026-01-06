package scraper

import (
    "fmt"
    "github.com/MikkelvtK/solipull/internal/cache"
    "github.com/MikkelvtK/solipull/internal/models"
    "github.com/gocolly/colly/v2"
    "strings"
    "time"
)

type Scraper interface {
    Run() (*cache.Cache[string, models.ComicBook], error)
}

type ParsingStrategy interface {
    Selector() string
    Parse(e *colly.HTMLElement)
}

func newDefaultCollector(domain string) (*colly.Collector, error) {
    // TODO: Add random user agent capabilities

    c := colly.NewCollector(
        colly.Async(true),
        colly.MaxDepth(2),
        colly.CacheDir(fmt.Sprintf("./%s_cache", strings.ReplaceAll(domain, "https:", ""))),
    )

    err := c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2, RandomDelay: time.Second})
    if err != nil {
        return nil, err
    }

    return c, nil
}
