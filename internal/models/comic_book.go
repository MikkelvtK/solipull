package models

import (
	"time"
)

type ComicBook struct {
	Title       string
	Issue       int
	price       float32
	Writers     []string
	Artists     []string
	Publisher   string
	ReleaseDate time.Time
}
