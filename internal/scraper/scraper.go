package scraper

import "github.com/MikkelvtK/pul/internal/models"

type Scraper interface {
	Scrape(month string, publishers []string) ([]models.ComicBook, error)
}
