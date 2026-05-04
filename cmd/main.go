package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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
	peers := flag.String("peers", "", "Comma-separated peer base URLs to replicate to")
	bootstrapFrom := flag.String("bootstrap-from", "", "Comma-separated peer base URLs to bootstrap from on startup")
	readOnly := flag.Bool("read-only", false, "Reject direct log writes and only accept replicated blocks")
	flag.Parse()

	parsePeers := func(raw string) []string {
		parts := strings.Split(raw, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

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

	if bootstrapPeers := parsePeers(*bootstrapFrom); len(bootstrapPeers) > 0 {
		log.Printf("Bootstrapping chain from peers: %v", bootstrapPeers)
		if err := network.SyncFromPeers(db, bootstrapPeers); err != nil {
			log.Fatalf("Failed to bootstrap chain from peers: %v", err)
		}
		log.Println("Chain bootstrap completed")
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
			log.Fatal(network.StartHTTPServer(*port, db, tamperChan, parsePeers(*peers), *readOnly))
		case "tcp":
			log.Printf("Starting TCP server on %s", *port)
			log.Fatal(network.StartTCPServer(*port, db, tamperChan))
		default:
			log.Fatalf("Unknown server type: %s", *serverType)
		}
	}()

	fmt.Println("Phase 2 & 3: gRPC/Network Server + Integrity Monitor")
	fmt.Println("========================================================")
	fmt.Printf("Mode: %s%s\n", func() string {
		if *readOnly {
			return "replica-only"
		}
		return "leader"
	}(), func() string {
		if *peers == "" {
			return ""
		}
		return fmt.Sprintf(" | peers=%s", *peers)
	}())
	fmt.Println("Server running. Press Ctrl+C to stop.")
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	close(tamperChan)
}
