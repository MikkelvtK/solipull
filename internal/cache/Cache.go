package cache

import (
    "fmt"
    "github.com/MikkelvtK/solipull/internal/models"
    "sync"
)

type Cache struct {
    cache  map[string]map[string]models.ComicBook
    mu     sync.Mutex
    length int
}

func NewCache() *Cache {
    return &Cache{
        cache:  make(map[string]map[string]models.ComicBook),
        mu:     sync.Mutex{},
        length: 0,
    }
}

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
