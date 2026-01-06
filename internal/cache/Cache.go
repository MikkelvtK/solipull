package cache

import (
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"sync"
)

// Cache provides a thread safe, in-memory way to store comic book data. It's used by the scraper before
// the data cached gets exported to Google Spreadsheets.
type Cache struct {

	// cache stores the comic book data using maps. It orders the data by publisher and then ID to avoid duplicates.
	cache map[string]map[string]models.ComicBook

	// mu ensures thread safety while scraping websites.
	mu sync.Mutex

	// length tracks how many comic books have been stored in the cache.
	length int
}

// NewCache returns an empty cache.
func NewCache() *Cache {
	return &Cache{
		cache:  make(map[string]map[string]models.ComicBook),
		mu:     sync.Mutex{},
		length: 0,
	}
}

// GetByTitle returns all the comic books in the cache by their title.
func (c *Cache) GetByTitle(title string) ([]models.ComicBook, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s := make([]models.ComicBook, 0)
	for _, cbs := range c.cache {
		for _, cb := range cbs {
			if cb.Title == title {
				s = append(s, cb)
			}
		}
	}

	if len(s) == 0 {
		return nil, fmt.Errorf("no values for %v were found", title)
	}

	return s, nil
}

// GetByPublisher returns all the comic books by their publisher.
func (c *Cache) GetByPublisher(publisher string) ([]models.ComicBook, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.cache[publisher]
	if !ok {
		return nil, fmt.Errorf("no values for publisher %v", publisher)
	}

	s := make([]models.ComicBook, 0, len(v))
	for _, cb := range v {
		s = append(s, cb)
	}

	return s, nil
}

// GetAll returns all the comic books as a slice.
func (c *Cache) GetAll() ([]models.ComicBook, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil || len(c.cache) == 0 {
		return nil, fmt.Errorf("cache is empty")
	}

	s := make([]models.ComicBook, 0, c.length)
	for _, cbs := range c.cache {
		for _, cb := range cbs {
			s = append(s, cb)
		}
	}

	return s, nil
}

// Put stores a comic book in the cache. It uses the publisher and ID method to place it in the cache.
func (c *Cache) Put(comic models.ComicBook) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[string]map[string]models.ComicBook)
	}

	if _, ok := c.cache[comic.Publisher]; !ok {
		c.cache[comic.Publisher] = make(map[string]models.ComicBook)
	}

	if _, ok := c.cache[comic.Publisher][comic.ID()]; ok {
		return fmt.Errorf("value for id %v of publisher %v already exists", comic.ID(), comic.Publisher)
	}

	c.cache[comic.Publisher][comic.ID()] = comic
	c.length++
	return nil
}
