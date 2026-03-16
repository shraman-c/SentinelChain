package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LogRequest struct {
	Timestamp int64  `json:"timestamp"`
	SourceIP  string `json:"source_ip"`
	EventType string `json:"event_type"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type LogResponse struct {
	Success bool   `json:"success"`
	Hash    string `json:"hash"`
	Message string `json:"message"`
}

func main() {
	serverURL := "http://localhost:8080/api/log"

	logEntry := LogRequest{
		Timestamp: time.Now().UnixNano(),
		SourceIP:  "192.168.1.100",
		EventType: "AUTH_FAILURE",
		Severity:  "WARNING",
		Message:   "Failed login attempt for user admin",
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("Failed to marshal: %v\n", err)
		return
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to decode response: %v\n", err)
		return
	}

	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Hash: %s\n", result.Hash)
	fmt.Printf("Message: %s\n", result.Message)
}
