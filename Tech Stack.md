# Tech Stack Documentation
## Project: Lightweight Private Blockchain for SIEM Logs

### 📌 Overview
This document outlines the technology stack selected for building a high-performance, lightweight private blockchain designed to secure Security Information and Event Management (SIEM) logs. The stack is optimized for ultra-low latency, high concurrent throughput, and minimal computational overhead, allowing for the precise measurement of **Tamper Detection Latency**.

---

### 🏗️ System Architecture



---

### 🛠️ Core Technologies

#### 1. Backend / Core Engine: Go (Golang)
* **Role:** Powers the blockchain logic, data ingestion, and background integrity validation.
* **Justification:** Go is a compiled language renowned for its execution speed and memory efficiency. Its native concurrency model (Goroutines) allows the system to process thousands of incoming logs simultaneously while a separate background thread continuously recalculates hashes to detect tampering without blocking ingestion.

#### 2. Storage / The Ledger: SQLite
* **Role:** Acts as the immutable storage layer (the "chain").
* **Justification:** As a C-language library that implements a small, fast, self-contained, high-reliability SQL database engine, SQLite eliminates the network overhead of traditional databases. Because it stores data in a single local file, it perfectly facilitates our testing phase—allowing us to manually edit the file to simulate an attacker altering historical logs.

#### 3. Cryptography: SHA-256 (Go Native `crypto/sha256`)
* **Role:** Secures the chain by linking blocks cryptographically.
* **Justification:** SHA-256 is the industry standard for cryptographic hashing. By utilizing Go's standard library, we eliminate third-party dependencies, reducing potential security vulnerabilities and maximizing hashing speed.

#### 4. Data Ingestion: gRPC
* **Role:** The communication protocol for streaming logs from external devices/servers into the blockchain.
* **Justification:** gRPC operates over HTTP/2 and uses Protocol Buffers (Protobuf) to serialize data. It is significantly faster and lighter than traditional REST APIs (JSON over HTTP/1.1), making it the ideal choice for high-velocity SIEM log ingestion where every millisecond counts.

#### 5. Frontend UI: React.js + Tailwind CSS
* **Role:** Provides the "SIEM Control Room" dashboard for monitoring and alerts.
* **Justification:** React allows for the creation of dynamic, real-time user interfaces. When the Go backend detects a tamper event, it can push an alert to the React frontend (via WebSockets), which will immediately flag the compromised block. Tailwind CSS ensures rapid, utility-first styling for a clean, professional security dashboard.

#### 6. Stress Testing & Evaluation: Apache JMeter
* **Role:** Simulates high-traffic network environments to validate system performance.
* **Justification:** To prove the system's viability, JMeter will be used to bombard the gRPC endpoint with thousands of mock logs per second. This establishes a baseline of system performance under load, ensuring the tamper detection latency remains low even during heavy SIEM traffic.