package models

import (
	"strings"
	"time"
)

type ComicBook struct {
	Title       string
	Issue       string
	Pages       string
	Format      string
	Price       string
	Creators    map[string][]string
	Publisher   string
	ReleaseDate time.Time
}

func (c ComicBook) ID() string {
	return strings.Join([]string{c.Title, c.Issue, c.Format, c.ReleaseDate.Format("2006-01-02")}, "|")
}
