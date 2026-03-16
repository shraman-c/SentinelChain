Product Requirements Document (PRD)
1. Problem Statement
In the event of a system breach, attackers often manipulate or delete system logs to cover their tracks. Traditional databases cannot natively prove whether a log has been historically altered.

2. Proposed Solution
A localized, Go-based blockchain ledger that ingests security logs via gRPC, cryptographically chains them using SHA-256, and stores them in SQLite. A continuous background process validates the chain's integrity, instantly flagging anomalies.

3. Functional Requirements
Log Ingestion: The system must accept incoming log data streams (timestamp, source, event type, message).

Block Creation: The system must automatically hash incoming logs with the previous block's hash before committing them to the database.

Integrity Validation: A background service must continuously or periodically traverse the ledger to recalculate and verify hashes.

Alerting: The system must immediately trigger an alert to the frontend if a hash mismatch is found.

Tamper Simulation: A dedicated script or endpoint must be available to intentionally bypass the application logic and alter a database row to test the detection mechanism.

4. Non-Functional Requirements
Performance: Must handle high-velocity log ingestion suitable for a SIEM environment.

Latency: Tamper detection latency should be strictly minimized.

Storage: The blockchain footprint must remain minimal, relying on a lightweight local database.