package storage

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenDB(baseDir string) (*sql.DB, error) {
	dbPath := filepath.Join(baseDir, "pomodoro.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	pragmaQueries := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
		`PRAGMA cache_size=-2000`,
		`PRAGMA temp_store=MEMORY`,
		`PRAGMA mmap_size=268435456`,
	}
	for _, pragma := range pragmaQueries {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, err
		}
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mode TEXT NOT NULL,
			planned_seconds INTEGER NOT NULL,
			actual_seconds INTEGER NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			completed INTEGER NOT NULL DEFAULT 0
		);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
