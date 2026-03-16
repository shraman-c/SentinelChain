# SentinelChain - Context & Documentation

## Project Overview

**SentinelChain** is a lightweight, private blockchain designed specifically for high-throughput SIEM log management and instant tamper detection. It uses Go (Golang), SQLite, and React.js to create a real-time security monitoring system.

### Key Metrics
- **Tamper Detection Latency**: ~100-500ms
- **Storage**: SQLite (single file blockchain.db)
- **Hashing**: SHA-256 via Go crypto/sha256

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      SentinelChain                               │
├─────────────────────────────────────────────────────────────────┤
│  Phase 1: Database & Core Models                                │
│  ├── SQLite with SHA-256 hashing                                │
│  ├── Genesis Block initialization                                │
│  └── Block CRUD operations                                      │
├─────────────────────────────────────────────────────────────────┤
│  Phase 2: Network Server (HTTP + WebSocket)                    │
│  ├── POST /api/log - Submit log                                │
│  ├── GET  /api/logs - Get all blocks                           │
│  └── WS   /ws/alerts - Real-time tamper alerts                 │
├─────────────────────────────────────────────────────────────────┤
│  Phase 3: Integrity Monitor                                     │
│  ├── Background goroutine (500ms interval)                      │
│  ├── Validates hash chain integrity                            │
│  └── Emits alerts on tamper detection                          │
├─────────────────────────────────────────────────────────────────┤
│  Phase 4: React Dashboard                                       │
│  ├── Real-time WebSocket connection                            │
│  ├── Blockchain ledger display                                  │
│  ├── Alert history panel                                        │
│  └── Red flash animation on tamper                             │
├─────────────────────────────────────────────────────────────────┤
│  Phase 5: Tamper Simulator                                      │
│  └── Direct DB modification (bypasses app)                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
SentinelChain/
├── bin/                          # Compiled binaries
│   ├── sentinelchain.exe         # Main server
│   ├── client.exe                # Test client
│   └── tamper_simulator.exe      # Tamper simulation
├── cmd/
│   ├── main.go                  # Main server entry
│   ├── client/                   # Test client
│   └── tamper_simulator/        # Tamper simulation
├── frontend/                     # React frontend (Vite + Tailwind)
│   ├── src/
│   │   ├── App.tsx             # Dashboard component
│   │   └── index.css           # Tailwind styles
│   ├── dist/                    # Production build
│   └── vite.config.ts          # Vite configuration
├── pkg/
│   ├── network/                 # HTTP/WebSocket server
│   │   └── server.go           # API + Integrity Monitor
│   ├── pb/                      # Message types
│   │   └── types.go            # LogRequest, LogResponse, TamperAlert
│   └── storage/                 # SQLite operations
│       ├── database.go          # DB connection & schema
│       ├── block.go             # Block model & hashing
│       └── genesis.go           # Genesis block init
├── proto/
│   └── schema.proto            # Protocol buffer definitions
├── blockchain.db               # SQLite database
├── Dockerfile                  # Docker configuration
├── railway.json               # Railway deployment config
├── DEPLOY.md                  # Deployment guide
└── go.mod                     # Go dependencies
```

---

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_timestamp INTEGER NOT NULL,
    source_ip TEXT NOT NULL,
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    prev_hash TEXT NOT NULL,
    hash TEXT NOT NULL,
    inserted_at INTEGER NOT NULL
);
CREATE INDEX idx_blocks_hash ON blocks(hash);
CREATE INDEX idx_blocks_prev_hash ON blocks(prev_hash);
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/log` | Submit a log entry |
| GET | `/api/logs` | Get all blocks |
| WS | `/ws/alerts` | Real-time tamper alerts |

### POST /api/log
```json
{
  "timestamp": 1234567890123456789,
  "source_ip": "192.168.1.100",
  "event_type": "AUTH_FAILURE",
  "severity": "WARNING",
  "message": "Failed login attempt"
}
```

### Response
```json
{
  "success": true,
  "hash": "abc123...",
  "message": "Log submitted successfully"
}
```

---

## Running Locally

### Backend
```bash
./bin/sentinelchain.exe --server http --port :8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# Visit http://localhost:3000
```

### Testing
```bash
# Send a log
curl -X POST http://localhost:8080/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"TEST","severity":"INFO","message":"Hello"}'

# Get all blocks
curl http://localhost:8080/api/logs

# Run tamper simulator (while server running)
./bin/tamper_simulator.exe
```

---

## Deployment

### Railway (Backend)
```bash
npm install -g @railway/cli
railway login
railway init
railway up
# Get URL: https://sentinelchain-xxx.up.railway.app
```

### Cloudflare Pages (Frontend)
```bash
cd frontend
npm run build
npm install -g wrangler
wrangler pages project create sentinelchain
wrangler pages deploy dist --project-name sentinelchain
# Set env var VITE_API_URL = your Railway URL
```

---

## Dependencies

### Go
- modernc.org/sqlite - Pure Go SQLite driver
- google.golang.org/grpc - gRPC support
- github.com/gorilla/websocket - WebSocket support

### Frontend
- React 18
- Vite
- Tailwind CSS 4
- TypeScript

---

## How It Works

1. **Log Ingestion**: Logs are received via HTTP POST, converted to blocks with SHA-256 hash
2. **Blockchain Chain**: Each block contains the previous block's hash, creating a cryptographic chain
3. **Integrity Monitor**: Background goroutine continuously validates the chain every 500ms
4. **Tamper Detection**: If any block is modified directly in the database, the hash mismatch is detected
5. **Alert Broadcasting**: Tamper alerts are sent via WebSocket to connected frontend clients

---

## Tamper Detection Latency

```
Alteration Time:  1773672611005485500 (nanoseconds)
Detection Time:   1773672611106511300 (nanoseconds)
Latency:          ~101ms
```

The integrity monitor checks every 500ms, so maximum detection latency is 500ms.

---

## Files Reference

| File | Purpose |
|------|---------|
| `pkg/storage/database.go` | SQLite connection & schema |
| `pkg/storage/block.go` | Block model & SHA256 hashing |
| `pkg/storage/genesis.go` | Genesis block creation |
| `pkg/network/server.go` | HTTP server, WebSocket, Integrity Monitor |
| `pkg/pb/types.go` | Message type definitions |
| `cmd/main.go` | Main application entry |
| `frontend/src/App.tsx` | React dashboard component |
| `cmd/tamper_simulator/main.go` | Direct DB tamper script |

---

## Development Notes

- Uses pure Go SQLite (modernc.org/sqlite) - no CGO required
- Genesis block has prev_hash = 64 zeros
- inserted_at stores Unix nanoseconds for precise latency measurement
- Integrity monitor never crashes the main server (goroutine isolation)
- Frontend polls /api/logs every 2s, connects to WebSocket for real-time alerts
