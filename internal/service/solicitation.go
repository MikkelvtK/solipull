package service

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
)

const (
	Domain = "comicreleases.com"
)

type DataProvider interface {
	GetData(ctx context.Context, url string, results chan<- models.ComicBook) error
}
