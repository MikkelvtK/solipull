package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/MikkelvtK/solipull/internal/database"
	"github.com/MikkelvtK/solipull/internal/models"
	"os"
	"slices"
	"testing"
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

//func TestComicBookRepository_BulkSave(t *testing.T) {
//    type fields struct {
//        db *sql.DB
//    }
//    type args struct {
//        ctx     context.Context
//        records []models.ComicBook
//    }
//    tests := []struct {
//        name    string
//        fields  fields
//        args    args
//        wantErr bool
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            c := &ComicBookRepository{
//                db: tt.fields.db,
//            }
//            if err := c.BulkSave(tt.args.ctx, tt.args.records); (err != nil) != tt.wantErr {
//                t.Errorf("BulkSave() error = %v, wantErr %v", err, tt.wantErr)
//            }
//        })
//    }
//}

func TestComicBookRepository_GetAll(t *testing.T) {
	db, teardown := setupDB(t)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing db: %s", err.Error())
		}

		teardown()
	})

	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx context.Context
		cbs []models.ComicBook
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantLen      int
		wantErr      bool
		wantCreators bool
	}{
		{
			name: "returns empty list",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				cbs: nil,
			},
			wantLen:      0,
			wantErr:      false,
			wantCreators: false,
		},
		{
			name: "returns non empty list",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				cbs: createRandomEntries(10, false, t),
			},
			wantLen:      10,
			wantErr:      false,
			wantCreators: false,
		},
		{
			name: "returns non empty list with creators",
			fields: fields{
				db: db,
			},
			args: args{
				ctx: context.Background(),
				cbs: createRandomEntries(10, true, t),
			},
			wantLen:      10,
			wantErr:      false,
			wantCreators: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComicBookRepository{
				db: tt.fields.db,
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
			if tt.wantCreators {
				creators := slices.Collect(func(yield func(creator []models.Creator) bool) {
					for _, cb := range got {
						if !yield(cb.Creators) {
							return
						}
					}
				})

				if len(creators) != tt.wantLen {
					t.Errorf("GetAll() gotCreators = %v, wantLen %v", len(got), tt.wantLen)
				}
			}
		})
	}
}

//func TestComicBookRepository_toComicBookEntity(t *testing.T) {
//    type fields struct {
//        db *sql.DB
//    }
//    type args struct {
//        cb models.ComicBook
//    }
//    tests := []struct {
//        name   string
//        fields fields
//        args   args
//        want   comicBookEntity
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            c := &ComicBookRepository{
//                db: tt.fields.db,
//            }
//            if got := c.toComicBookEntity(tt.args.cb); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("toComicBookEntity() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func TestComicBookRepository_toCreatorEntity(t *testing.T) {
//    type fields struct {
//        db *sql.DB
//    }
//    type args struct {
//        cbUUID  string
//        creator models.Creator
//    }
//    tests := []struct {
//        name   string
//        fields fields
//        args   args
//        want   creatorEntity
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            c := &ComicBookRepository{
//                db: tt.fields.db,
//            }
//            if got := c.toCreatorEntity(tt.args.cbUUID, tt.args.creator); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("toCreatorEntity() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
//
//func TestNewComicBookRepository(t *testing.T) {
//    type args struct {
//        db *sql.DB
//    }
//    tests := []struct {
//        name string
//        args args
//        want *ComicBookRepository
//    }{
//        // TODO: Add test cases.
//    }
//    for _, tt := range tests {
//        t.Run(tt.name, func(t *testing.T) {
//            if got := NewComicBookRepository(tt.args.db); !reflect.DeepEqual(got, tt.want) {
//                t.Errorf("NewComicBookRepository() = %v, want %v", got, tt.want)
//            }
//        })
//    }
//}
