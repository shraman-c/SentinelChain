package main

import (
	"fmt"
	"log"

	"sentinelchain/pkg/storage"
)

func main() {
	fmt.Println("Initializing SentinelChain Database...")

	db, err := storage.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}
	fmt.Println("Database schema initialized")

	if err := db.InitGenesisBlock(); err != nil {
		log.Fatalf("Failed to initialize genesis block: %v", err)
	}

	lastBlock, err := db.GetLastBlock()
	if err != nil {
		log.Fatalf("Failed to get last block: %v", err)
	}

	if lastBlock != nil {
		fmt.Printf("Last Block - ID: %d, Hash: %s, PrevHash: %s\n", lastBlock.ID, lastBlock.Hash, lastBlock.PrevHash)
	}

	blocks, err := db.GetAllBlocks()
	if err != nil {
		log.Fatalf("Failed to get all blocks: %v", err)
	}

	fmt.Printf("Total blocks in database: %d\n", len(blocks))
	for _, b := range blocks {
		fmt.Printf("  Block %d: %s\n", b.ID, b.Message)
	}

	fmt.Println("\nPhase 1: Database & Core Models - SUCCESS")
}
