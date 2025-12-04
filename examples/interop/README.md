# PeerPigeon-Go Interoperability Tests

Node.js test clients to verify WebSocket protocol compatibility with deployed PeerPigeon-Go servers.

## Setup

```bash
cd examples/interop
npm install
```

## Single Server Test

Test basic connectivity, announce, and signaling against one server:

```bash
# Test local server
npm test

# Test remote server
SERVER_URL=ws://your-server-ip:3000 npm test

# Test with authentication
SERVER_URL=ws://server:3000 AUTH_TOKEN=your-token npm test

# Test custom network namespace
SERVER_URL=ws://server:3000 NETWORK_NAME=mynetwork npm test
```

### What it tests:
- WebSocket connection establishment
- Announce message and peer discovery
- Ping/pong health check
- Signaling messages (offer, answer, ICE candidates)
- Proper message formatting and protocol compliance

## Dual Hub Test

Test cross-hub peer discovery with two hub instances:

```bash
# Test local hubs on different ports
HUB1_URL=ws://localhost:3000 HUB2_URL=ws://localhost:3001 npm run test:dual

# Test deployed Oracle hubs
HUB1_URL=ws://instance1-ip:3000 HUB2_URL=ws://instance2-ip:3000 npm run test:dual

# With authentication
HUB1_URL=ws://hub1:3000 HUB2_URL=ws://hub2:3000 AUTH_TOKEN=token npm run test:dual
```

### What it tests:
- Connects client A to hub 1
- Connects client B to hub 2
- Verifies client A discovers client B (cross-hub)
- Verifies client B discovers client A (cross-hub)
- Tests signaling relay across hubs
- Validates bootstrap hub connectivity

### Expected output (success):
```
✅ Client 1 discovered Client 2 (cross-hub discovery working!)
✅ Client 2 discovered Client 1 (cross-hub discovery working!)
📊 Test Results
==================================================
Client 1 peers discovered: 1
Client 2 peers discovered: 1
Cross-hub discovery C1→C2: ✅ SUCCESS
Cross-hub discovery C2→C1: ✅ SUCCESS

✅ DUAL HUB TEST PASSED! Cross-hub discovery is working.
```

## Configuration Options

All tests support these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_URL` | `ws://localhost:3000` | WebSocket server URL (single test) |
| `HUB1_URL` | `ws://localhost:3000` | First hub URL (dual test) |
| `HUB2_URL` | `ws://localhost:3001` | Second hub URL (dual test) |
| `NETWORK_NAME` | `global` | Network namespace to join |
| `AUTH_TOKEN` | (none) | Authentication token if required |

## Using the Client Programmatically

```javascript
const { PeerPigeonClient, generatePeerId } = require('./test-client');

const client = new PeerPigeonClient('ws://server:3000', generatePeerId(), 'global');

// Custom message handlers
client.on('peer-discovered', (msg) => {
  console.log('Peer discovered:', msg.data.peerId);
});

client.on('offer', (msg) => {
  console.log('Received offer from:', msg.fromPeerId);
  // Send answer back
  client.sendAnswer(msg.fromPeerId, 'mock-answer-sdp');
});

// Connect and announce
await client.connect();
client.announce({ myCustomData: 'value' });

// Send signaling
client.sendOffer(targetPeerId, sdpOffer);
client.sendIceCandidate(targetPeerId, candidate);

// Cleanup
client.disconnect();
```

## Troubleshooting

### Connection refused
- Verify server is running: `curl http://server:3000/health`
- Check firewall rules allow port 3000
- For Oracle Cloud, verify Security List ingress rules

### Authentication errors
- Set `AUTH_TOKEN` environment variable
- Verify token matches server configuration

### No peers discovered
- Ensure both clients use same `NETWORK_NAME`
- Check if server is in hub mode for cross-hub discovery
- Verify bootstrap hub connection: `curl http://server:3000/hubstats`

### Cross-hub discovery fails
1. Verify both servers are hubs (IS_HUB=true)
2. Check Hub 2 has BOOTSTRAP_HUBS pointing to Hub 1
3. Check hub stats endpoint:
   ```bash
   curl http://hub1:3000/hubstats
   curl http://hub2:3000/hubstats
   ```
4. Look for bootstrap connection status
5. Verify both hubs use same HUB_MESH_NAMESPACE

### Timeouts
- Increase wait times in test scripts
- Check network latency between hubs
- Verify no packet loss: `ping hub-ip`

## Protocol Reference

### Message Types Tested

**Client → Server:**
- `announce` - Join network and broadcast presence
- `offer` - WebRTC offer signaling
- `answer` - WebRTC answer signaling
- `ice-candidate` - ICE candidate signaling
- `ping` - Health check
- `goodbye` - Graceful disconnect

**Server → Client:**
- `connected` - Connection confirmation
- `peer-discovered` - New peer announcement
- `peer-disconnected` - Peer left network
- `offer` / `answer` / `ice-candidate` - Forwarded signaling
- `pong` - Ping response

### Message Format

All messages are JSON with this structure:

```json
{
  "type": "message-type",
  "networkName": "global",
  "fromPeerId": "sender-40-char-hex-id",
  "targetPeer": "recipient-40-char-hex-id",
  "data": { /* type-specific payload */ },
  "timestamp": 1234567890
}
```

## Next Steps

After successful interop tests:
1. Integrate with your WebRTC application
2. Implement proper SDP handling
3. Add connection quality monitoring
4. Set up production authentication
5. Configure TLS/WSS for production use
