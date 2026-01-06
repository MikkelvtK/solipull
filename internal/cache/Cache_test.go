package cache

import (
    "github.com/MikkelvtK/solipull/internal/models"
    "reflect"
    "sync"
    "testing"
)

func defaultComicBook() models.ComicBook {
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

func setupCache(t *testing.T) *Cache {
    t.Helper()

    c := NewCache()
    cb := defaultComicBook()

    if err := c.Put(cb); err != nil {
        t.Errorf("Error putting cache comic book: %v", err)
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
            c:       setupCache(t),
            want:    []models.ComicBook{defaultComicBook()},
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
            args:    args{cb: defaultComicBook()},
            want:    []models.ComicBook{defaultComicBook()},
            wantErr: false,
        },
        {
            name: "valid map with non existing key returned",
            c:    setupCache(t),
            args: args{cb: models.ComicBook{
                Title:     "Batman",
                Issue:     "2",
                Pages:     "32",
                Format:    "comic",
                Price:     "4.99",
                Publisher: "dc",
                Code:      "0123456789",
            }},
            want: []models.ComicBook{defaultComicBook(), {
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
            name:    "valid map with existing key returned",
            c:       setupCache(t),
            args:    args{cb: defaultComicBook()},
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
                cb := defaultComicBook()
                m := make(map[string]map[string]models.ComicBook)
                m[cb.Publisher] = make(map[string]models.ComicBook)
                m[cb.Publisher][cb.ID()] = cb
                return m
            }(), mu: sync.Mutex{}},
            funcToExec: func(t *testing.T, cache *Cache) {
                t.Helper()

                if err := cache.Put(defaultComicBook()); err != nil {
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
    type fields struct {
        cache  map[string]map[string]models.ComicBook
        mu     sync.Mutex
        length int
    }
    type args struct {
        publisher string
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    []models.ComicBook
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := &Cache{
                cache:  tt.fields.cache,
                mu:     tt.fields.mu,
                length: tt.fields.length,
            }
            got, err := c.GetByPublisher(tt.args.publisher)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetByPublisher() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GetByPublisher() got = %v, want %v", got, tt.want)
            }
        })
    }
}
