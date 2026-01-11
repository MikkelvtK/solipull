package creleases

import (
	"github.com/MikkelvtK/solipull/internal/scraper"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"reflect"
	"testing"
)

func setupCollector(domain string, t *testing.T) *colly.Collector {
	t.Helper()

	c, err := scraper.NewDefaultCollector(domain, nil)
	if err != nil {
		t.Errorf("%v", err.Error())
	}
	return c
}

func setupQueue(t *testing.T) *queue.Queue {
	t.Helper()

	q, err := queue.New(5, &queue.InMemoryQueueStorage{MaxSize: 10_000})
	if err != nil {
		t.Errorf("%v", err.Error())
	}
	return q
}

func TestNewComicReleasesScraper(t *testing.T) {
	type args struct {
		list   *colly.Collector
		detail *colly.Collector
		q      *queue.Queue
	}
	tests := []struct {
		name string
		args args
		want scraper.Scraper
	}{
		{
			name: "nil == no errors",
			args: args{
				list:   nil,
				detail: nil,
				q:      nil,
			},
			want: &comicReleasesScraper{url: "https://www." + Domain + "/sitemap.xml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewComicReleasesScraper(tt.args.list, tt.args.detail, tt.args.q); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComicReleasesScraper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_comicReleasesScraper_Run(t *testing.T) {
	type fields struct {
		listCollector   *colly.Collector
		detailCollector *colly.Collector
		queue           *queue.Queue
		url             string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nil == no errors",
			fields: fields{
				listCollector:   nil,
				detailCollector: nil,
				queue:           nil,
			},
			wantErr: true,
		},
		{
			name: "default test case",
			fields: fields{
				listCollector:   setupCollector("example.com", t),
				detailCollector: setupCollector("example.com", t),
				queue:           setupQueue(t),
				url:             "https://www.example.com",
			},
			wantErr: false,
		},
		{
			name: "bad url for list collector",
			fields: fields{
				listCollector:   setupCollector("example2.com", t),
				detailCollector: setupCollector("example.com", t),
				queue:           setupQueue(t),
				url:             "https://www.example.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &comicReleasesScraper{
				listCollector:   tt.fields.listCollector,
				detailCollector: tt.fields.detailCollector,
				queue:           tt.fields.queue,
				url:             tt.fields.url,
			}
			if err := s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
