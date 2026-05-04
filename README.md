# SentinelChain

> 🛡️ SentinelChain is a lightweight, private blockchain built for high-throughput SIEM log management with instant tamper detection. It provides cryptographic integrity for security logs, ensuring that any unauthorized modifications are detected in real-time through an immutable SHA-256 hash chain. Designed for enterprise security teams, SentinelChain combines a high-performance Go backend with a real-time React dashboard to monitor and protect critical log data from tampering.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)
![SQLite](https://img.shields.io/badge/SQLite-Loaded-003B57?style=flat&logo=sqlite)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

---

## 🚀 Quick Navigation

- [**Quick Start**](#-quick-start) - Get up and running in 5 minutes
- [**API Endpoints**](#-api-endpoints) - Explore available endpoints
- [**Architecture**](#-architecture) - Understand the system design
- [**Two-PC Setup**](#-two-pc-blockchain-setup) - Deploy across multiple machines
- [**Deployment**](#-deployment) - Deploy to cloud platforms

---

## ✨ Features

<table>
<tr>
<td>🔐 <b>Cryptographic Integrity</b><br>SHA-256 hashing creates an immutable chain of SIEM logs</td>
<td>⚡ <b>Instant Tamper Detection</b><br>Background integrity monitor detects modifications within ~100-500ms</td>
</tr>
<tr>
<td>📡 <b>Real-time Alerts</b><br>WebSocket-powered live dashboard with visual tamper notifications</td>
<td>🏎️ <b>High Performance</b><br>Pure Go implementation with no external dependencies (no CGO)</td>
</tr>
<tr>
<td>💾 <b>Lightweight</b><br>Single SQLite database file for the entire blockchain</td>
<td>🔗 <b>Two-PC Replication</b><br>Leader-replica architecture for distributed log management</td>
</tr>
</table>

## 🚀 Quick Start

### 📋 Prerequisites

Ensure you have the following installed:

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [npm](https://www.npmjs.com/)

### 📦 Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/SentinelChain.git
cd SentinelChain

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend && npm install && cd ..
```

### ▶️ Running the Application

<div>
<b>Step 1:</b> Start the Backend (Terminal 1)
</div>

```bash
go build -o bin/sentinelchain ./cmd
./bin/sentinelchain --server http --port :8080
```

<div>
<b>Step 2:</b> Start the Frontend (Terminal 2)
</div>

```bash
cd frontend && npm run dev:lan
```

🌐 **Open your browser:** [http://localhost:3000](http://localhost:3000)

### 🧪 Testing Tamper Detection

```bash
# While server is running, run the tamper simulator
go build -o bin/tamper_simulator ./cmd/tamper_simulator
./bin/tamper_simulator
```

👀 Watch the dashboard flash red when tampering is detected!

## 🏗️ Architecture

<details>
<summary><b>Click to expand architecture diagram</b></summary>

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

</details>

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/log` | Submit a log entry |
| `POST` | `/api/replica/log` | Accept a replicated block from a peer |
| `GET` | `/api/logs` | Get all blocks |
| `GET` | `/api/health` | Local node health and mode |
| `GET` | `/api/peer-status` | Connectivity to configured peers |
| `WS` | `/ws/alerts` | Real-time tamper alerts |

<details>
<summary><b>📝 View Example Usage</b></summary>

```bash
# Submit a log
curl -X POST http://localhost:8080/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"AUTH_FAILURE","severity":"WARNING","message":"Failed login attempt"}'

# Get all blocks
curl http://localhost:8080/api/logs

# Check local node health
curl http://localhost:8080/api/health

# Check whether this node can reach configured peers
curl http://localhost:8080/api/peer-status
```

</details>

<details>
<summary><b>🔗 Verify 2-PC Connection</b></summary>

On the leader, start with replica peer configured:

```bash
./bin/sentinelchain --server http --port :8080 --peers http://REPLICA_IP:8080
```

On the replica, start read-only and bootstrap from leader:

```bash
./bin/sentinelchain --server http --port :8080 --read-only --bootstrap-from http://LEADER_IP:8080
```

Then validate from each side:

```bash
# From leader: should show connected_peers: 1 and reachable: true
curl http://LEADER_IP:8080/api/peer-status

# From replica: should return status ok and read_only true
curl http://REPLICA_IP:8080/api/health
```

</details>

## Tamper Detection Latency

```
Alteration Time:  1773672611005485500 (nanoseconds)
Detection Time:   1773672611106511300 (nanoseconds)
Latency:          ~101ms
```

The integrity monitor checks every 500ms, so maximum detection latency is 500ms.

## Deployment

## Two-PC Blockchain Setup

Use one machine as the leader and the second as a read-only replica.

Leader PC:

```bash
./bin/sentinelchain --server http --port :8080 --peers http://REPLICA_IP:8080
```

Replica PC:

```bash
./bin/sentinelchain --server http --port :8080 --read-only --bootstrap-from http://LEADER_IP:8080
```

The replica pulls the current chain from the leader at startup and then accepts replicated blocks on `/api/replica/log`. The leader keeps writing logs and broadcasts each new block to its peers.

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

## 📄 License

This project is licensed under the MIT License - see below for details:

<details>
<summary><b>View MIT License</b></summary>

```
MIT License

Copyright (c) 2026 Shraman Chaudhuri

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

</details>
