package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/database"
	"github.com/MikkelvtK/solipull/internal/models"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func setupDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	var path = "./test/test_db.db"

	db := database.MustOpen(path, "sqlite")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("sqlite db not created")
	}

	return db, func() {
		if err := os.RemoveAll("./test"); err != nil {
			t.Helper()

			t.Fatalf("Error removing test database: %s", err.Error())
		}

		if _, err := os.Stat(path); err == nil {
			t.Fatalf("Test database still exists")
		}
	}
}

func createRandomEntries(num int, withCreators bool, t *testing.T) []models.ComicBook {
	t.Helper()

	cbs := make([]models.ComicBook, 0, num)
	for i := 0; i < num; i++ {
		cb := models.ComicBook{
			Title: fmt.Sprintf("title-%d", i),
		}

		if withCreators {
			cb.Creators = []models.Creator{
				{
					Name: fmt.Sprintf("creator-%d", i),
				},
			}
		}

		cbs = append(cbs, cb)
	}

	return cbs
}

func TestComicBookRepository_GetAll_BulkSave(t *testing.T) {
	type args struct {
		ctx context.Context
		cbs []models.ComicBook
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name: "returns empty list",
			args: args{
				ctx: context.Background(),
				cbs: nil,
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "returns non empty list",
			args: args{
				ctx: context.Background(),
				cbs: createRandomEntries(10, false, t),
			},
			wantLen: 10,
			wantErr: false,
		},
		{
			name: "returns non empty list with creators",
			args: args{
				ctx: context.Background(),
				cbs: createRandomEntries(10, true, t),
			},
			wantLen: 10,
			wantErr: false,
		},
		{
			name: "returns all fields of the comic book",
			args: args{
				ctx: context.Background(),
				cbs: []models.ComicBook{
					{
						Title:  "title",
						Issue:  "1",
						Pages:  "32",
						Format: "single",
						Price:  "$4.99",
						Creators: []models.Creator{
							{
								Name: "creator-1",
								Role: "writer",
							},
						},
						Publisher:   "dc",
						ReleaseDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := setupDB(t)
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Errorf("Error closing db: %s", err.Error())
				}

				teardown()
			})

			c := &ComicBookRepository{
				db: db,
			}

			if tt.args.cbs != nil {
				if err := c.BulkSave(tt.args.ctx, tt.args.cbs); (err != nil) != tt.wantErr {
					t.Errorf("BulkSave() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			got, err := c.GetAll(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantLen != len(got) {
				t.Errorf("GetAll() got = %v, wantLen %v", len(got), tt.wantLen)
			}
			slices.SortFunc(got, func(a, b models.ComicBook) int {
				return strings.Compare(a.Title, b.Title)
			})
			if !reflect.DeepEqual(got, tt.args.cbs) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.args.cbs)
			}
		})
	}
}

func TestNewComicBookRepository(t *testing.T) {
	db, teardown := setupDB(t)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing db: %s", err.Error())
		}

		teardown()
	})

	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want *ComicBookRepository
	}{
		{
			name: "returns a new ComicBookRepository instance",
			args: args{
				db: db,
			},
			want: &ComicBookRepository{db: db},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewComicBookRepository(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComicBookRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}
