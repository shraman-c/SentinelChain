package pb

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

type Empty struct{}

type TamperAlert struct {
	DetectedAt      int64  `json:"detected_at"`
	TamperedBlockID int64  `json:"tampered_block_id"`
	Details         string `json:"details"`
}
