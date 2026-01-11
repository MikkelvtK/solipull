package creleases

import (
	"github.com/MikkelvtK/solipull/internal/cache"
	"github.com/MikkelvtK/solipull/internal/scraper"
	"reflect"
	"regexp"
	"testing"
)

func TestNewDetailParser(t *testing.T) {
	type args struct {
		c *cache.Cache
		e scraper.ComicBookExtractor
	}
	tests := []struct {
		name string
		args args
		want scraper.ParsingStrategy
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDetailParser(tt.args.c, tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDetailParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewListParser(t *testing.T) {
	type args struct {
		months     []string
		publishers []string
		q          *queue.Queue
	}
	tests := []struct {
		name string
		args args
		want scraper.ParsingStrategy
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewListParser(tt.args.months, tt.args.publishers, tt.args.q); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewListParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_detailParser_Bind(t *testing.T) {
	type fields struct {
		c *cache.Cache
		e scraper.ComicBookExtractor
	}
	type args struct {
		c *colly.Collector
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &detailParser{
				c: tt.fields.c,
				e: tt.fields.e,
			}
			p.Bind(tt.args.c)
		})
	}
}

func Test_detailParser_Parse(t *testing.T) {
	type fields struct {
		c *cache.Cache
		e scraper.ComicBookExtractor
	}
	type args struct {
		e *colly.HTMLElement
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &detailParser{
				c: tt.fields.c,
				e: tt.fields.e,
			}
			p.Parse(tt.args.e)
		})
	}
}

func Test_detailParser_Selector(t *testing.T) {
	type fields struct {
		c *cache.Cache
		e scraper.ComicBookExtractor
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &detailParser{
				c: tt.fields.c,
				e: tt.fields.e,
			}
			if got := p.Selector(); got != tt.want {
				t.Errorf("Selector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateUrlRegex(t *testing.T) {
	type args struct {
		months     []string
		publishers []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateUrlRegex(tt.args.months, tt.args.publishers); got != tt.want {
				t.Errorf("generateUrlRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listParser_Bind(t *testing.T) {
	type fields struct {
		reUrl *regexp.Regexp
		q     *queue.Queue
	}
	type args struct {
		c *colly.Collector
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &listParser{
				reUrl: tt.fields.reUrl,
				q:     tt.fields.q,
			}
			p.Bind(tt.args.c)
		})
	}
}

func Test_listParser_Parse(t *testing.T) {
	type fields struct {
		reUrl *regexp.Regexp
		q     *queue.Queue
	}
	type args struct {
		e *colly.XMLElement
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &listParser{
				reUrl: tt.fields.reUrl,
				q:     tt.fields.q,
			}
			p.Parse(tt.args.e)
		})
	}
}

func Test_listParser_Selector(t *testing.T) {
	type fields struct {
		reUrl *regexp.Regexp
		q     *queue.Queue
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &listParser{
				reUrl: tt.fields.reUrl,
				q:     tt.fields.q,
			}
			if got := p.Selector(); got != tt.want {
				t.Errorf("Selector() = %v, want %v", got, tt.want)
			}
		})
	}
}
