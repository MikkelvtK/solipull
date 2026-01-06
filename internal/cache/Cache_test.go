package cache

import (
	"github.com/MikkelvtK/solipull/internal/models"
	"reflect"
	"sync"
	"testing"
)

func defaultBatmanComicBook() models.ComicBook {
	return models.ComicBook{
		Title:     "Batman",
		Issue:     "1",
		Pages:     "32",
		Format:    "comic",
		Price:     "4.99",
		Publisher: "dc",
		Code:      "0123456789",
	}
}

func defaultSupermanComicBook() models.ComicBook {
	return models.ComicBook{
		Title:     "Superman",
		Issue:     "1",
		Pages:     "32",
		Format:    "comic",
		Price:     "4.99",
		Publisher: "dc",
		Code:      "0123456789",
	}
}

func setupCache(t *testing.T, comics []models.ComicBook) *Cache {
	t.Helper()

	c := NewCache()
	for _, comic := range comics {
		if err := c.Put(comic); err != nil {
			t.Errorf("Error putting comic: %v", err)
		}
	}

	return c
}

func setupNilCache(t *testing.T) *Cache {
	t.Helper()

	return &Cache{}
}

func TestCache_GetAll(t *testing.T) {
	type testCase struct {
		name    string
		c       *Cache
		want    []models.ComicBook
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil map == no crash",
			c:       setupNilCache(t),
			wantErr: true,
		},
		{
			name:    "valid map returned",
			c:       setupCache(t, []models.ComicBook{defaultBatmanComicBook()}),
			want:    []models.ComicBook{defaultBatmanComicBook()},
			wantErr: false,
		},
		{
			name:    "cache has correct length",
			c:       setupCache(t, []models.ComicBook{defaultBatmanComicBook(), defaultSupermanComicBook()}),
			want:    []models.ComicBook{defaultBatmanComicBook(), defaultSupermanComicBook()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
			if len(got) != len(tt.want) {
				t.Errorf("len(GetAll()) = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}

func TestCache_Put(t *testing.T) {
	type args struct {
		cb models.ComicBook
	}
	type testCase struct {
		name    string
		c       *Cache
		args    args
		want    []models.ComicBook
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "nil map still returns value",
			c:       setupNilCache(t),
			args:    args{cb: defaultBatmanComicBook()},
			want:    []models.ComicBook{defaultBatmanComicBook()},
			wantErr: false,
		},
		{
			name: "valid map with non existing key returned",
			c:    setupCache(t, []models.ComicBook{defaultBatmanComicBook()}),
			args: args{cb: models.ComicBook{
				Title:     "Batman",
				Issue:     "2",
				Pages:     "32",
				Format:    "comic",
				Price:     "4.99",
				Publisher: "dc",
				Code:      "0123456789",
			}},
			want: []models.ComicBook{defaultBatmanComicBook(), {
				Title:     "Batman",
				Issue:     "2",
				Pages:     "32",
				Format:    "comic",
				Price:     "4.99",
				Publisher: "dc",
				Code:      "0123456789",
			}},
			wantErr: false,
		},
		{
			name:    "duplicates not allowed check",
			c:       setupCache(t, []models.ComicBook{defaultBatmanComicBook()}),
			args:    args{cb: defaultBatmanComicBook()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Put(tt.args.cb)
			if (err != nil) == tt.wantErr {
				return
			}

			got, err := tt.c.GetByPublisher(tt.args.cb.Publisher)
			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCache(t *testing.T) {
	tests := []struct {
		name       string
		want       *Cache
		funcToExec func(t *testing.T, cache *Cache)
	}{
		{
			name: "New cache is successfully created",
			want: &Cache{cache: make(map[string]map[string]models.ComicBook), mu: sync.Mutex{}},
		},
		{
			name: "New cache can be used to put data in",
			want: &Cache{cache: func() map[string]map[string]models.ComicBook {
				cb := defaultBatmanComicBook()
				m := make(map[string]map[string]models.ComicBook)
				m[cb.Publisher] = make(map[string]models.ComicBook)
				m[cb.Publisher][cb.ID()] = cb
				return m
			}(), mu: sync.Mutex{}, length: 1},
			funcToExec: func(t *testing.T, cache *Cache) {
				t.Helper()

				if err := cache.Put(defaultBatmanComicBook()); err != nil {
					t.Errorf("Error putting comic book in cache: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCache()
			if tt.funcToExec != nil {
				tt.funcToExec(t, got)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_GetByPublisher(t *testing.T) {
	type args struct {
		publisher string
	}
	tests := []struct {
		name    string
		cache   *Cache
		args    args
		want    []models.ComicBook
		wantErr bool
		len     int
	}{
		{
			name:    "publisher not found",
			cache:   NewCache(),
			args:    args{publisher: "dc"},
			want:    nil,
			wantErr: true,
			len:     0,
		},
		{
			name:    "publisher found",
			cache:   setupCache(t, []models.ComicBook{defaultBatmanComicBook()}),
			args:    args{publisher: "dc"},
			want:    []models.ComicBook{defaultBatmanComicBook()},
			wantErr: false,
			len:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cache.GetByPublisher(tt.args.publisher)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPublisher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByPublisher() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.len {
				t.Errorf("Expected length %v, got %v", tt.len, len(got))
			}
		})
	}
}

func TestCache_GetByTitle(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name    string
		cache   *Cache
		args    args
		want    []models.ComicBook
		wantErr bool
		len     int
	}{
		{
			name:    "title not found",
			cache:   NewCache(),
			args:    args{title: "Batman"},
			want:    nil,
			wantErr: true,
			len:     0,
		},
		{
			name:    "title found",
			cache:   setupCache(t, []models.ComicBook{defaultBatmanComicBook()}),
			args:    args{title: "Batman"},
			want:    []models.ComicBook{defaultBatmanComicBook()},
			wantErr: false,
			len:     1,
		},
		{
			name: "correct value found",
			cache: func() *Cache {
				c := NewCache()
				if err := c.Put(defaultBatmanComicBook()); err != nil {
					t.Errorf("Error putting comic book in cache: %v", err)
				}
				if err := c.Put(defaultSupermanComicBook()); err != nil {
					t.Errorf("Error putting comic book in cache: %v", err)
				}
				return c
			}(),
			args:    args{title: "Superman"},
			want:    []models.ComicBook{defaultSupermanComicBook()},
			wantErr: false,
			len:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cache.GetByTitle(tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByTitle() got = %v, want %v", got, tt.want)
			}
			if len(got) != tt.len {
				t.Errorf("Expected length %v, got %v", tt.len, len(got))
			}
		})
	}
}
