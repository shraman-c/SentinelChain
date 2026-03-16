# SentinelChain

A lightweight, private blockchain designed specifically for high-throughput SIEM log management and instant tamper detection.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)
![SQLite](https://img.shields.io/badge/SQLite-Loaded-003B57?style=flat&logo=sqlite)

## Features

- **Cryptographic Integrity**: SHA-256 hashing creates an immutable chain of SIEM logs
- **Instant Tamper Detection**: Background integrity monitor detects modifications within ~100-500ms
- **Real-time Alerts**: WebSocket-powered live dashboard with visual tamper notifications
- **High Performance**: Pure Go implementation with no external dependencies (no CGO)
- **Lightweight**: Single SQLite database file for the entire blockchain

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 18+
- npm

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/SentinelChain.git
cd SentinelChain

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend && npm install && cd ..
```

### Running

**Terminal 1 - Start Backend:**
```bash
./bin/sentinelchain.exe --server http --port :8080
```

**Terminal 2 - Start Frontend:**
```bash
cd frontend && npm run dev
```

Visit http://localhost:3000 to see the dashboard.

### Testing Tamper Detection

```bash
# While server is running, run the tamper simulator
./bin/tamper_simulator.exe
```

Watch the dashboard flash red when tampering is detected!

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

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/log` | Submit a log entry |
| GET | `/api/logs` | Get all blocks |
| WS | `/ws/alerts` | Real-time tamper alerts |

### Example Usage

```bash
# Submit a log
curl -X POST http://localhost:8080/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"AUTH_FAILURE","severity":"WARNING","message":"Failed login attempt"}'

# Get all blocks
curl http://localhost:8080/api/logs
```

## Tamper Detection Latency

```
Alteration Time:  1773672611005485500 (nanoseconds)
Detection Time:   1773672611106511300 (nanoseconds)
Latency:          ~101ms
```

The integrity monitor checks every 500ms, so maximum detection latency is 500ms.

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

See [DEPLOY.md](DEPLOY.md) for detailed deployment instructions.

## Project Structure

```
SentinelChain/
├── bin/                    # Compiled binaries
├── cmd/
│   ├── main.go            # Main server entry
│   ├── client/            # Test client
│   └── tamper_simulator/  # Tamper simulation
├── frontend/              # React frontend
│   ├── src/
│   │   ├── App.tsx
│   │   └── index.css
│   └── dist/
├── pkg/
│   ├── network/           # HTTP/WebSocket server
│   ├── pb/               # Message types
│   └── storage/          # SQLite operations
├── proto/
│   └── schema.proto
├── blockchain.db         # SQLite database
├── Dockerfile
├── railway.json
└── DEPLOY.md
```

## Technology Stack

- **Backend**: Go 1.26+, SQLite (modernc.org/sqlite)
- **Frontend**: React 18, Vite, Tailwind CSS 4, TypeScript
- **Real-time**: Gorilla WebSocket

## License

MIT
