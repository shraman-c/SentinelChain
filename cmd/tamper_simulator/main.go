package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          SentinelChain Tamper Simulator               ║")
	fmt.Println("║    This script directly modifies the database         ║")
	fmt.Println("║    to simulate a database tamper attack                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	db, err := sql.Open("sqlite", "blockchain.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&count)
	if err != nil {
		fmt.Printf("Failed to count blocks: %v\n", err)
		return
	}

	fmt.Printf("Total blocks in database: %d\n", count)

	if count <= 1 {
		fmt.Println("Not enough blocks to tamper. Need at least 2 blocks.")
		fmt.Println("Run the server and add some logs first.")
		return
	}

	rand.Seed(time.Now().UnixNano())
	blockID := rand.Int63n(int64(count-1)) + 2

	var message, eventType string
	err = db.QueryRow("SELECT message, event_type FROM blocks WHERE id = ?", blockID).Scan(&message, &eventType)
	if err != nil {
		fmt.Printf("Failed to get block: %v\n", err)
		return
	}

	fmt.Printf("\nSelected Block ID: %d\n", blockID)
	fmt.Printf("Original message: %s\n", message)
	fmt.Printf("Original event_type: %s\n", eventType)

	alterationTime := time.Now().UnixNano()
	fmt.Printf("\n>>> ALTERATION TIME (nanoseconds): %d\n", alterationTime)

	newMessage := message + "_TAMPERED"
	newEventType := "ATTACK"

	_, err = db.Exec("UPDATE blocks SET message = ?, event_type = ? WHERE id = ?", newMessage, newEventType, blockID)
	if err != nil {
		fmt.Printf("Failed to tamper: %v\n", err)
		return
	}

	fmt.Printf(">>> Tamper executed: message='%s', event_type='%s'\n", newMessage, newEventType)
	fmt.Println("\nThe Integrity Monitor should detect this tamper within 500ms.")
	fmt.Printf("Detection Latency = Detection_Time - %d\n", alterationTime)
}
