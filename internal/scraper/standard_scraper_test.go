package scraper

import (
	"github.com/MikkelvtK/pul/internal/models"
	"github.com/gocolly/colly/v2"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("no-content-type") != "" {
			w.Header()["Content-Type"] = nil
		} else {
			w.Header().Set("Content-Type", "text/html")
		}
		w.WriteHeader(200)
		_, err := w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
<div>Batman</div>
</body>
</html>`))

		if err != nil {
			t.Fatal(err)
		}
	})

	return httptest.NewUnstartedServer(mux)
}

func setupTestSuite(t *testing.T) (*httptest.Server, func()) {
	s := setupTestServer(t)
	s.Start()
	return s, func() {
		s.Close()
	}
}

func TestStandardScraper_CollectorVisit(t *testing.T) {
	t.Run("run collector visit", func(t *testing.T) {
		ts, teardown := setupTestSuite(t)
		defer teardown()

		u := []string{ts.URL + "/"}
		r := make(chan models.ComicBook, 50)
		c := colly.NewCollector()

		s := &StandardScraper{
			scraper: c,
			urls:    u,
			results: r,
		}

		onHtmlReached := false

		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode != 200 {
				t.Error("OnResponse", r.StatusCode)
			}
		})

		c.OnHTML("div", func(e *colly.HTMLElement) {
			onHtmlReached = true

			cb := models.ComicBook{
				Title: e.Text,
			}

			s.results <- cb
		})

		c.OnError(func(_ *colly.Response, err error) {
			t.Fatal(err)
		})

		got, err := s.Scrape()
		if err != nil {
			t.Errorf("Scrape() error = %v, wantErr %v", err, false)
			return
		}

		if !onHtmlReached {
			t.Errorf("OnHtml was not reached")
		}

		want := []models.ComicBook{{
			Title: "Batman",
		}}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Scrape() got = %v, want %v", got, want)
		}
	})
}

func TestStandardScraper_NonExistentURL(t *testing.T) {
	t.Run("run collector visit on a non existent url", func(t *testing.T) {
		c := colly.NewCollector()
		r := make(chan models.ComicBook, 50)
		u := []string{"nonexistent"}

		s := &StandardScraper{
			scraper: c,
			results: r,
			urls:    u,
		}

		_, err := s.Scrape()
		if err == nil {
			t.Error("Scrape() error = nil, want error")
		}
	})
}
