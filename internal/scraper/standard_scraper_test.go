package scraper

import (
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"reflect"
	"testing"
)

func TestStandardScraper_Scrape(t *testing.T) {
	type fields struct {
		scraper *colly.Collector
		urls    []string
		results chan models.ComicBook
		errs    chan error
	}
	tests := []struct {
		name    string
		fields  fields
		want    []models.ComicBook
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StandardScraper{
				scraper: tt.fields.scraper,
				urls:    tt.fields.urls,
				results: tt.fields.results,
				errs:    tt.fields.errs,
			}
			got, err := s.Scrape()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scrape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scrape() got = %v, want %v", got, tt.want)
			}
		})
	}
}
