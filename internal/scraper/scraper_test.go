package scraper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

var location = `
?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl"?>
<urlset>
    <url>
        <loc>%s/dc-march-2026-solicitations/</loc>
        <lastmod>2025-12-19T18:47:21+00:00</lastmod>
    </url>
<url>
        <loc>%s/marvel-march-2026-solicitations/</loc>
        <lastmod>2025-12-19T18:47:21+00:00</lastmod>
    </url>
</urlset>`

var batmanHtml = `
<div class="wp-block-columns is-layout-flex wp-container-core-columns-is-layout-28f84493 wp-block-columns-is-layout-flex">
	<div class="wp-block-column is-layout-flow wp-block-column-is-layout-flow" style="flex-basis:66.66%">
		<p>BATMAN #1</p>
		<p>
			Writer(s): MATT FRACTION<br>
			Artist(s): JORGE JIMENEZ<br>
			Cover Artist(s): JORGE JIMENEZ<br>
			Variant Covers:<br>
			Variant covers by DUSTIN NGUYEN, JORGE MOLINA, and RYAN SOOK<br>
			Foil variant cover by JORGE JIMENEZ<br>
			1:25 variant cover by DAVID AJA<br>
			Corner box variant by JORGE JIMENEZ<br>
			Women’s History Month variant cover by LEIRIX<br>
			Symbol variant cover<br>$4.99 US | 40 pages | Variant $5.99 US (card stock) | Variant $7.99 US (foil)
		</p>
		<p>On Sale: 3/4/26</p>
		<p>As Batman is beckoned to Arkham Towers by the mysterious man in Room Ten, nothing will prepare him for who he finds there. Some might call him the Caped Crusader’s archnemesis. Others might call him Batman’s best friend. Everyone calls him the Joker. </p>
	</div>
</div>`

var marvelDateHtml = `
<div>
	<p>
		FOC 01/26/26,<strong>ON-SALE 03/11/26</strong>
	</p>
	<ul class="wp-block-list">
		<li>BATMAN #1</li>
	</ul>
</div>`

var marvelDateHtmlVariant = `
<div>
	<p>
		FOC 01/26/26,<strong>ON-SALE 03/11/26</strong>
	</p>
	<ul class="wp-block-list">
		<li>BATMAN #1 (ON SALE)</li>
	</ul>
</div>`

func setupTestServer(html string, t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, html); err != nil {
			t.Fatal(err)
		}
	}))
}

func setupTestServerXml(xml string, t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		if _, err := fmt.Fprintln(w, xml); err != nil {
			t.Fatal(err)
		}
	}))
}

func setupDefaultScraper(ex ComicBookExtractor, t *testing.T) *comicReleasesScraper {
	t.Helper()

	q, _ := queue.New(1, &queue.InMemoryQueueStorage{MaxSize: 10_000})

	return &comicReleasesScraper{
		navCol: colly.NewCollector(),
		solCol: colly.NewCollector(),
		queue:  q,
		ex:     ex,
	}
}

type MockExtractor struct {
	mock.Mock
}

func (m *MockExtractor) MatchURL(ctx context.Context, s string, obs models.ErrorObserver) bool {
	args := m.Called(ctx, s, obs)
	return args.Bool(0)
}

func (m *MockExtractor) SetUrlMatcher(months []string, pubs []string) {
	m.Called(months, pubs)
}

func (m *MockExtractor) Title(ctx context.Context, s string, obs models.ErrorObserver) string {
	args := m.Called(ctx, s, obs)
	return args.String(0)
}

func (m *MockExtractor) Issue(s string) string {
	args := m.Called(s)
	return args.String(0)
}

func (m *MockExtractor) Pages(ctx context.Context, s string, obs models.ErrorObserver) string {
	args := m.Called(ctx, s, obs)
	return args.String(0)
}

func (m *MockExtractor) Price(ctx context.Context, s string, obs models.ErrorObserver) string {
	args := m.Called(ctx, s, obs)
	return args.String(0)
}

func (m *MockExtractor) Publisher(ctx context.Context, s string, obs models.ErrorObserver) string {
	args := m.Called(ctx, s, obs)
	return args.String(0)
}

func (m *MockExtractor) Creators(node HTMLNode) []models.Creator {
	args := m.Called(node)
	return args.Get(0).([]models.Creator)
}

func (m *MockExtractor) ReleaseDate(ctx context.Context, s string, obs models.ErrorObserver) time.Time {
	args := m.Called(ctx, s, obs)
	return args.Get(0).(time.Time)
}

type mockObserver struct {
	mock.Mock
}

func (m *mockObserver) OnError(ctx context.Context, level slog.Level, msg string, args ...any) {
	m.Called(ctx, level, msg, args)
}

func (m *mockObserver) OnUrlFound(n int) {
	m.Called(n)
}

func (m *mockObserver) OnNavigationComplete() {
	m.Called()
}

func (m *mockObserver) OnComicBookScraped(n int) {
	m.Called(n)
}

func (m *mockObserver) OnScrapingComplete() {
	m.Called()
}

type MockQueryNode struct {
	mock.Mock
}

func (m *MockQueryNode) Each(f func(HTMLNode)) {
	m.Called()
}

func (m *MockQueryNode) Text() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockQueryNode) NodeName() string {
	args := m.Called()
	return args.String(0)
}

func Test_comicReleasesScraper_SetInputs(t *testing.T) {
	mockEx := new(MockExtractor)

	mockEx.On("SetUrlMatcher", mock.Anything, mock.Anything).Once()

	s := &comicReleasesScraper{ex: mockEx}
	if err := s.SetInputs([]string{"1", "2"}, []string{"3"}); err != nil {
		t.Errorf("error setting inputs: %v", err)
	}

	mockEx.AssertExpectations(t)

	s.ex = nil

	if err := s.SetInputs([]string{"1", "2"}, []string{"3"}); err == nil {
		t.Errorf("error setting inputs: expected error")
	}
}

func Test_comicReleasesScraper_bindCallbacks(t *testing.T) {
	mockObs := new(mockObserver)

	navCol := colly.NewCollector()
	solCol := colly.NewCollector()

	s := &comicReleasesScraper{observer: mockObs, navCol: navCol, solCol: solCol}

	s.bindCallbacks(context.Background())

	mockObs.AssertExpectations(t)
}

func Test_comicReleasesScraper_parseComicBookBatman(t *testing.T) {
	resp := &colly.Response{
		Request: &colly.Request{
			Ctx: &colly.Context{},
			URL: &url.URL{},
		},
		Ctx: &colly.Context{},
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(batmanHtml)))
	sel := doc.Find("div.wp-block-columns")
	el := colly.NewHTMLElementFromSelectionNode(resp, sel, sel.Nodes[0], 0)

	mockObs := new(mockObserver)
	mockObs.On("OnError", nil, slog.LevelWarn, mock.Anything, []interface{}(nil)).Maybe()

	mockEx := new(MockExtractor)

	ctx := context.Background()
	mockEx.On("Publisher", ctx, mock.Anything, mockObs).Once().Return("")
	mockEx.On("Title", ctx, mock.Anything, mockObs).Return("Batman")
	mockEx.On("Issue", mock.Anything).Return("7")
	mockEx.On("Pages", ctx, mock.Anything, mockObs).Return("32")
	mockEx.On("Price", ctx, mock.Anything, mockObs).Return("5.99")
	mockEx.On("Creators", mock.Anything).Return([]models.Creator{})
	mockEx.On("ReleaseDate", ctx, mock.Anything, mockObs).Return(time.Now())

	s := &comicReleasesScraper{ex: mockEx, observer: mockObs}

	s.parseComicBook(context.Background(), el)

	mockEx.AssertExpectations(t)
	mockObs.AssertExpectations(t)
}

func Test_comicReleasesScraper_parseComicBookMarvel(t *testing.T) {
	resp := &colly.Response{
		Request: &colly.Request{
			Ctx: &colly.Context{},
			URL: &url.URL{},
		},
		Ctx: &colly.Context{},
	}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(batmanHtml + marvelDateHtml + marvelDateHtml + marvelDateHtmlVariant)))
	sel := doc.Find("div.wp-block-columns")
	el := colly.NewHTMLElementFromSelectionNode(resp, sel, sel.Nodes[0], 0)

	mockObs := new(mockObserver)
	mockObs.On("OnError", nil, slog.LevelWarn, mock.Anything, []interface{}(nil)).Maybe()

	mockEx := new(MockExtractor)

	ctx := context.Background()
	mockEx.On("Publisher", ctx, mock.Anything, mockObs).Once().Return("")
	mockEx.On("Title", ctx, mock.Anything, mockObs).Return("")
	mockEx.On("Issue", mock.Anything).Return("")
	mockEx.On("Pages", ctx, mock.Anything, mockObs).Return("")
	mockEx.On("Price", ctx, mock.Anything, mockObs).Return("")
	mockEx.On("Creators", mock.Anything).Return([]models.Creator{})
	mockEx.On("ReleaseDate", ctx, mock.Anything, mockObs).Return(time.Time{})

	s := &comicReleasesScraper{ex: mockEx, observer: mockObs}

	s.parseComicBook(context.Background(), el)

	mockEx.AssertExpectations(t)
	mockObs.AssertExpectations(t)
}

func TestNewComicReleasesScraper(t *testing.T) {
	type args struct {
		cfg *SConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil config",
			args:    args{},
			wantErr: true,
		},
		{
			name: "valid config",
			args: args{
				cfg: &SConfig{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewComicReleasesScraper(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewComicReleasesScraper() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == true {
				return
			}
			if _, ok := got.(*comicReleasesScraper); !ok {
				t.Errorf("wanted comicReleasesScraper, got %v", reflect.TypeOf(got))
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	type args struct {
		domain      string
		parallelism int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil == error",
			args:    args{},
			wantErr: true,
		},
		{
			name: "default collector",
			args: args{
				domain:      "example.com",
				parallelism: 2,
			},
			wantErr: false,
		},
		{
			name: "invalid domain",
			args: args{
				domain:      "example%com",
				parallelism: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCollector(tt.args.domain, tt.args.parallelism)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == true {
				return
			}
			if got == nil {
				t.Errorf("wanted colly.Collector, got %v", got)
			}
		})
	}
}

func Test_comicReleasesScraper_GetDataStarts(t *testing.T) {
	ts := setupTestServer(`<html><body><a href="/comic/1">Link</a></body></html>`, t)
	defer ts.Close()

	ex := NewComicReleasesExtractor(nil)
	scraper := setupDefaultScraper(ex, t)
	results := make(chan models.ComicBook, 10)
	ctx := context.Background()
	obs := &mockObserver{}

	obs.On("OnNavigationComplete").Once()

	err := scraper.GetData(ctx, ts.URL, results, obs)

	if err != nil {
		t.Errorf("GetData failed: %v", err)
	}
}

func Test_comicReleasesScraper_GetDataAbortsWhenCancelledContext(t *testing.T) {
	ts := setupTestServer(`<html><body><a href="/comic/1">Link</a></body></html>`, t)
	defer ts.Close()

	ex := NewComicReleasesExtractor(nil)
	scraper := setupDefaultScraper(ex, t)
	obs := &mockObserver{}
	results := make(chan models.ComicBook, 10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	obs.On("OnNavigationComplete").Once()

	err := scraper.GetData(ctx, ts.URL, results, obs)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

func Test_comicReleasesScraper_GetDataHandlesRequestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	ex := NewComicReleasesExtractor(nil)
	scraper := setupDefaultScraper(ex, t)
	obs := &mockObserver{}
	results := make(chan models.ComicBook, 10)
	ctx := context.Background()

	input := []interface{}{"url", ts.URL + "/", "status", "500", "error", "Internal Server Error"}

	obs.On("OnError", ctx, slog.LevelError, "request failed", input).Once()

	if err := scraper.GetData(ctx, ts.URL, results, obs); err == nil {
		t.Errorf("expected error, got nil")
	} else {
		var httpErr interface{ StatusCode() int }
		if errors.As(err, &httpErr) && httpErr.StatusCode() != http.StatusInternalServerError {
			t.Errorf("expected error 500, got %v", err)
		}
	}
}

func Test_comicReleasesScraper_GetDataScrapesComicBook(t *testing.T) {
	tsCb := setupTestServer(batmanHtml, t)
	defer tsCb.Close()

	tsLoc := setupTestServerXml(fmt.Sprintf(location, tsCb.URL, tsCb.URL), t)
	defer tsLoc.Close()

	ex := NewComicReleasesExtractor(nil)
	obs := &mockObserver{}
	results := make(chan models.ComicBook, 10)
	ctx := context.Background()

	scraper := setupDefaultScraper(ex, t)
	if err := scraper.SetInputs([]string{"march"}, []string{"dc"}); err != nil {
		t.Errorf("SetInputs failed: %v", err)
	}

	obs.On("OnUrlFound", 1).Once()
	obs.On("OnNavigationComplete").Once()
	obs.On("OnComicBookScraped", 1).Once()
	obs.On("OnScrapingComplete").Once()

	err := scraper.GetData(ctx, tsLoc.URL, results, obs)

	if err != nil {
		t.Errorf("GetData failed: %v", err)
	}
}
