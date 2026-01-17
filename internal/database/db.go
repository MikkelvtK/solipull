package database

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/pressly/goose/v3"
	"io"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrations embed.FS

func MustOpen(path, driver string) *sql.DB {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(fmt.Sprintf("Error opening db: %s", err.Error()))
	}

	db, err := sql.Open(driver, path)
	if err != nil {
		panic(fmt.Sprintf("Error opening db: %s", err.Error()))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("Error opening db: %s", err.Error()))
	}

	// TODO: eventually redirect this to custom logger
	discardLogger := log.New(io.Discard, "", 0)
	goose.SetLogger(discardLogger)

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect(driver); err != nil {
		panic(fmt.Sprintf("Error opening db: %s", err.Error()))
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(fmt.Sprintf("Error opening db: %s", err.Error()))
	}

	return db
}
