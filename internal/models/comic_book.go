package models

import (
	"context"
	"time"
)

type ComicBookRepository interface {
	BulkSave(ctx context.Context, records []ComicBook) error
	GetById(ctx context.Context, id int) (*ComicBook, error)
	GetAll(ctx context.Context) ([]ComicBook, error)
}

type ComicBook struct {
	Id          int
	Title       string
	Issue       string
	Pages       string
	Format      string
	Price       string
	Creators    []Creator
	Publisher   string
	ReleaseDate time.Time
	CreatedAt   time.Time
}
