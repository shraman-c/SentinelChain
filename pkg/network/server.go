package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
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
	peers      []string
	readOnly   bool
	mu         sync.RWMutex
}

func NewLogHandler(db *storage.DB, tamperChan chan *pb.TamperAlert, peers []string, readOnly bool) *LogHandler {
	return &LogHandler{
		db:         db,
		tamperChan: tamperChan,
		peers:      peers,
		readOnly:   readOnly,
	}
}

func (h *LogHandler) applyLogRequest(req pb.LogRequest) (*storage.Block, error) {
	block := &storage.Block{
		LogTimestamp: req.Timestamp,
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		SourceIP:     req.SourceIP,
		EventType:    req.EventType,
		Severity:     req.Severity,
		Message:      req.Message,
	}

	lastBlock, err := h.db.GetLastBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block: %w", err)
	}

	if lastBlock != nil {
		block.PrevHash = lastBlock.Hash
	} else {
		block.PrevHash = storage.GenesisPrevHash
	}

	if err := h.db.InsertBlock(block); err != nil {
		return nil, err
	}

	return block, nil
}

func (h *LogHandler) broadcastToPeers(block *storage.Block) {
	for _, peer := range h.peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		payload, err := json.Marshal(pb.LogRequest{
			Timestamp:  block.LogTimestamp,
			DeviceID:   block.DeviceID,
			DeviceName: block.DeviceName,
			SourceIP:   block.SourceIP,
			EventType:  block.EventType,
			Severity:   block.Severity,
			Message:    block.Message,
			PrevHash:   block.PrevHash,
			Hash:       block.Hash,
		})
		if err != nil {
			log.Printf("Failed to marshal replication payload: %v", err)
			continue
		}

		url := strings.TrimRight(peer, "/") + "/api/replica/log"
		resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
		if err != nil {
			log.Printf("Failed to replicate block to %s: %v", peer, err)
			continue
		}
		resp.Body.Close()
	}
}

func (h *LogHandler) SubmitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.readOnly {
		http.Error(w, "Node is read-only", http.StatusForbidden)
		return
	}

	var req pb.LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	block, err := h.applyLogRequest(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert block: %v", err), http.StatusInternalServerError)
		return
	}

		h.broadcastToPeers(block)

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
		ID         int64  `json:"id"`
		Timestamp  int64  `json:"timestamp"`
		DeviceID   string `json:"device_id"`
		DeviceName string `json:"device_name"`
		SourceIP   string `json:"source_ip"`
		EventType  string `json:"event_type"`
		Severity   string `json:"severity"`
		Message    string `json:"message"`
		PrevHash   string `json:"prev_hash"`
		Hash       string `json:"hash"`
		InsertedAt int64  `json:"inserted_at"`
	}

	var response []blockResponse
	for _, b := range blocks {
		response = append(response, blockResponse{
			ID:         b.ID,
			Timestamp:  b.LogTimestamp,
			DeviceID:   b.DeviceID,
			DeviceName: b.DeviceName,
			SourceIP:   b.SourceIP,
			EventType:  b.EventType,
			Severity:   b.Severity,
			Message:    b.Message,
			PrevHash:   b.PrevHash,
			Hash:       b.Hash,
			InsertedAt: b.InsertedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LogHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"read_only": h.readOnly,
		"peer_count": len(h.peers),
	})
}

func (h *LogHandler) GetPeerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type peerStatus struct {
		Peer       string `json:"peer"`
		Reachable  bool   `json:"reachable"`
		StatusCode int    `json:"status_code,omitempty"`
		LatencyMs  int64  `json:"latency_ms"`
		Error      string `json:"error,omitempty"`
	}

	client := &http.Client{Timeout: 2 * time.Second}
	statuses := make([]peerStatus, 0, len(h.peers))
	connected := 0

	for _, peer := range h.peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		url := strings.TrimRight(peer, "/") + "/api/health"
		start := time.Now()
		resp, err := client.Get(url)
		latency := time.Since(start).Milliseconds()

		if err != nil {
			statuses = append(statuses, peerStatus{
				Peer:      peer,
				Reachable: false,
				LatencyMs: latency,
				Error:     err.Error(),
			})
			continue
		}

		reachable := resp.StatusCode >= 200 && resp.StatusCode < 300
		if reachable {
			connected++
		}

		statuses = append(statuses, peerStatus{
			Peer:       peer,
			Reachable:  reachable,
			StatusCode: resp.StatusCode,
			LatencyMs:  latency,
		})
		resp.Body.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"read_only":        h.readOnly,
		"configured_peers": len(statuses),
		"connected_peers":  connected,
		"peers":            statuses,
	})
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
			prevBlock.DeviceID,
			prevBlock.DeviceName,
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
			currentBlock.DeviceID,
			currentBlock.DeviceName,
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

func StartHTTPServer(port string, db *storage.DB, tamperChan chan *pb.TamperAlert, peers []string, readOnly bool) error {
	handler := NewLogHandler(db, tamperChan, peers, readOnly)

	http.HandleFunc("/api/log", handler.SubmitLog)
	http.HandleFunc("/api/logs", handler.GetLogs)
	http.HandleFunc("/api/health", handler.GetHealth)
	http.HandleFunc("/api/peer-status", handler.GetPeerStatus)
	http.HandleFunc("/api/replica/log", handler.SubmitReplicaLog)
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
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
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

func (h *LogHandler) SubmitReplicaLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pb.LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Hash == "" {
		http.Error(w, "Missing block hash", http.StatusBadRequest)
		return
	}

	exists, err := h.db.HasBlockHash(req.Hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check existing block: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		resp := pb.LogResponse{
			Success: true,
			Hash:    req.Hash,
			Message: "Replica block already exists",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	lastBlock, err := h.db.GetLastBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read chain tip: %v", err), http.StatusInternalServerError)
		return
	}
	if lastBlock == nil {
		if req.PrevHash != storage.GenesisPrevHash {
			http.Error(w, "Replica block does not extend the genesis chain", http.StatusBadRequest)
			return
		}
	} else if lastBlock.Hash != req.PrevHash {
		http.Error(w, "Replica block does not extend the current chain tip", http.StatusBadRequest)
		return
	}

	block := &storage.Block{
		LogTimestamp: req.Timestamp,
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		SourceIP:     req.SourceIP,
		EventType:    req.EventType,
		Severity:     req.Severity,
		Message:      req.Message,
		PrevHash:     req.PrevHash,
	}
	if err := h.db.InsertBlockWithExpectedHash(block, req.Hash); err != nil {
		http.Error(w, fmt.Sprintf("Failed to replicate block: %v", err), http.StatusBadRequest)
		return
	}

	resp := pb.LogResponse{
		Success: true,
		Hash:    block.Hash,
		Message: "Replica block accepted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SyncFromPeers(db *storage.DB, peers []string) error {
	if len(peers) == 0 {
		return nil
	}

	var lastHash string
	lastBlock, err := db.GetLastBlock()
	if err != nil {
		return err
	}
	if lastBlock != nil {
		lastHash = lastBlock.Hash
	}

	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}

		url := strings.TrimRight(peer, "/") + "/api/logs"
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Failed to sync from peer %s: %v", peer, err)
			continue
		}

		var blocks []struct {
			Timestamp  int64  `json:"timestamp"`
			DeviceID   string `json:"device_id"`
			DeviceName string `json:"device_name"`
			SourceIP   string `json:"source_ip"`
			EventType  string `json:"event_type"`
			Severity   string `json:"severity"`
			Message    string `json:"message"`
			PrevHash   string `json:"prev_hash"`
			Hash       string `json:"hash"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
			resp.Body.Close()
			log.Printf("Failed to decode peer chain from %s: %v", peer, err)
			continue
		}
		resp.Body.Close()

		for _, remoteBlock := range blocks {
			exists, err := db.HasBlockHash(remoteBlock.Hash)
			if err != nil {
				return err
			}
			if exists {
				lastHash = remoteBlock.Hash
				continue
			}

			if lastHash != "" && remoteBlock.PrevHash != lastHash {
				return fmt.Errorf("peer %s chain diverged at hash %s", peer, remoteBlock.Hash)
			}

			block := &storage.Block{
				LogTimestamp: remoteBlock.Timestamp,
				DeviceID:     remoteBlock.DeviceID,
				DeviceName:   remoteBlock.DeviceName,
				SourceIP:     remoteBlock.SourceIP,
				EventType:    remoteBlock.EventType,
				Severity:     remoteBlock.Severity,
				Message:      remoteBlock.Message,
				PrevHash:     remoteBlock.PrevHash,
			}

			if err := db.InsertBlockWithExpectedHash(block, remoteBlock.Hash); err != nil {
				return err
			}
			lastHash = block.Hash
		}
	}

	return nil
}
