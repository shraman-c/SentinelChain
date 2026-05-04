# Deploying SentinelChain to Cloudflare Container Workers

This guide covers deploying the SentinelChain backend as a Cloudflare Container Worker. Container Workers allow running Docker containers globally on Cloudflare's edge network, making it ideal for Go backends with SQLite.

## Prerequisites

- **Cloudflare account** (paid plan required for Containers)
- **Wrangler CLI** v4.87+
- **Docker** (or Podman with Docker CLI compatibility)
- **Node.js** 18+

## Installation

### 1. Install Wrangler CLI

```bash
npm install -g wrangler
```

### 2. Install dependencies

```bash
npm install
```

### 3. Setup Container Runtime

**For Docker (Recommended):**

```bash
# Fedora/RHEL
sudo dnf install docker docker-compose
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
newgrp docker
```

**For Podman (Alternative):**

```bash
# Install podman-docker for Docker CLI compatibility
sudo dnf install podman-docker

# Start the Podman socket
systemctl --user enable --now podman.socket

# Set DOCKER_HOST environment variable
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock

# Make it permanent
echo 'export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock' >> ~/.bashrc
source ~/.bashrc
```

### 4. Login to Cloudflare

```bash
wrangler login
```

## Project Structure

```
SentinelChain/
├── Dockerfile              # Container build configuration
├── wrangler.jsonc          # Cloudflare Worker configuration
├── src/
│   └── index.ts           # Worker entry point (routes to containers)
├── cmd/
│   └── main.go            # Go backend application
└── package.json            # NPM dependencies
```

## Configuration Files

### wrangler.jsonc

```jsonc
{
  "name": "sentinelchain",
  "main": "src/index.ts",
  "compatibility_date": "2026-05-04",
  "containers": [
    {
      "class_name": "Backend",
      "image": "./Dockerfile",
      "max_instances": 3
    }
  ],
  "durable_objects": {
    "bindings": [
      { "class_name": "Backend", "name": "BACKEND" }
    ]
  },
  "migrations": [
    { "new_sqlite_classes": ["Backend"], "tag": "v1" }
  ]
}
```

### src/index.ts

```typescript
import { Container, getRandom } from "@cloudflare/containers";

export class Backend extends Container {
  defaultPort = 8080;
  sleepAfter = "2h";
}

export default {
  async fetch(request, env) {
    const instance = await getRandom(env.BACKEND, 3);
    return instance.fetch(request);
  },
};
```

### Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/blockchain.db ./

EXPOSE 8080

CMD ["sh", "-c", "echo 'Running SentinelChain...' && ./main --server http --port $PORT"]
```

## Local Development

### Test locally before deploying

```bash
wrangler dev
```

This will:
1. Build the Docker container image
2. Start the Worker locally
3. Route requests to the container

Test the backend:

```bash
curl http://localhost:8787/api/health
```

## Deployment

### Deploy to Cloudflare

```bash
wrangler deploy
```

During deployment:
1. Wrangler builds your Docker image locally
2. Pushes the image to Cloudflare's registry (backed by R2)
3. Deploys the Worker and provisions container instances globally

**Note:** After first deployment, wait several minutes for container instances to be provisioned globally.

### Verify deployment

```bash
# Replace with your deployed URL
curl https://sentinelchain.your-subdomain.workers.dev/api/health
```

## Redeploying

When you make changes to the Go code or Dockerfile:

```bash
wrangler deploy
```

Wrangler will only push changed image layers, making redeployments fast.

## Environment Variables

To add environment variables to your containers, update `wrangler.jsonc`:

```jsonc
{
  "containers": [
    {
      "class_name": "Backend",
      "image": "./Dockerfile",
      "max_instances": 3,
      "vars": {
        "DATABASE_PATH": "/app/blockchain.db",
        "LOG_LEVEL": "info"
      }
    }
  ]
}
```

Or use secrets:

```bash
wrangler secret put MY_SECRET_KEY
```

## Troubleshooting

### Docker daemon not running

```bash
# Start Docker
sudo systemctl start docker
```

### "InvalidSymbol" error in wrangler.jsonc

JSONC requires double quotes around all keys and string values:

```json
❌ { name: "sentinelchain" }
✅ { "name": "sentinelchain" }
```

### Missing compatibility_date

Add to `wrangler.jsonc`:

```json
{
  "compatibility_date": "2026-05-04"
}
```

Or pass via CLI:

```bash
wrangler deploy --compatibility-date 2026-05-04
```

### Podman compatibility issues

If Wrangler fails to detect Podman, explicitly set the Docker binary:

```bash
export WRANGLER_DOCKER_BIN=podman
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
```

### Container provisioning timeout

After first deployment, containers take a few minutes to provision globally. Wait and retry requests.

## Scaling

The `max_instances` setting controls maximum container instances:

```jsonc
{
  "containers": [
    {
      "class_name": "Backend",
      "max_instances": 10  // Increase for higher traffic
    }
  ]
}
```

The `getRandom()` helper in `src/index.ts` load balances across instances. Future versions will include automatic latency-aware routing.

## Limitations

- **Container size:** Limited by available disk space on Cloudflare's infrastructure
- **Memory:** Each container has memory limits (check Cloudflare docs for current limits)
- **Cold starts:** First request to a new container instance takes longer while it boots
- **SQLite persistence:** Database state is container-local. For shared state across instances, consider Cloudflare D1 or R2 storage
- **Paid plan required:** Container Workers require a paid Cloudflare Workers plan

## Cleanup

To remove your deployment:

```bash
wrangler delete
```

To remove container images from the registry:

```bash
wrangler containers images list
wrangler containers images delete <IMAGE_NAME>
```
