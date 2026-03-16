package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sentinelchain/pkg/network"
	"sentinelchain/pkg/pb"
	"sentinelchain/pkg/storage"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           SentinelChain - SIEM Blockchain                ║")
	fmt.Println("║        Lightweight Private Blockchain for Logs          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	serverType := flag.String("server", "http", "Server type: http, tcp, or grpc")
	port := flag.String("port", ":8080", "Server port")
	flag.Parse()

	log.Println("Initializing SentinelChain Database...")

	db, err := storage.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}
	log.Println("Database schema initialized")

	if err := db.InitGenesisBlock(); err != nil {
		log.Fatalf("Failed to initialize genesis block: %v", err)
	}

	tamperChan := make(chan *pb.TamperAlert, 100)
	go func() {
		for alert := range tamperChan {
			fmt.Printf("\n🚨 TAMPER ALERT 🚨\n")
			fmt.Printf("   Detected At: %d (nanoseconds)\n", alert.DetectedAt)
			fmt.Printf("   Block ID:    %d\n", alert.TamperedBlockID)
			fmt.Printf("   Details:     %s\n\n", alert.Details)
		}
	}()

	go func() {
		switch *serverType {
		case "http":
			log.Printf("Starting HTTP server on %s", *port)
			log.Fatal(network.StartHTTPServer(*port, db, tamperChan))
		case "tcp":
			log.Printf("Starting TCP server on %s", *port)
			log.Fatal(network.StartTCPServer(*port, db, tamperChan))
		default:
			log.Fatalf("Unknown server type: %s", *serverType)
		}
	}()

	fmt.Println("Phase 2 & 3: gRPC/Network Server + Integrity Monitor")
	fmt.Println("========================================================")
	fmt.Println("Server running. Press Ctrl+C to stop.")
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	close(tamperChan)
}
