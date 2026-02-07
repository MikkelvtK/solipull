package models

import (
	"context"
	"time"
)

type ComicBookRepository interface {
	BulkSave(ctx context.Context, records []ComicBook) error
	GetAll(ctx context.Context) ([]ComicBook, error)
}

type ComicBook struct {
	Title       string
	Issue       string
	Pages       string
	Format      string
	Price       string
	Creators    []Creator
	Publisher   string
	ReleaseDate time.Time
}
