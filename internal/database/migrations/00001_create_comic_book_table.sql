-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS comic_books (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    issue TEXT,
    pages TEXT,
    format TEXT,
    price TEXT,
    publisher TEXT,
    release_date DATETIME,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS creators (
    id TEXT PRIMARY KEY,
    comic_book_id TEXT NOT NULL,
    role TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME,
    FOREIGN KEY (comic_book_id) REFERENCES comic_books(id) ON DELETE CASCADE
);

CREATE INDEX idx_comics_publisher ON comic_books(publisher);

CREATE UNIQUE INDEX IF NOT EXISTS idx_comics_unique
    ON comic_books(title, issue, publisher, release_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE creators;
DROP TABLE comic_books;
-- +goose StatementEnd
