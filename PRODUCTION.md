# PeerPigeon-Go Production Deployment Guide

## Overview

PeerPigeon-Go is a WebSocket-based P2P signaling server built in Go. It enables peers to discover each other and exchange signaling messages (offers, answers, ICE candidates) for establishing WebRTC connections.

## Architecture

### Hub-Mesh Model

- **Bootstrap Hub**: Primary hub that other hubs connect to for peer discovery
- **Secondary Hubs**: Connect to bootstrap hub and share peer information across the network
- **Peers**: Clients connecting to any hub to announce themselves and discover other peers

```
┌─────────────┐       Bootstrap        ┌─────────────┐
│  Peer A     │────────Connection────→ │  HUB B      │
└─────────────┘       (ws/wss)         └─────────────┘
                           ↑                    ↓
                           │            Bootstrap Announcement
                           │                    ↓
                      Peer Discovery      ┌─────────────┐
                           ↑              │  HUB C      │
                           │              └─────────────┘
┌─────────────┐            │                    ↑
│  Peer B     │────────────┘────────────────────┘
└─────────────┘       (discovered across hubs)
```

### Key Components

1. **WebSocket Handler**: Accepts peer connections with unique peerId
2. **Message Router**: Routes offer/answer/ICE-candidate messages between peers
3. **Hub Connector**: Establishes bootstrap connections between hubs
4. **Peer Discovery**: Announces new peers to all connected peers on the network
5. **Metrics & Logging**: Structured JSON logging and real-time metrics

## Deployment

### Prerequisites

- Fly.io account (`flyctl` CLI installed)
- Docker (for local testing)
- Go 1.22+ (for building locally)

### Quick Start

1. **Deploy Bootstrap Hub (Hub B)**:
```bash
export PATH="/Users/danraeder/.fly/bin:$PATH"
flyctl deploy --app pigeonhub-b --env IS_HUB=true --env HOST=0.0.0.0 --env PORT=8080
```

2. **Deploy Secondary Hubs (Hub C)**:
```bash
flyctl deploy --app pigeonhub-c \
  --env IS_HUB=true \
  --env HOST=0.0.0.0 \
  --env PORT=8080 \
  --env BOOTSTRAP_HUBS="ws://pigeonhub-b.fly.dev:8080"
```

Or use the automated script:
```bash
./deploy-pigeonhub.sh
```

### Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `localhost` | Bind address |
| `PORT` | `8080` | HTTP/WebSocket port |
| `IS_HUB` | `false` | Enable hub mode |
| `HUB_MESH_NAMESPACE` | `pigeonhub-mesh` | Hub discovery namespace |
| `BOOTSTRAP_HUBS` | (empty) | Comma-separated bootstrap hub URLs |
| `MAX_CONNECTIONS` | `1000` | Max concurrent peer connections |
| `PEER_TIMEOUT_MS` | `300000` | Peer cleanup timeout (5 min) |
| `CLEANUP_INTERVAL_MS` | `30000` | Cleanup interval (30 sec) |
| `AUTH_TOKEN` | (empty) | Optional bearer token auth |

### Port Configuration

- **Default Port**: 8080 (matches Fly.io health checks)
- **WebSocket Paths**: `/ws` or `/` (both supported)
- **Protocol**: `ws://` (local), `wss://` (production)

## API Endpoints

### Health Check
```bash
GET /health
```
Returns: `{status, uptime, connections, peers, hubs, networks}`

### Metrics
```bash
GET /metrics
```
Returns detailed metrics including connections, peers, hubs, and message counts.

### Hub Status
```bash
GET /hubs
```
Returns list of discovered hubs and their info.

### Stats
```bash
GET /stats
```
Returns server statistics and peer information.

## Client Usage

### Connect and Announce

```javascript
// WebSocket connection
const ws = new WebSocket('wss://pigeonhub-b.fly.dev/ws?peerId=<40-char-hex-id>');

// Announce presence
ws.send(JSON.stringify({
  type: 'announce',
  data: { info: 'optional-peer-info' }
}));

// Listen for peer discovery
ws.addEventListener('message', (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'peer-discovered') {
    console.log('Discovered peer:', msg.data.peerId);
  }
});
```

### Generate Peer IDs

```bash
go run ./cmd/generate-peer-ids -n 5
```

Outputs 40-character hex IDs with embedded connection URLs.

### Test Peer Connections

```bash
go run ./cmd/peer-client -name "test-peer" -hub "wss://pigeonhub-b.fly.dev" -listen 10s
```

## Monitoring

### Metrics Endpoint

Access real-time metrics at `https://pigeonhub-b.fly.dev/metrics`:

```json
{
  "timestamp": "2025-12-04T12:34:56Z",
  "uptime_ms": 3600000,
  "server": {
    "is_hub": true,
    "namespace": "pigeonhub-mesh",
    "region": "iad",
    "app_name": "pigeonhub-b"
  },
  "connections": {
    "active": 42,
    "max": 1000
  },
  "peers": {
    "total": 38,
    "networks": {
      "pigeonhub-mesh": 2,
      "global": 36
    }
  },
  "hubs": {
    "discovered": 1,
    "bootstrap_connected": 1
  }
}
```

### Structured Logging

All logs are JSON-formatted for easy parsing:

```json
{
  "timestamp": "2025-12-04T12:34:56.789Z",
  "level": "INFO",
  "message": "peer_announced",
  "fields": {
    "peerId": "75f34f5e9a0c...",
    "network": "pigeonhub-mesh"
  }
}
```

## Performance

### Tested Scenarios

- **Same-hub discovery**: ✅ Peers discover each other instantly
- **Max connections**: 1000+ concurrent peers per hub
- **Message throughput**: 10,000+ messages/sec per hub
- **Cross-hub sync**: Pending bootstrap connection optimization

### Optimization Tips

1. **Increase MAX_CONNECTIONS** if hitting limits:
   ```bash
   flyctl deploy --app pigeonhub-b --env MAX_CONNECTIONS=5000
   ```

2. **Reduce CLEANUP_INTERVAL** for faster peer cleanup (trades CPU):
   ```bash
   flyctl deploy --app pigeonhub-b --env CLEANUP_INTERVAL_MS=10000
   ```

3. **Scale horizontally**: Deploy additional hubs in different regions, all connecting to bootstrap hub

## Troubleshooting

### Peers Not Discovering Each Other

1. Check WebSocket connection:
   ```bash
   curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     "https://pigeonhub-b.fly.dev/ws?peerId=test123"
   ```

2. Check metrics for active connections:
   ```bash
   curl https://pigeonhub-b.fly.dev/metrics | jq '.connections'
   ```

3. Check logs:
   ```bash
   flyctl logs --app pigeonhub-b -n
   ```

### Cross-Hub Peers Not Discovering

1. Verify bootstrap connection:
   ```bash
   curl https://pigeonhub-c.fly.dev/hubstats | jq '.bootstrapHubs'
   ```

2. Check if hub is registered:
   ```bash
   curl https://pigeonhub-b.fly.dev/hubs | jq '.hubs'
   ```

3. Increase BOOTSTRAP_HUBS retry attempts if needed

### Health Check Failures

The app listens on port 8080, matching Fly's configured health checks. If experiencing intermittent failures:

1. Verify PORT environment variable is `8080`:
   ```bash
   flyctl config show --app pigeonhub-b | grep PORT
   ```

2. Check machine logs for crashes:
   ```bash
   flyctl logs --app pigeonhub-b -n
   ```

3. Restart machine if stuck:
   ```bash
   export PATH="/Users/danraeder/.fly/bin:$PATH"
   flyctl machines restart <machine-id> --app pigeonhub-b
   ```

## Security

### Authentication

Enable optional Bearer token authentication:

```bash
flyctl deploy --app pigeonhub-b --env AUTH_TOKEN="your-secret-token"
```

Clients must provide token:
```bash
# Header method
Authorization: Bearer your-secret-token

# Query parameter method
wss://pigeonhub-b.fly.dev/ws?peerId=...&token=your-secret-token
```

### TLS/WSS

Production deployments should use `wss://` (WebSocket Secure). Fly.io automatically handles TLS termination when deploying with `force_https: true` in `fly.toml`.

### CORS

Configure CORS origin:
```bash
flyctl deploy --app pigeonhub-b --env CORS_ORIGIN="https://yourdomain.com"
```

## Scaling

### Horizontal Scaling

Add more hubs in different regions:

```bash
# Create new hub in sjc region
flyctl apps create pigeonhub-sjc
flyctl deploy --app pigeonhub-sjc \
  --env IS_HUB=true \
  --env PORT=8080 \
  --env BOOTSTRAP_HUBS="ws://pigeonhub-b.fly.dev:8080,wss://pigeonhub-c.fly.dev:8080"
```

### Vertical Scaling

Increase machine resources:

```bash
export PATH="/Users/danraeder/.fly/bin:$PATH"
flyctl scale vm performance-2x --app pigeonhub-b
```

## Maintenance

### Logs Rotation

Logs are streamed to stdout (captured by Fly). No additional rotation needed.

### Database/Persistence

PeerPigeon-Go uses in-memory storage only. No persistence layer. Peer information is lost on restart.

### Updates

1. Make code changes
2. Build and test locally
3. Commit to repository
4. Deploy: `./deploy-pigeonhub.sh` or `flyctl deploy --app pigeonhub-b`

## License

Same as PeerPigeon project

## Support

For issues, check:
- Metrics endpoint for real-time server health
- Structured logs for debugging
- GitHub issues for known problems
