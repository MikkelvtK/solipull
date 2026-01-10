package scraper

import (
	"github.com/MikkelvtK/solipull/internal/models"
	"time"
)

type ComicBookExtractor interface {
	Title(string) string
	Issue(string) string
	Pages(string) string
	Price(string) string
	Publisher(string) string
	Creators(HTMLNode) []models.Creator
	ReleaseDate(string) time.Time
}
