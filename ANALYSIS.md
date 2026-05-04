# SentinelChain - Analysis of Results

## 1. System Architecture Analysis

**SentinelChain** is a lightweight, private blockchain designed for SIEM (Security Information and Event Management) log tamper detection. The system consists of:

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend | Go 1.25+ | Core blockchain logic, ingestion, integrity monitoring |
| Storage | SQLite (modernc.org/sqlite - pure Go, no CGO) | Immutable ledger storage |
| Cryptography | SHA-256 (Go `crypto/sha256`) | Block hashing and chain integrity |
| Frontend | React 18 + Vite + Tailwind CSS 4 | Real-time SIEM dashboard |
| Real-time Alerts | Gorilla WebSocket | Live tamper alert pushing |
| Network Protocols | HTTP + TCP | Log ingestion and peer replication |

## 2. Core Results & Performance Metrics

### 2.1 Tamper Detection Latency
- **Measured Latency**: ~101ms (from README example)
- **Maximum Theoretical Latency**: 500ms (integrity check interval)
- **Mechanism**: Background goroutine polls every 500ms, recalculating SHA-256 hashes for all blocks

### 2.2 Hash Chain Validation Logic (`server.go:327-391`)
The integrity validator performs **two checks per block pair**:
1. **PrevHash Verification**: Confirms `block[i].prev_hash == hash(block[i-1])`
2. **Self-Hash Verification**: Confirms `block[i].hash == computeHash(block[i])`

### 2.3 Block Structure (`block.go:11-23`)
Each block contains:
- **ID**: Auto-incrementing integer (SQLite `AUTOINCREMENT`)
- **LogTimestamp**: Nanosecond-precision Unix timestamp
- **DeviceID/DeviceName**: Source device identification
- **SourceIP**: Network source
- **EventType**: Log classification (e.g., `AUTH_FAILURE`, `LOGIN_SUCCESS`, `BRUTE_FORCE_ATTEMPT`)
- **Severity**: `INFO`, `WARNING`, `ERROR`, `CRITICAL`
- **Message**: Log payload
- **PrevHash**: SHA-256 hash of previous block
- **Hash**: SHA-256 hash of current block content
- **InsertedAt**: Nanosecond insertion timestamp

## 3. Functional Analysis

### 3.1 Log Ingestion Flow
```
Client → POST /api/log → GetLastBlock() → Set PrevHash → ComputeHash() → InsertBlock() → Broadcast to Peers
```

### 3.2 Peer Replication (Leader-Replica Model)
- **Leader**: Accepts direct writes via `/api/log`, broadcasts to peers via `/api/replica/log`
- **Replica**: Read-only mode (`--read-only` flag), bootstraps from leader via `SyncFromPeers()`
- **Validation**: Replica verifies expected hash matches computed hash before accepting blocks

### 3.3 Tamper Detection Flow
```
IntegrityWatcher (500ms ticker) → GetAllBlocks() → ValidateChain() → TamperAlert → tamperChan → WebSocket + Console
```

### 3.4 Tamper Simulator Results (`tamper_simulator/main.go`)
The simulator:
1. Selects a random block (ID 2 to N)
2. Records alteration time in nanoseconds
3. Modifies `message` (appends `_TAMPERED`) and `event_type` (sets to `ATTACK`)
4. Bypasses application logic (direct DB modification)

**Expected Result**: Detection within 500ms, with alert containing:
- `detected_at`: Nanosecond timestamp
- `tampered_block_id`: Compromised block ID
- `details`: Specific mismatch description

## 4. Database Schema Analysis (`database.go:49-66`)

```sql
CREATE TABLE IF NOT EXISTS blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    log_timestamp INTEGER NOT NULL,
    device_id TEXT NOT NULL DEFAULT '',
    device_name TEXT NOT NULL DEFAULT '',
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

**Indexes**: Two indexes on `hash` and `prev_hash` optimize chain traversal and lookup operations.

## 5. Frontend Dashboard Analysis (`App.tsx`)

### 5.1 Features
- **Block Count Display**: Real-time count of blockchain entries
- **Login Attack Simulator**: Form to generate test events with:
  - Device presets (5 predefined devices with IPs)
  - Simulation modes: Failed Login, Successful Login, Brute Force
  - Configurable username/password/attempts
- **Blockchain Ledger Table**: Displays all blocks with:
  - ID, Device, Source IP, Event Type, Severity (color-coded), Hash (truncated)
- **Auto-refresh**: Polls `/api/logs` every 2 seconds

### 5.2 Severity Color Coding
| Severity | Color |
|----------|-------|
| CRITICAL | Red (`text-red-400 bg-red-900/30`) |
| ERROR | Orange (`text-orange-400 bg-orange-900/30`) |
| WARNING | Yellow (`text-yellow-400 bg-yellow-900/30`) |
| INFO | Blue (`text-blue-400 bg-blue-900/30`) |

## 6. API Endpoints Analysis

| Method | Endpoint | Description | Key Logic |
|--------|----------|-------------|-----------|
| POST | `/api/log` | Submit log entry | Creates block, broadcasts to peers |
| POST | `/api/replica/log` | Accept replicated block | Validates hash, checks chain continuity |
| GET | `/api/logs` | Get all blocks | Returns ordered block list |
| GET | `/api/health` | Node health/mode | Returns status, read_only flag, peer count |
| GET | `/api/peer-status` | Peer connectivity | Pings each peer, measures latency |
| WS | `/ws/alerts` | Real-time tamper alerts | Streams alerts via WebSocket |

## 7. Security Analysis

### 7.1 Strengths
- **Cryptographic chaining**: SHA-256 prevents silent modifications
- **Continuous monitoring**: 500ms interval ensures rapid detection
- **Read-only replica mode**: Prevents accidental writes on replica nodes
- **Hash verification on replication**: `InsertBlockWithExpectedHash()` validates incoming blocks

### 7.2 Vulnerabilities/Limitations
- **No consensus mechanism**: Single leader can be compromised
- **SQLite file-level access**: Direct DB bypass is possible (by design for testing)
- **WebSocket CheckOrigin = true**: Allows any origin (security risk in production)
- **No authentication**: API endpoints lack auth mechanisms
- **Hash computation excludes device_id/device_name in some cases** (`block.go:91-94`): Conditional hash format could cause issues

## 8. Performance Considerations

### 8.1 Strengths
- **Pure Go SQLite**: No CGO overhead, easier cross-compilation
- **Goroutine-based concurrency**: Non-blocking ingestion and monitoring
- **Lightweight footprint**: Single database file, minimal dependencies

### 8.2 Bottlenecks
- **Full chain scan on every check**: `GetAllBlocks()` loads entire chain every 500ms (O(n) operation)
- **No pagination**: `/api/logs` returns all blocks
- **Linear block lookup**: `GetLastBlock()` uses `ORDER BY id DESC LIMIT 1`

## 9. Project Goals Achievement

| Goal | Status | Evidence |
|------|--------|----------|
| Instant Tamper Detection | Achieved | ~101ms latency demonstrated |
| Measure Detection Latency | Achieved | Nanosecond-precision timestamps in alerts |
| High Throughput, Low Overhead | Partially | No PoW/mining, but full-chain scans limit scalability |

## 10. Recommendations

1. **Optimize integrity checks**: Track last-validated block ID instead of full chain scan
2. **Add authentication**: JWT or API keys for `/api/log` and admin endpoints
3. **Implement pagination**: For `/api/logs` endpoint
4. **Restrict WebSocket origins**: Configure `CheckOrigin` properly
5. **Add unit/integration tests**: No test files found in codebase
6. **Consider Merkle trees**: For more efficient large-chain verification
7. **Fix conditional hash computation**: Ensure consistent hash format regardless of device fields
