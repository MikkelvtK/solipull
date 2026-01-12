package models

import (
	"context"
)

type DataProvider interface {
	GetData(ctx context.Context, url string, results chan<- ComicBook) error
	ErrNum() int
}
