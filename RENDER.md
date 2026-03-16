# Render Deployment Guide

## Prerequisites
- GitHub account
- Render account (free tier)

---

## Part 1: Deploy Backend (Go API)

### Step 1: Push Code to GitHub

```bash
# Initialize git (if not already)
git init
git add .
git commit -m "Initial commit"

# Create GitHub repo and push
git remote add origin https://github.com/YOUR_USERNAME/SentinelChain.git
git branch -M main
git push -u origin main
```

### Step 2: Deploy to Render

1. Go to [render.com](https://render.com) and sign up with GitHub
2. Click **"New +"** → **"Web Service"**
3. Connect your GitHub repository
4. Configure:

| Setting | Value |
|---------|-------|
| Name | `sentinelchain-api` |
| Branch | `main` |
| Runtime | `Go` |
| Build Command | `go build -o main ./cmd/main.go` |
| Start Command | `./main --server http --port $PORT` |

5. Click **"Create Web Service"**

6. Wait for deployment (~2-3 minutes)

7. **Important**: After deployment, note your URL:
   ```
   https://sentinelchain-api.onrender.com
   ```

---

## Part 2: Deploy Frontend (React)

### Step 1: Update Environment Variable

1. In your GitHub repo, edit `frontend/.env.production`:
```bash
VITE_API_URL=https://sentinelchain-api.onrender.com
VITE_WS_URL=wss://sentinelchain-api.onrender.com
```

2. Commit and push:
```bash
git add .
git commit -m "Update production API URL"
git push
```

### Step 2: Deploy to Render

1. Go to [render.com](https://render.com)
2. Click **"New +"** → **"Static Site"**
3. Connect your GitHub repository
4. Configure:

| Setting | Value |
|---------|-------|
| Name | `sentinelchain-dashboard` |
| Branch | `main` |
| Build Command | `cd frontend && npm install && npm run build` |
| Publish directory | `frontend/dist` |

5. Click **"Create Static Site"**

6. Wait for deployment

7. Your frontend URL:
   ```
   https://sentinelchain-dashboard.onrender.com
   ```

---

## Part 3: Test

Visit your frontend URL (e.g., `https://sentinelchain-dashboard.onrender.com`)

### Test API:
```bash
# Replace with your backend URL
BACKEND_URL="https://sentinelchain-api.onrender.com"

# Submit a log
curl -X POST $BACKEND_URL/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"10.0.0.1","event_type":"TEST","severity":"INFO","message":"Hello"}'

# Get all blocks
curl $BACKEND_URL/api/logs
```

---

## Troubleshooting

### Backend Shows "Service Unavailable"
- Check Build Logs in Render dashboard
- Ensure `go.mod` has correct module name
- Verify Build Command runs locally first

### Frontend Can't Connect to Backend
- Verify `.env.production` has correct URL
- Backend must allow CORS (already configured)

### WebSocket Not Working
- Render's free tier may have WebSocket limitations
- Use HTTP polling as fallback (frontend already supports this)

---

## Quick Commands Reference

```bash
# Local development
./bin/sentinelchain.exe --server http --port :8080
cd frontend && npm run dev

# After deployment test
curl -X POST https://YOUR-API-URL/api/log \
  -H "Content-Type: application/json" \
  -d '{"timestamp":1234567890,"source_ip":"192.168.1.1","event_type":"LOGIN","severity":"INFO","message":"Test"}'
```
