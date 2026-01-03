package scraper

import "github.com/MikkelvtK/pul/internal/models"

type Scraper interface {
	Init() error
	Scrape(month string, publishers []string) ([]models.ComicBook, error)
}
