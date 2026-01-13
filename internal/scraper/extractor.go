package scraper

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

type ComicBookExtractor interface {
	MatchURL(string) bool
	Title(string) string
	Issue(string) string
	Pages(string) string
	Price(string) string
	Publisher(string) string
	Creators(HTMLNode) []models.Creator
	ReleaseDate(string) time.Time
}

type comicReleasesExtractor struct {
	reUrl         *regexp.Regexp
	rePublisher   *regexp.Regexp
	rePages       *regexp.Regexp
	rePrice       *regexp.Regexp
	reReleaseDate *regexp.Regexp
	creatorParser *creatorParser
	logger        *slog.Logger
	stats         *models.RunStats
}

func NewComicReleasesExtractor(months, publishers []string, l *slog.Logger, s *models.RunStats) ComicBookExtractor {
	return &comicReleasesExtractor{
		reUrl:         regexp.MustCompile(generateUrlRegex(months, publishers)),
		rePublisher:   regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
		rePages:       regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
		rePrice:       regexp.MustCompile(`\$(\d+\.\d{2})`),
		reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
		creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
		logger:        l,
		stats:         s,
	}
}

func (c *comicReleasesExtractor) MatchURL(url string) bool {
	return c.reUrl.MatchString(url)
}

func (c *comicReleasesExtractor) Title(s string) string {
	split := strings.Split(s, "#")
	if len(split[0]) == 0 {
		c.logger.Warn("title not found", "string", s)
		c.stats.ErrorCount.Add(1)
		return ""
	}

	title := cases.Title(language.English).String(split[0])
	return strings.TrimSpace(title)
}

func (c *comicReleasesExtractor) Issue(s string) string {
	split := strings.Split(s, "#")
	if len(split) < 2 {
		return ""
	}

	issue := cases.Title(language.English).String(split[1])
	return strings.TrimSpace(issue)
}

func (c *comicReleasesExtractor) Pages(s string) string {
	if c.rePages == nil {
		c.logger.Warn("pages regex is nil")
		c.stats.ErrorCount.Add(1)
		return ""
	}

	matches := c.rePages.FindStringSubmatch(s)
	if matches == nil {
		c.logger.Warn("no matches for pages found", "string", s, "matches", matches)
		c.stats.ErrorCount.Add(1)
		return ""
	}

	i := c.rePages.SubexpIndex("Pages")
	if i < 0 {
		c.logger.Warn("no index for pages found", "string", s, "i", i)
		c.stats.ErrorCount.Add(1)
		return ""
	}

	return matches[i]
}

func (c *comicReleasesExtractor) Price(s string) string {
	if c.rePrice == nil {
		c.logger.Warn("price regex is nil")
		c.stats.ErrorCount.Add(1)
		return ""
	}
	return c.rePrice.FindString(s)
}

func (c *comicReleasesExtractor) Publisher(s string) string {
	if c.rePublisher == nil {
		c.logger.Warn("publisher regex is nil")
		c.stats.ErrorCount.Add(1)
		return ""
	}

	matches := c.rePublisher.FindStringSubmatch(s)
	if matches == nil {
		c.logger.Warn("no matches for publisher found", "string", s, "matches", matches)
		c.stats.ErrorCount.Add(1)
		return ""
	}

	i := c.rePublisher.SubexpIndex("Pub")
	if i < 0 {
		c.logger.Warn("no index for publisher found", "string", s, "i", i)
		c.stats.ErrorCount.Add(1)
		return ""
	}
	return strings.ToLower(matches[i])
}

func (c *comicReleasesExtractor) Creators(n HTMLNode) []models.Creator {
	return c.creatorParser.parse(n)
}

func (c *comicReleasesExtractor) ReleaseDate(s string) time.Time {
	if c.reReleaseDate == nil {
		c.logger.Warn("release date regex is nil")
		c.stats.ErrorCount.Add(1)
		return time.Time{}
	}

	d := c.reReleaseDate.FindString(s)
	if d == "" {
		return time.Time{}
	}

	for _, layout := range []string{"1/2/06", "1/2/2006"} {
		t, err := time.Parse(layout, d)
		if err == nil {
			return t
		}
	}

	c.logger.Warn("failed to parse release date", "string", s)
	c.stats.ErrorCount.Add(1)
	return time.Time{}
}

type HTMLNode interface {
	Each(func(HTMLNode))
	Text() string
	NodeName() string
}

type creatorParser struct {
	roles []string
}

func newCreatorParser(roles []string) *creatorParser {
	return &creatorParser{roles}
}

func (c *creatorParser) parse(n HTMLNode) []models.Creator {
	results := make([]models.Creator, 0)

	if n == nil {
		return results
	}

	n.Each(func(s HTMLNode) {
		if s.NodeName() == "br" {
			return
		}

		for _, role := range c.roles {
			v := strings.ToLower(strings.TrimSpace(s.Text()))

			if !strings.HasPrefix(v, role) {
				continue
			}

			split := strings.Split(v, ":")
			names := strings.Split(split[1], ",")
			for _, name := range names {
				namesNoAnd := strings.ReplaceAll(name, "and", "")
				NamesNoAmpersand := strings.ReplaceAll(namesNoAnd, "&", "")
				namesNoSpace := strings.TrimSpace(NamesNoAmpersand)
				nameFinal := cases.Title(language.English).String(namesNoSpace)
				results = append(results, models.Creator{Name: nameFinal, Role: role})

			}
		}
	})

	return results
}

func generateUrlRegex(months []string, publishers []string) string {
	m := strings.Join(months, "|")
	p := strings.Join(publishers, "|")
	y := time.Now().Year()
	return fmt.Sprintf("(?i)(%s)-(%s)-(%d|%d)-solicitations", p, m, y, y+1)
}
