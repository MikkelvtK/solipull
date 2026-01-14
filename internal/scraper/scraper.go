package scraper

import (
	"context"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/MikkelvtK/solipull/internal/service"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var (
	reBrackets = regexp.MustCompile(`[(\[].*?[)\]]`)
	reAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
)

func NewCollector(domain string, parallelism int) (*colly.Collector, error) {
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

	l := &colly.LimitRule{DomainGlob: "*" + domain + "*", Parallelism: parallelism, RandomDelay: 5 * time.Second}
	if err := c.Limit(l); err != nil {
		return nil, err
	}

	return c, nil
}

type comicReleasesScraper struct {
	navCol *colly.Collector
	solCol *colly.Collector
	queue  *queue.Queue
	ex     ComicBookExtractor
	logger *slog.Logger

	observer service.ScrapingObserver
	ctx      context.Context
	res      chan<- models.ComicBook
}

func (s *comicReleasesScraper) GetData(ctx context.Context, url string, results chan<- models.ComicBook) error {
	s.ctx = ctx
	s.res = results

	defer func() {
		s.ctx = nil
		s.res = nil
	}()

	if err := s.navCol.Visit(url); err != nil {
		return err
	}
	s.navCol.Wait()

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := s.queue.Run(s.solCol); err != nil {
		return err
	}
	s.solCol.Wait()

	return nil
}

func (s *comicReleasesScraper) bindCallbacks() {
	checkCtx := func(r *colly.Request) {
		if s.ctx != nil && s.ctx.Err() != nil {
			r.Abort()
		}
	}

	logErr := func(r *colly.Response, e error) {
		s.logger.Error("request failed",
			"url", r.Request.URL,
			"status", r.StatusCode,
			"error", e.Error())
		s.observer.OnError(1)
	}

	s.navCol.OnRequest(checkCtx)
	s.solCol.OnRequest(checkCtx)

	s.navCol.OnError(logErr)
	s.solCol.OnError(logErr)

	s.navCol.OnXML("//loc", func(e *colly.XMLElement) {
		if !s.ex.MatchURL(e.Text) {
			return
		}

		if err := s.queue.AddURL(e.Text); err != nil {
			s.logger.Warn("failed to add url to queue",
				"url", e.Text,
				"err", err.Error())
			s.observer.OnError(1)
			return
		}

		s.observer.OnUrlFound(1)
	})

	s.solCol.OnHTML("div.wp-block-columns", func(e *colly.HTMLElement) {
		cb := s.parseComicBook(e)
		if s.res != nil {
			s.observer.OnComicBookScraped(1)
			s.res <- cb
		}
	})
}

func (s *comicReleasesScraper) parseComicBook(e *colly.HTMLElement) models.ComicBook {
	var fullTitle string
	cb := models.ComicBook{}
	cb.Publisher = s.ex.Publisher(e.Request.URL.String())
	cb.Format, _ = e.DOM.PrevAll().Filter("#singles, #trades, #hardcovers").First().Attr("id")

	e.DOM.Children().Find("p").Each(func(i int, sel *goquery.Selection) {
		switch i {
		case 0:
			fullTitle = sel.Text()
			cb.Title = s.ex.Title(sel.Text())
			cb.Issue = s.ex.Issue(sel.Text())
		case 1:
			cb.Pages = s.ex.Pages(sel.Text())
			cb.Price = s.ex.Price(sel.Text())
			cb.Creators = s.ex.Creators(Wrap(sel))
		case 2:
			cb.ReleaseDate = s.ex.ReleaseDate(sel.Text())
		}
	})

	if !cb.ReleaseDate.IsZero() {
		return cb
	}

	e.DOM.NextAllFiltered(":contains('ON-SALE'), :contains('FOC')").Each(func(_ int, p *goquery.Selection) {
		p.Next().Find("li").Each(func(_ int, pe *goquery.Selection) {
			if strings.EqualFold(normalizeTitle(pe.Text()), normalizeTitle(fullTitle)) {
				if strings.Contains(pe.Text(), "ON SALE") {
					cb.ReleaseDate = s.ex.ReleaseDate(pe.Text())
				} else {
					cb.ReleaseDate = s.ex.ReleaseDate(p.Text())
				}
			}
		})
	})

	if cb.ReleaseDate.IsZero() {
		s.logger.Warn("no release date found", "string", s)
		s.observer.OnError(1)
	}
	return cb
}

func normalizeTitle(s string) string {
	s = strings.ToLower(s)
	s = reBrackets.ReplaceAllString(s, "")
	s = reAlphaNum.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(s), " ")
}

type SConfig struct {
	Nav    *colly.Collector
	Sol    *colly.Collector
	Q      *queue.Queue
	Ex     ComicBookExtractor
	Logger *slog.Logger
}

func NewComicReleasesScraper(cfg *SConfig) models.DataProvider {
	s := &comicReleasesScraper{
		navCol: cfg.Nav,
		solCol: cfg.Sol,
		queue:  cfg.Q,
		ex:     cfg.Ex,
		logger: cfg.Logger,
	}

	s.bindCallbacks()
	return s
}
