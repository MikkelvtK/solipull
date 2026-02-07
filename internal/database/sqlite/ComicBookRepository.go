package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"slices"
	"time"

	"github.com/MikkelvtK/solipull/internal/models"
)

type comicBookEntity struct {
	id        string
	createdAt time.Time
	models.ComicBook
}

type creatorEntity struct {
	id          string
	comicBookId string
	createdAt   time.Time
	models.Creator
}

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
	defer tx.Rollback()

	comicStmt := `
        INSERT INTO comic_books(id, title, issue, pages, format, price, publisher, release_date, created_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(title, issue, publisher, release_date)
        DO UPDATE SET title=excluded.title
        RETURNING id;`

	creatorStmt := `
        INSERT INTO creators(id, comic_book_id, role, name, created_at)
        VALUES (?, ?, ?, ?, ?);`

	for _, r := range records {
		e := c.toComicBookEntity(r)
		var dbID string

		err := tx.QueryRowContext(ctx, comicStmt,
			e.id, e.Title, e.Issue, e.Pages, e.Format, e.Price, e.Publisher, e.ReleaseDate, e.createdAt).Scan(&dbID)
		if err != nil {
			return fmt.Errorf("failed to store comic book: %v", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM creators WHERE comic_book_id = ?", dbID)
		if err != nil {
			return fmt.Errorf("failed to delete creators: %v", err)
		}

		for _, creator := range r.Creators {
			ce := c.toCreatorEntity(dbID, creator)

			if _, err := tx.ExecContext(ctx, creatorStmt, ce.id, ce.comicBookId, ce.Role, ce.Name, ce.createdAt); err != nil {
				return fmt.Errorf("failed to store creators: %v", err)
			}
		}
	}

	return tx.Commit()
}

func (c *ComicBookRepository) GetAll(ctx context.Context) ([]models.ComicBook, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt := `SELECT cb.id, cb.title, cb.issue, cb.pages, cb.format, cb.price, cb.publisher, cb.release_date,
            cr.role, cr.name
        FROM comic_books AS cb
        LEFT JOIN creators AS cr
        ON cb.id = cr.comic_book_id;`

	rows, err := tx.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comic books: %v", err)
	}
	defer rows.Close()

	cbs := make(map[string]*comicBookEntity)

	for rows.Next() {
		var cb comicBookEntity
		var role, name sql.NullString
		err := rows.Scan(&cb.id, &cb.Title, &cb.Issue, &cb.Pages, &cb.Format, &cb.Price, &cb.Publisher,
			&cb.ReleaseDate, &role, &name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve comic books: %v", err)
		}

		if _, ok := cbs[cb.id]; !ok {
			cbs[cb.id] = &cb
		}

		if role.Valid || name.Valid {
			cbs[cb.id].Creators = append(cbs[cb.id].Creators, models.Creator{Name: name.String, Role: role.String})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read comic books: %v", err)
	}

	return slices.Collect(func(yield func(book models.ComicBook) bool) {
		for _, v := range cbs {
			if !yield(v.ComicBook) {
				return
			}
		}
	}), nil
}

func (c *ComicBookRepository) toComicBookEntity(cb models.ComicBook) comicBookEntity {
	return comicBookEntity{
		id:        uuid.New().String(),
		createdAt: time.Now(),
		ComicBook: cb,
	}
}

func (c *ComicBookRepository) toCreatorEntity(cbUUID string, creator models.Creator) creatorEntity {
	return creatorEntity{
		id:          uuid.New().String(),
		comicBookId: cbUUID,
		createdAt:   time.Now(),
		Creator:     creator,
	}
}
