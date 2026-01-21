package scraper

import (
	"context"
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
	MatchURL(context.Context, string, models.ErrorObserver) bool
	SetUrlMatcher([]string, []string)
	Title(context.Context, string, models.ErrorObserver) string
	Issue(string) string
	Pages(context.Context, string, models.ErrorObserver) string
	Price(context.Context, string, models.ErrorObserver) string
	Publisher(context.Context, string, models.ErrorObserver) string
	Creators(HTMLNode) []models.Creator
	ReleaseDate(context.Context, string, models.ErrorObserver) time.Time
}

type comicReleasesExtractor struct {
	reUrl         *regexp.Regexp
	rePublisher   *regexp.Regexp
	rePages       *regexp.Regexp
	rePrice       *regexp.Regexp
	reReleaseDate *regexp.Regexp
	creatorParser *creatorParser
	logger        *slog.Logger
}

func NewComicReleasesExtractor(l *slog.Logger) ComicBookExtractor {
	return &comicReleasesExtractor{
		rePublisher:   regexp.MustCompile(`(?i)/(?P<Pub>\w+)-[a-zA-Z]+-\d{4}-solicitations`),
		rePages:       regexp.MustCompile(`(?P<Pages>\d+)\s*(?i)(?:pages?|pgs?.?)`),
		rePrice:       regexp.MustCompile(`\$(\d+\.\d{2})`),
		reReleaseDate: regexp.MustCompile(`(?i)(\d{1,2}/\d{1,2}/\d{2,4})`),
		creatorParser: newCreatorParser([]string{"writer", "artist", "cover artist"}),
		logger:        l,
	}
}

func (c *comicReleasesExtractor) SetUrlMatcher(months, publishers []string) {
	c.reUrl = regexp.MustCompile(generateUrlRegex(months, publishers))
}

func (c *comicReleasesExtractor) MatchURL(ctx context.Context, url string, observer models.ErrorObserver) bool {
	if observer == nil {
		panic("nil observer")
	}

	if c.reUrl == nil {
		observer.OnError(ctx, slog.LevelWarn, "url regex compilation failed")
		return false
	}

	return c.reUrl.MatchString(url)
}

func (c *comicReleasesExtractor) Title(ctx context.Context, s string, observer models.ErrorObserver) string {
	split := strings.Split(s, "#")
	if len(split[0]) == 0 {
		observer.OnError(ctx, slog.LevelWarn, "title not found", "string", s)
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

func (c *comicReleasesExtractor) Pages(ctx context.Context, s string, observer models.ErrorObserver) string {
	if c.rePages == nil {
		observer.OnError(ctx, slog.LevelWarn, "pages regex is nil")
		return ""
	}

	matches := c.rePages.FindStringSubmatch(s)
	if matches == nil {
		observer.OnError(ctx, slog.LevelWarn, "no matches for pages found", "string", s)
		return ""
	}

	i := c.rePages.SubexpIndex("Pages")
	if i < 0 {
		observer.OnError(ctx, slog.LevelWarn, "no index for pages found", "string", s)
		return ""
	}

	return matches[i]
}

func (c *comicReleasesExtractor) Price(ctx context.Context, s string, observer models.ErrorObserver) string {
	if c.rePrice == nil {
		observer.OnError(ctx, slog.LevelWarn, "price regex is nil")
		return ""
	}
	return c.rePrice.FindString(s)
}

func (c *comicReleasesExtractor) Publisher(ctx context.Context, s string, observer models.ErrorObserver) string {
	if c.rePublisher == nil {
		observer.OnError(ctx, slog.LevelWarn, "publisher regex is nil")
		return ""
	}

	matches := c.rePublisher.FindStringSubmatch(s)
	if matches == nil {
		observer.OnError(ctx, slog.LevelWarn, "no matches for publisher found", "string", s)
		return ""
	}

	i := c.rePublisher.SubexpIndex("Pub")
	if i < 0 {
		observer.OnError(ctx, slog.LevelWarn, "no index for publisher found", "string", s)
		return ""
	}
	return strings.ToLower(matches[i])
}

func (c *comicReleasesExtractor) Creators(n HTMLNode) []models.Creator {
	return c.creatorParser.parse(n)
}

func (c *comicReleasesExtractor) ReleaseDate(ctx context.Context, s string, observer models.ErrorObserver) time.Time {
	if c.reReleaseDate == nil {
		observer.OnError(ctx, slog.LevelWarn, "release date regex is nil")
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

	observer.OnError(ctx, slog.LevelWarn, "failed to parse release date", "string", s)
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
	if len(publishers) == 0 || len(months) == 0 {
		return ""
	}

	m := strings.Join(months, "|")
	p := strings.Join(publishers, "|")
	y := time.Now().Year()
	return fmt.Sprintf("(?i)(%s)-(%s)-(%d|%d)-solicitations", p, m, y, y+1)
}
