# SentinelChain Deployment Guide

## Architecture

```
┌─────────────────────┐         ┌─────────────────────┐
│   Cloudflare Pages  │  ────>  │   Railway/Render    │
│   (React Frontend)  │         │   (Go Backend)     │
└─────────────────────┘         └─────────────────────┘
```

## Quick Deploy Options

### Option 1: Railway (Backend) + Cloudflare Pages (Frontend)

#### Step 1: Deploy Backend to Railway

1. **Create Railway Account**
   - Go to [railway.app](https://railway.app)
   - Sign up with GitHub

2. **Deploy**
   ```bash
   # Install Railway CLI
   npm install -g @railway/cli
   
   # Login
   railway login
   
   # Initialize project
   railway init
   # Select "Empty Project"
   
   # Deploy
   railway up
   ```

3. **Get Backend URL**
   - After deployment, Railway provides a URL like: `https://sentinelchain-backend.up.railway.app`

#### Step 2: Deploy Frontend to Cloudflare Pages

1. **Create Cloudflare Account**
   - Go to [dash.cloudflare.com](https://dash.cloudflare.com)

2. **Deploy**
   ```bash
   # Install Wrangler CLI
   npm install -g wrangler
   
   # Login
   wrangler login
   
   # Create pages project
   wrangler pages project create sentinelchain
   
   # Deploy
   wrangler pages deploy frontend/dist --project-name sentinelchain
   ```

3. **Configure Environment**
   - In Cloudflare Dashboard → Pages → sentinelchain → Settings → Environment Variables
   - Add: `VITE_API_URL` = your Railway backend URL

### Option 2: Render (Free Alternative)

1. **Deploy Backend**
   - Connect your GitHub repo to [render.com](https://render.com)
   - Create new Web Service
   - Build Command: `go build -o main ./cmd/main.go`
   - Start Command: `./main --server http --port $PORT`

2. **Deploy Frontend**
   - Create Static Site in render
   - Build Command: `cd frontend && npm install && npm run build`
   - Output Directory: `frontend/dist`

### Option 3: Fly.io (Edge Deployment)

```bash
# Install flyctl
winget install flyctl

# Launch
fly launch

# Deploy
fly deploy
```

## Manual Deployment

### Backend (Docker)

```bash
# Build
docker build -t sentinelchain .

# Run
docker run -p 8080:8080 sentinelchain
```

### Frontend

```bash
cd frontend

# Development
npm run dev

# Production build
npm run build

# Preview
npm run preview
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://localhost:8080` |
| `VITE_WS_URL` | WebSocket URL | `ws://localhost:8080` |

## Demo

A live demo is available at:
- **Frontend**: Coming soon
- **Backend**: Coming soon

## Testing the Deployment

```bash
# Test API
curl -X POST https://your-backend-url/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"TEST","severity":"INFO","message":"Hello"}'

# Get all blocks
curl https://your-backend-url/api/logs
```
