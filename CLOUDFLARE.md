# Cloudflare D1 Deployment Guide

## Prerequisites
- Cloudflare account (free)
- No credit card required

---

## Step 1: Install Wrangler CLI

```bash
npm install -g wrangler
wrangler login
```

---

## Step 2: Create D1 Database

```bash
# Create D1 database
wrangler d1 create sentinelchain-db

# It will output something like:
# database_id = "abc123..."
database_id = 0c56ddf8-28bd-4be5-88e0-a3c75811753c

# Copy that ID and update wrangler.toml
```

---

## Step 3: Update Configuration

Edit `wrangler.toml` with your database ID:

```toml
[[d1_databases]]
binding = "DB"
database_name = "sentinelchain-db"
database_id = "YOUR_DATABASE_ID_HERE"
```

---

## Step 4: Deploy Worker

```bash
# Deploy the worker
wrangler deploy
```

Your API will be live at: `https://sentinelchain-worker.YOUR_ACCOUNT.workers.dev`
  https://sentinelchain-worker.shramanchaudhuri.workers.dev
---

## Step 5: Deploy Frontend

```bash
cd frontend
npm run build

# Deploy to Cloudflare Pages
wrangler pages deploy dist --project-name sentinelchain
```

---

## Step 6: Update Frontend Environment

After deploying, update your frontend to point to the worker URL:

```bash
# In Cloudflare Dashboard:
# Pages → sentinelchain → Settings → Environment Variables

VITE_API_URL=https://sentinelchain-worker.YOUR_ACCOUNT.workers.dev
```

---

## Test Your Deployment

```bash
# Replace with your worker URL
WORKER_URL="https://sentinelchain-worker.YOUR_ACCOUNT.workers.dev"

# Submit a log
curl -X POST $WORKER_URL/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"TEST","severity":"INFO","message":"Hello"}'

# Get all blocks
curl $WORKER_URL/api/logs
```

---

## Quick Deploy Commands

```bash
# Install
npm install -g wrangler
wrangler login

# Create DB
wrangler d1 create sentinelchain-db

# Update wrangler.toml with your database_id

# Deploy Worker
wrangler deploy

# Deploy Frontend
cd frontend && npm run build
wrangler pages deploy dist --project-name sentinelchain
```

---

## Troubleshooting

### Worker not found
- Make sure you ran `wrangler deploy`
- Check your account ID in the dashboard

### Database errors
- Run `wrangler d1 execute sentinelchain-db --local` to test locally
- Check database bindings in wrangler.toml

### CORS issues
- The worker already includes CORS headers
- Make sure to use full worker URL in frontend
