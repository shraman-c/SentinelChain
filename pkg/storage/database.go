package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const (
	DbFile = "blockchain.db"
)

var (
	GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
)

type DB struct {
	conn *sql.DB
}

func NewDB() (*DB, error) {
	dbFile := DbFile
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS blocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		log_timestamp INTEGER NOT NULL,
		source_ip TEXT NOT NULL,
		event_type TEXT NOT NULL,
		severity TEXT NOT NULL,
		message TEXT NOT NULL,
		prev_hash TEXT NOT NULL,
		hash TEXT NOT NULL,
		inserted_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash);
	CREATE INDEX IF NOT EXISTS idx_blocks_prev_hash ON blocks(prev_hash);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
