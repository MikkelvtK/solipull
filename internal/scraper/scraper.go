package scraper

import "github.com/MikkelvtK/pul/internal/models"

type Scraper interface {
	Scrape() ([]models.ComicBook, error)
}
