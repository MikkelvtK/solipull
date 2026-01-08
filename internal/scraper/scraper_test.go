package scraper

import (
	"github.com/gocolly/colly/v2"
	"testing"
)

func setupCollector(domain string, t *testing.T) *colly.Collector {
	t.Helper()

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(1),
		colly.AllowedDomains(domain),
	)

	c.IgnoreRobotsTxt = false
	return c
}

//func TestScraper_Run(t *testing.T) {
//    type fields struct {
//        listCollector   *colly.Collector
//        detailCollector *colly.Collector
//        queue           *queue.Queue
//        url             string
//    }
//    tests := []struct {
//        name    string
//        fields  fields
//        wantErr bool
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            s := &Scraper{
//                listCollector:   tt.fields.listCollector,
//                detailCollector: tt.fields.detailCollector,
//                queue:           tt.fields.queue,
//                url:             tt.fields.url,
//            }
//            if err := s.Run(); (err != nil) != tt.wantErr {
//                t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
//            }
//        })
//    }
//}

func Test_newDefaultCollector(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name    string
		args    args
		want    *colly.Collector
		wantErr bool
	}{
		{
			name:    "Collector is successfully created",
			args:    args{domain: "example.com"},
			want:    setupCollector("example.com", t),
			wantErr: false,
		},
		{
			name:    "Collector is not successfully created",
			args:    args{domain: ""},
			want:    setupCollector("", t),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDefaultCollector(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDefaultCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}
			if got.IgnoreRobotsTxt != tt.want.IgnoreRobotsTxt {
				t.Errorf("newDefaultCollector().IgnoreRobotsTxt  got = %v, want %v", got.IgnoreRobotsTxt, tt.want.IgnoreRobotsTxt)
			}
			if got.MaxDepth != tt.want.MaxDepth {
				t.Errorf("newDefaultCollector().MaxDepth  got = %v, want %v", got.MaxDepth, tt.want.MaxDepth)
			}
			if got.Async != tt.want.Async {
				t.Errorf("newDefaultCollector().Async  got = %v, want %v", got.Async, tt.want.Async)
			}
		})
	}
}
