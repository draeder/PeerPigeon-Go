# PeerPigeon

Language: English | [中文文档](README.zh-CN.md)

**Status**: ✅ Production Ready

PeerPigeon-Go is a high-performance WebSocket signaling server for P2P peer discovery and WebRTC signaling. Built in Go with hub-to-hub mesh networking for cross-network peer awareness.

## Key Features

- ⚡ **High Performance**: 100+ concurrent peers per hub, 90+ msg/sec throughput
- 🌐 **Hub Mesh Networking**: Connect multiple hubs for cross-network peer discovery
- 📊 **Production Observability**: Real-time metrics endpoint, structured JSON logging
- 🔐 **Optional Auth**: Bearer token authentication for secure deployments
- 🚀 **Easy Deployment**: Docker multi-stage build, Fly.io ready with custom health checks
- 📝 **WebSocket API**: Simple JSON message protocol for signaling

## Performance

Validated with 50+ concurrent peers:
- **Connection Success Rate**: 100%
- **Peer Discovery**: 2400+ discoveries per 15 seconds  
- **Message Throughput**: 90+ messages/second
- **Same-Hub Latency**: < 10ms peer discovery

## Quick Start

### Local Development

```bash
# Start a single hub
PORT=8080 HOST=localhost IS_HUB=true go run ./cmd/peerpigeon

# Check health
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

### Docker

```bash
docker build -t peerpigeon .
docker run -p 8080:8080 \
  -e IS_HUB=true \
  -e HOST=0.0.0.0 \
  peerpigeon
```

### Production Deployment (Fly.io)

See [PRODUCTION.md](PRODUCTION.md) for detailed deployment guide.

```bash
./deploy-pigeonhub.sh  # Deploy hub-b and hub-c with bootstrap
```

## Usage

### Generate Peer IDs

```bash
go run ./cmd/generate-peer-ids -n 5
```

Output includes connection URLs for all hubs.

### Connect a Peer

```bash
# Simple peer client for testing
go run ./cmd/peer-client \
  -name "test-peer" \
  -hub "wss://pigeonhub-b.fly.dev" \
  -listen 10s
```

### Load Testing

```bash
# Test 100 concurrent peers
go run ./cmd/load-test \
  -hub "wss://pigeonhub-b.fly.dev" \
  -peers 100 \
  -duration 30 \
  -interval 5
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | Bind address |
| `PORT` | `8080` | HTTP/WebSocket port |
| `IS_HUB` | `false` | Enable hub mode |
| `HUB_MESH_NAMESPACE` | `pigeonhub-mesh` | Hub discovery namespace |
| `BOOTSTRAP_HUBS` | (empty) | Comma-separated bootstrap hub URLs |
| `MAX_CONNECTIONS` | `1000` | Max concurrent connections |
| `PEER_TIMEOUT_MS` | `300000` | Peer idle timeout (5 min) |
| `CLEANUP_INTERVAL_MS` | `30000` | Cleanup interval (30 sec) |
| `AUTH_TOKEN` | (empty) | Optional bearer token authentication |
| `CORS_ORIGIN` | `*` | CORS allow origin |

### Examples

**Single hub (local)**:
```bash
PORT=8080 HOST=localhost go run ./cmd/peerpigeon
```

**Hub with bootstrap**:
```bash
PORT=8080 IS_HUB=true BOOTSTRAP_HUBS="wss://hub-b.example.com" \
  go run ./cmd/peerpigeon
```

**Production with auth**:
```bash
PORT=8080 IS_HUB=true AUTH_TOKEN="secret-token" \
  BOOTSTRAP_HUBS="wss://hub-b.example.com" \
  go run ./cmd/peerpigeon
```

## API Endpoints

### Health Check
```
GET /health
```

Returns health status, uptime, connection counts, and peer info.

### Metrics
```
GET /metrics
```

Returns detailed metrics including connections, peers, hubs, message counts.

### Hub Status
```
GET /hubs
GET /hubstats
GET /stats
```

Returns hub information, bootstrap connections, and server statistics.

## WebSocket Protocol

### Connect
```
ws://<host>:<port>/ws?peerId=<40-hex-id>
```

### Announce
```json
{
  "type": "announce",
  "networkName": "global",
  "data": { "info": "optional-peer-info" }
}
```

### Signaling
```json
{
  "type": "offer|answer|ice-candidate",
  "targetPeerId": "<peer-id>",
  "networkName": "global",
  "data": { "sdp": "..." }
}
```

### Peer Discovery (received)
```json
{
  "type": "peer-discovered",
  "data": { "peerId": "<peer-id>", "info": "..." },
  "networkName": "global"
}
```

## Architecture

See [PRODUCTION.md](PRODUCTION.md) for detailed architecture documentation.

### Hub Mesh Model

Multiple hubs can connect to a bootstrap hub for cross-network peer awareness:

```
┌──────────────────────────────────┐
│      Bootstrap Hub (Hub B)       │
│   - Primary hub                  │
│   - Stores peer information      │
│   - Forwards cross-hub messages  │
└─────────────────────┬────────────┘
                      │
        Bootstrap Connection (ws/wss)
                      │
┌─────────────────────▼────────────┐
│     Secondary Hub (Hub C)        │
│   - Connects to Hub B            │
│   - Shares local peers with B    │
│   - Receives remote peer info    │
└──────────────────────────────────┘
```

## Testing

### Local Load Test

```bash
# Connect 100 peers, run for 30 seconds
go run ./cmd/load-test -hub "ws://localhost:8080" -peers 100 -duration 30
```

### Production Load Test

```bash
go run ./cmd/load-test \
  -hub "wss://pigeonhub-b.fly.dev" \
  -peers 100 \
  -duration 30 \
  -interval 5
```

## Development

### Build

```bash
go build -o peerpigeon ./cmd/peerpigeon
```

### Code Structure

```
internal/
  server/        # Core server logic
    server.go    # WebSocket and HTTP handlers
    hubs.go      # Hub mesh connections
    types.go     # Message types
  logging/       # Structured JSON logging
  metrics/       # Observability metrics

cmd/
  peerpigeon/    # Main server binary
  peer-client/   # Test peer client
  load-test/     # Load testing utility
  generate-peer-ids/  # Peer ID generation
```

## License

Same as PeerPigeon project

## Related Projects

- **PeerPigeon (JavaScript)**: https://github.com/PeerPigeon/PeerPigeon
- **WebRTC Signaling**: Used for establishing peer connections
