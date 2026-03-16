package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	ID           int64
	LogTimestamp int64
	SourceIP     string
	EventType    string
	Severity     string
	Message      string
	PrevHash     string
	Hash         string
	InsertedAt   int64
}

func (db *DB) IsEmpty() (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count blocks: %w", err)
	}
	return count == 0, nil
}

func (db *DB) GetLastBlock() (*Block, error) {
	var block Block
	err := db.conn.QueryRow(`
		SELECT id, log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at
		FROM blocks ORDER BY id DESC LIMIT 1
	`).Scan(&block.ID, &block.LogTimestamp, &block.SourceIP, &block.EventType, &block.Severity, &block.Message, &block.PrevHash, &block.Hash, &block.InsertedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last block: %w", err)
	}
	return &block, nil
}

func (db *DB) GetBlockByID(id int64) (*Block, error) {
	var block Block
	err := db.conn.QueryRow(`
		SELECT id, log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at
		FROM blocks WHERE id = ?
	`, id).Scan(&block.ID, &block.LogTimestamp, &block.SourceIP, &block.EventType, &block.Severity, &block.Message, &block.PrevHash, &block.Hash, &block.InsertedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block by id: %w", err)
	}
	return &block, nil
}

func ComputeHash(timestamp int64, sourceIP, eventType, severity, message, prevHash string) string {
	data := fmt.Sprintf("%d%s%s%s%s%s", timestamp, sourceIP, eventType, severity, message, prevHash)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (db *DB) InsertBlock(block *Block) error {
	block.InsertedAt = time.Now().UnixNano()
	block.Hash = ComputeHash(block.LogTimestamp, block.SourceIP, block.EventType, block.Severity, block.Message, block.PrevHash)

	_, err := db.conn.Exec(`
		INSERT INTO blocks (log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, block.LogTimestamp, block.SourceIP, block.EventType, block.Severity, block.Message, block.PrevHash, block.Hash, block.InsertedAt)

	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	return nil
}

func (db *DB) GetAllBlocks() ([]*Block, error) {
	rows, err := db.conn.Query(`
		SELECT id, log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at
		FROM blocks ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*Block
	for rows.Next() {
		var block Block
		if err := rows.Scan(&block.ID, &block.LogTimestamp, &block.SourceIP, &block.EventType, &block.Severity, &block.Message, &block.PrevHash, &block.Hash, &block.InsertedAt); err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}
		blocks = append(blocks, &block)
	}

	return blocks, nil
}
