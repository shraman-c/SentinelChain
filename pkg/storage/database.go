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
		device_id TEXT NOT NULL DEFAULT '',
		device_name TEXT NOT NULL DEFAULT '',
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

	if err := db.ensureColumn("blocks", "device_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := db.ensureColumn("blocks", "device_name", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	return nil
}

func (db *DB) ensureColumn(tableName, columnName, columnDefinition string) error {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return fmt.Errorf("failed to inspect table %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("failed to read table info for %s: %w", tableName, err)
		}
		if name == columnName {
			return nil
		}
	}

	if _, err := db.conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDefinition)); err != nil {
		return fmt.Errorf("failed to add column %s.%s: %w", tableName, columnName, err)
	}

	return nil
}
