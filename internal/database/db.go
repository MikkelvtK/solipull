package database

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrations embed.FS

func InitDB(path, driver string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	db, err := sql.Open(driver, path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// TODO: eventually redirect this to custom logger
	//discardLogger := log.New(io.Discard, "", 0)
	//goose.SetLogger(discardLogger)

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect(driver); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return db, nil
}
