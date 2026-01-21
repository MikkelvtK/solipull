package scraper

import (
	"context"
	"github.com/MikkelvtK/solipull/internal/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/url"
	"testing"
	"time"
)

type MockHtmlElementWrapper struct {
	mock.Mock
}

func (m *MockHtmlElementWrapper) dom() *goquery.Selection {
	args := m.Called()
	return args.Get(0).(*goquery.Selection)
}

func (m *MockHtmlElementWrapper) request() *colly.Request {
	args := m.Called()
	return args.Get(0).(*colly.Request)
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

func Test_comicReleasesScraper_parseComicBook(t *testing.T) {
	mockWr := new(MockHtmlElementWrapper)

	mockWr.On("dom").Times(3).Return(&goquery.Selection{})
	mockWr.On("request").Return(&colly.Request{URL: &url.URL{Path: "/comic/book"}})

	mockObs := new(mockObserver)
	mockObs.On("OnError", nil, slog.LevelWarn, mock.Anything, []interface{}(nil)).Maybe()

	mockEx := new(MockExtractor)

	mockEx.On("Publisher", context.Background(), mock.Anything, mockObs).Once().Return("")

	s := &comicReleasesScraper{ex: mockEx, observer: mockObs}

	s.parseComicBook(context.Background(), mockWr)

	mockWr.AssertExpectations(t)
	mockEx.AssertExpectations(t)
	mockObs.AssertExpectations(t)
}
