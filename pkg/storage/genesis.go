package storage

import (
	"fmt"
	"time"
)

func (db *DB) InitGenesisBlock() error {
	empty, err := db.IsEmpty()
	if err != nil {
		return fmt.Errorf("failed to check if database is empty: %w", err)
	}

	if !empty {
		return nil
	}

	genesisBlock := &Block{
		LogTimestamp: 0,
		SourceIP:     "0.0.0.0",
		EventType:    "GENESIS",
		Severity:     "INFO",
		Message:      "Genesis Block",
		PrevHash:     GenesisPrevHash,
		InsertedAt:   time.Now().UnixNano(),
	}
	genesisBlock.Hash = ComputeHash(
		genesisBlock.LogTimestamp,
		genesisBlock.SourceIP,
		genesisBlock.EventType,
		genesisBlock.Severity,
		genesisBlock.Message,
		genesisBlock.PrevHash,
	)

	_, err = db.conn.Exec(`
		INSERT INTO blocks (log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, genesisBlock.LogTimestamp, genesisBlock.SourceIP, genesisBlock.EventType, genesisBlock.Severity, genesisBlock.Message, genesisBlock.PrevHash, genesisBlock.Hash, genesisBlock.InsertedAt)

	if err != nil {
		return fmt.Errorf("failed to insert genesis block: %w", err)
	}

	fmt.Println("Genesis Block initialized successfully")
	return nil
}
