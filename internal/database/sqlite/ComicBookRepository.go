package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MikkelvtK/solipull/internal/models"
)

type ComicBookRepository struct {
	db *sql.DB
}

func NewComicBookRepository(db *sql.DB) *ComicBookRepository {
	return &ComicBookRepository{db}
}

func (c *ComicBookRepository) BulkSave(ctx context.Context, records []models.ComicBook) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT OR IGNORE INTO comic_books(title, issue, pages, format, price, publisher, release_date, created_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?);
    `)
	if err != nil {
		fmt.Println("Error in bulk save comic_books: ", err.Error())
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.ExecContext(ctx,
			r.Title, r.Issue, r.Pages, r.Format, r.Price, r.Publisher, r.ReleaseDate, time.Now()); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (c *ComicBookRepository) GetById(ctx context.Context, id int) (*models.ComicBook, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	stmt := `SELECT title, issue FROM comic_books WHERE id = ?;`

	row := tx.QueryRowContext(ctx, stmt, id)
	cb := &models.ComicBook{}
	if err := row.Scan(&cb.Title, &cb.Issue); err != nil {
		return nil, err
	}

	return cb, nil
}
