package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"sentinelchain/pkg/pb"
	"sentinelchain/pkg/storage"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogHandler struct {
	db         *storage.DB
	tamperChan chan *pb.TamperAlert
	mu         sync.RWMutex
}

func NewLogHandler(db *storage.DB, tamperChan chan *pb.TamperAlert) *LogHandler {
	return &LogHandler{
		db:         db,
		tamperChan: tamperChan,
	}
}

func (h *LogHandler) SubmitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pb.LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	block := &storage.Block{
		LogTimestamp: req.Timestamp,
		SourceIP:     req.SourceIP,
		EventType:    req.EventType,
		Severity:     req.Severity,
		Message:      req.Message,
	}

	lastBlock, err := h.db.GetLastBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get last block: %v", err), http.StatusInternalServerError)
		return
	}

	if lastBlock != nil {
		block.PrevHash = lastBlock.Hash
	} else {
		block.PrevHash = storage.GenesisPrevHash
	}

	if err := h.db.InsertBlock(block); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert block: %v", err), http.StatusInternalServerError)
		return
	}

	resp := pb.LogResponse{
		Success: true,
		Hash:    block.Hash,
		Message: "Log submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *LogHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	blocks, err := h.db.GetAllBlocks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get blocks: %v", err), http.StatusInternalServerError)
		return
	}

	type blockResponse struct {
		ID        int64  `json:"id"`
		Timestamp int64  `json:"timestamp"`
		SourceIP  string `json:"source_ip"`
		EventType string `json:"event_type"`
		Severity  string `json:"severity"`
		Message   string `json:"message"`
		Hash      string `json:"hash"`
	}

	var response []blockResponse
	for _, b := range blocks {
		response = append(response, blockResponse{
			ID:        b.ID,
			Timestamp: b.LogTimestamp,
			SourceIP:  b.SourceIP,
			EventType: b.EventType,
			Severity:  b.Severity,
			Message:   b.Message,
			Hash:      b.Hash,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type integrityWatcher struct {
	db         *storage.DB
	tamperChan chan *pb.TamperAlert
	interval   time.Duration
	stopChan   chan bool
}

func NewIntegrityWatcher(db *storage.DB, tamperChan chan *pb.TamperAlert, interval time.Duration) *integrityWatcher {
	return &integrityWatcher{
		db:         db,
		tamperChan: tamperChan,
		interval:   interval,
		stopChan:   make(chan bool),
	}
}

func (iw *integrityWatcher) Start() {
	go iw.run()
}

func (iw *integrityWatcher) Stop() {
	close(iw.stopChan)
}

func (iw *integrityWatcher) run() {
	ticker := time.NewTicker(iw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iw.validateChain()
		case <-iw.stopChan:
			return
		}
	}
}

func (iw *integrityWatcher) validateChain() {
	blocks, err := iw.db.GetAllBlocks()
	if err != nil {
		log.Printf("Integrity check: failed to get blocks: %v", err)
		return
	}

	if len(blocks) <= 1 {
		return
	}

	for i := 1; i < len(blocks); i++ {
		currentBlock := blocks[i]
		prevBlock := blocks[i-1]

		computedHash := storage.ComputeHash(
			prevBlock.LogTimestamp,
			prevBlock.SourceIP,
			prevBlock.EventType,
			prevBlock.Severity,
			prevBlock.Message,
			prevBlock.PrevHash,
		)

		if computedHash != currentBlock.PrevHash {
			detectedAt := time.Now().UnixNano()
			alert := &pb.TamperAlert{
				DetectedAt:      detectedAt,
				TamperedBlockID: currentBlock.ID,
				Details:         fmt.Sprintf("Block %d: prev_hash mismatch. Expected %s, got %s", currentBlock.ID, computedHash, currentBlock.PrevHash),
			}

			log.Printf("🚨 TAMPER DETECTED! Block ID: %d, Detected At: %d", currentBlock.ID, detectedAt)
			log.Printf("   Details: %s", alert.Details)

			iw.tamperChan <- alert
		}

		currentComputedHash := storage.ComputeHash(
			currentBlock.LogTimestamp,
			currentBlock.SourceIP,
			currentBlock.EventType,
			currentBlock.Severity,
			currentBlock.Message,
			currentBlock.PrevHash,
		)

		if currentComputedHash != currentBlock.Hash {
			detectedAt := time.Now().UnixNano()
			alert := &pb.TamperAlert{
				DetectedAt:      detectedAt,
				TamperedBlockID: currentBlock.ID,
				Details:         fmt.Sprintf("Block %d: hash mismatch. Expected %s, got %s", currentBlock.ID, currentComputedHash, currentBlock.Hash),
			}

			log.Printf("🚨 TAMPER DETECTED! Block ID: %d, Detected At: %d", currentBlock.ID, detectedAt)
			log.Printf("   Details: %s", alert.Details)

			iw.tamperChan <- alert
		}
	}
}

func StartHTTPServer(port string, db *storage.DB, tamperChan chan *pb.TamperAlert) error {
	handler := NewLogHandler(db, tamperChan)

	http.HandleFunc("/api/log", handler.SubmitLog)
	http.HandleFunc("/api/logs", handler.GetLogs)
	http.HandleFunc("/ws/alerts", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, tamperChan)
	})

	integrityWatcher := NewIntegrityWatcher(db, tamperChan, 500*time.Millisecond)
	integrityWatcher.Start()

	log.Printf("HTTP server listening on %s", port)
	log.Printf("WebSocket alerts endpoint: ws://<host>%s/ws/alerts", port)
	log.Printf("Integrity monitor running (checking every 500ms)")

	if err := http.ListenAndServe(port, nil); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, tamperChan chan *pb.TamperAlert) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket client connected")

	for {
		select {
		case <-r.Context().Done():
			return
		case alert := <-tamperChan:
			alertJSON, err := json.Marshal(alert)
			if err != nil {
				log.Printf("Failed to marshal alert: %v", err)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, alertJSON); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func StartTCPServer(port string, db *storage.DB, tamperChan chan *pb.TamperAlert) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	integrityWatcher := NewIntegrityWatcher(db, tamperChan, 500*time.Millisecond)
	integrityWatcher.Start()

	log.Printf("TCP server listening on %s", port)
	log.Printf("Integrity monitor running (checking every 500ms)")

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, db, tamperChan)
	}
}

func handleConnection(conn net.Conn, db *storage.DB, tamperChan chan *pb.TamperAlert) {
	defer conn.Close()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read: %v", err)
		return
	}

	var req pb.LogRequest
	if err := json.Unmarshal(buf[:n], &req); err != nil {
		conn.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))
		return
	}

	block := &storage.Block{
		LogTimestamp: req.Timestamp,
		SourceIP:     req.SourceIP,
		EventType:    req.EventType,
		Severity:     req.Severity,
		Message:      req.Message,
	}

	lastBlock, err := db.GetLastBlock()
	if err != nil {
		conn.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))
		return
	}

	if lastBlock != nil {
		block.PrevHash = lastBlock.Hash
	} else {
		block.PrevHash = storage.GenesisPrevHash
	}

	if err := db.InsertBlock(block); err != nil {
		conn.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))
		return
	}

	resp := pb.LogResponse{
		Success: true,
		Hash:    block.Hash,
		Message: "Log submitted successfully",
	}

	respBytes, _ := json.Marshal(resp)
	conn.Write(respBytes)
}
