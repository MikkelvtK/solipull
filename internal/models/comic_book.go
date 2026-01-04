package models

import "time"

type ComicBook struct {
	Title       string
	Issue       int
	Pages       int
	Type        string
	Price       string
	Creators    map[string][]string
	Publisher   string
	ReleaseDate time.Time
}
