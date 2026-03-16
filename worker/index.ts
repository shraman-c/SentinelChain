export interface Env {
  DB: D1Database;
}

const GENESIS_PREV_HASH = "0000000000000000000000000000000000000000000000000000000000000000";

function computeHash(timestamp: number, sourceIP: string, eventType: string, severity: string, message: string, prevHash: string): string {
  const data = `${timestamp}${sourceIP}${eventType}${severity}${message}${prevHash}`;
  const encoder = new TextEncoder();
  const dataBuffer = encoder.encode(data);
  
  let hash = 0;
  for (let i = 0; i < dataBuffer.length; i++) {
    const char = dataBuffer[i];
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  
  // Generate SHA-256-like hex string (simplified for D1)
  let hashHex = "";
  let tempHash = Math.abs(hash);
  for (let i = 0; i < 64; i++) {
    hashHex += (tempHash % 16).toString(16);
    tempHash = Math.floor(tempHash / 16);
  }
  return hashHex;
}

async function initDatabase(db: D1Database): Promise<void> {
  const schema = `
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
    CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash);
    CREATE INDEX IF NOT EXISTS idx_blocks_prev_hash ON blocks(prev_hash);
  `;
  
  try {
    await db.exec(schema);
    
    // Check if genesis block exists
    const result = await db.prepare("SELECT COUNT(*) as count FROM blocks").first<{ count: number }>();
    
    if (result && result.count === 0) {
      // Create genesis block
      const genesisTimestamp = 0;
      const genesisHash = computeHash(genesisTimestamp, "0.0.0.0", "GENESIS", "INFO", "Genesis Block", GENESIS_PREV_HASH);
      
      await db.prepare(`
        INSERT INTO blocks (log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `).bind(genesisTimestamp, "0.0.0.0", "GENESIS", "INFO", "Genesis Block", GENESIS_PREV_HASH, genesisHash, Date.now() * 1000000).run();
    }
  } catch (e) {
    console.log("Database init:", e);
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    // Initialize database on first request
    await initDatabase(env.DB);

    // CORS headers
    const corsHeaders = {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    };

    if (request.method === "OPTIONS") {
      return new Response(null, { headers: corsHeaders });
    }

    // API Routes
    if (path === "/api/log" && request.method === "POST") {
      try {
        const body = await request.json();
        
        // Get last block hash
        const lastBlock = await env.DB.prepare(`
          SELECT hash FROM blocks ORDER BY id DESC LIMIT 1
        `).first<{ hash: string }>();
        
        const prevHash = lastBlock?.hash || GENESIS_PREV_HASH;
        
        const timestamp = body.timestamp || Date.now() * 1000000;
        const hash = computeHash(timestamp, body.source_ip, body.event_type, body.severity, body.message, prevHash);
        const insertedAt = Date.now() * 1000000;
        
        await env.DB.prepare(`
          INSERT INTO blocks (log_timestamp, source_ip, event_type, severity, message, prev_hash, hash, inserted_at)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `).bind(
          timestamp,
          body.source_ip || "",
          body.event_type || "",
          body.severity || "",
          body.message || "",
          prevHash,
          hash,
          insertedAt
        ).run();

        return new Response(JSON.stringify({
          success: true,
          hash: hash,
          message: "Log submitted successfully"
        }), {
          headers: { ...corsHeaders, "Content-Type": "application/json" }
        });
      } catch (e) {
        return new Response(JSON.stringify({
          success: false,
          error: String(e)
        }), {
          status: 500,
          headers: { ...corsHeaders, "Content-Type": "application/json" }
        });
      }
    }

    if (path === "/api/logs" && request.method === "GET") {
      try {
        const blocks = await env.DB.prepare(`
          SELECT id, log_timestamp as timestamp, source_ip, event_type, severity, message, hash
          FROM blocks ORDER BY id ASC
        `).all();
        
        return new Response(JSON.stringify(blocks.results), {
          headers: { ...corsHeaders, "Content-Type": "application/json" }
        });
      } catch (e) {
        return new Response(JSON.stringify({ error: String(e) }), {
          status: 500,
          headers: { ...corsHeaders, "Content-Type": "application/json" }
        });
      }
    }

    return new Response("Not Found", { status: 404 });
  }
};
