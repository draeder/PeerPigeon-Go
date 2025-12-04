#!/usr/bin/env node

/**
 * PeerPigeon-Go Interoperability Test Client
 * 
 * Tests WebSocket connection, announce, peer discovery, and signaling
 * against deployed PeerPigeon-Go servers.
 */

const WebSocket = require('ws');
const crypto = require('crypto');

// Configuration from environment variables
const SERVER_URL = process.env.SERVER_URL || 'ws://localhost:3000';
const NETWORK_NAME = process.env.NETWORK_NAME || 'global';
const AUTH_TOKEN = process.env.AUTH_TOKEN || '';

// Generate a random 40-character hex peer ID
function generatePeerId() {
  return crypto.randomBytes(20).toString('hex');
}

class PeerPigeonClient {
  constructor(serverUrl, peerId, networkName = 'global') {
    this.serverUrl = serverUrl;
    this.peerId = peerId;
    this.networkName = networkName;
    this.ws = null;
    this.connected = false;
    this.discoveredPeers = new Map();
    this.messageHandlers = new Map();
  }

  connect() {
    return new Promise((resolve, reject) => {
      const url = `${this.serverUrl}/ws?peerId=${this.peerId}`;
      const options = {};
      
      if (AUTH_TOKEN) {
        options.headers = {
          'Authorization': `Bearer ${AUTH_TOKEN}`
        };
      }

      console.log(`🔌 Connecting to ${url}`);
      this.ws = new WebSocket(url, options);

      this.ws.on('open', () => {
        console.log(`✅ Connected as ${this.peerId.substring(0, 8)}...`);
        this.connected = true;
        resolve();
      });

      this.ws.on('message', (data) => {
        try {
          const msg = JSON.parse(data.toString());
          this.handleMessage(msg);
        } catch (err) {
          console.error('❌ Failed to parse message:', err);
        }
      });

      this.ws.on('close', (code, reason) => {
        console.log(`🔌 Disconnected: ${code} ${reason || ''}`);
        this.connected = false;
      });

      this.ws.on('error', (err) => {
        console.error('❌ WebSocket error:', err.message);
        reject(err);
      });

      // Timeout
      setTimeout(() => {
        if (!this.connected) {
          reject(new Error('Connection timeout'));
        }
      }, 5000);
    });
  }

  handleMessage(msg) {
    console.log(`📩 Received [${msg.type}]:`, JSON.stringify(msg, null, 2));

    switch (msg.type) {
      case 'connected':
        console.log('✅ Server confirmed connection');
        break;
      
      case 'peer-discovered':
        const peerId = msg.data?.peerId;
        if (peerId && peerId !== this.peerId) {
          this.discoveredPeers.set(peerId, msg.data);
          console.log(`👥 Discovered peer: ${peerId.substring(0, 8)}... (isHub: ${msg.data.isHub || false})`);
        }
        break;
      
      case 'peer-disconnected':
        const disconnectedId = msg.data?.peerId;
        if (disconnectedId) {
          this.discoveredPeers.delete(disconnectedId);
          console.log(`👋 Peer disconnected: ${disconnectedId.substring(0, 8)}...`);
        }
        break;
      
      case 'offer':
      case 'answer':
      case 'ice-candidate':
        console.log(`📞 Signaling message from ${msg.fromPeerId?.substring(0, 8)}...`);
        break;
      
      case 'pong':
        console.log('🏓 Pong received');
        break;
      
      default:
        console.log(`📬 Unknown message type: ${msg.type}`);
    }

    // Call custom handlers
    const handler = this.messageHandlers.get(msg.type);
    if (handler) {
      handler(msg);
    }
  }

  on(messageType, handler) {
    this.messageHandlers.set(messageType, handler);
  }

  send(msg) {
    if (!this.connected) {
      console.error('❌ Not connected');
      return false;
    }
    this.ws.send(JSON.stringify(msg));
    return true;
  }

  announce(data = {}) {
    console.log(`📢 Announcing to network: ${this.networkName}`);
    return this.send({
      type: 'announce',
      networkName: this.networkName,
      data: { ...data, isHub: false }
    });
  }

  sendOffer(targetPeerId, sdp) {
    console.log(`📤 Sending offer to ${targetPeerId.substring(0, 8)}...`);
    return this.send({
      type: 'offer',
      targetPeer: targetPeerId,
      networkName: this.networkName,
      data: { sdp }
    });
  }

  sendAnswer(targetPeerId, sdp) {
    console.log(`📤 Sending answer to ${targetPeerId.substring(0, 8)}...`);
    return this.send({
      type: 'answer',
      targetPeer: targetPeerId,
      networkName: this.networkName,
      data: { sdp }
    });
  }

  sendIceCandidate(targetPeerId, candidate) {
    console.log(`📤 Sending ICE candidate to ${targetPeerId.substring(0, 8)}...`);
    return this.send({
      type: 'ice-candidate',
      targetPeer: targetPeerId,
      networkName: this.networkName,
      data: { candidate }
    });
  }

  ping() {
    console.log('🏓 Sending ping');
    return this.send({ type: 'ping' });
  }

  disconnect() {
    if (this.ws) {
      console.log('👋 Disconnecting...');
      this.send({ type: 'goodbye' });
      this.ws.close();
    }
  }
}

// Main test
async function runTest() {
  console.log('🧪 PeerPigeon-Go Interoperability Test');
  console.log('======================================\n');
  console.log(`Server: ${SERVER_URL}`);
  console.log(`Network: ${NETWORK_NAME}`);
  console.log(`Auth: ${AUTH_TOKEN ? 'Enabled' : 'Disabled'}\n`);

  const peerId = generatePeerId();
  const client = new PeerPigeonClient(SERVER_URL, peerId, NETWORK_NAME);

  try {
    // Connect
    await client.connect();

    // Wait a moment
    await new Promise(resolve => setTimeout(resolve, 500));

    // Announce
    client.announce({ testClient: true, timestamp: Date.now() });

    // Wait for peer discovery
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Ping test
    client.ping();

    // If peers discovered, test signaling
    if (client.discoveredPeers.size > 0) {
      const targetPeer = Array.from(client.discoveredPeers.keys())[0];
      console.log(`\n🔄 Testing signaling with peer ${targetPeer.substring(0, 8)}...`);
      
      client.sendOffer(targetPeer, 'mock-sdp-offer-data');
      await new Promise(resolve => setTimeout(resolve, 500));
      
      client.sendIceCandidate(targetPeer, 'mock-ice-candidate');
    } else {
      console.log('\n⚠️  No peers discovered (expected if testing alone)');
    }

    // Wait before disconnecting
    await new Promise(resolve => setTimeout(resolve, 2000));

    console.log('\n✅ Test completed successfully!');
    console.log(`📊 Discovered ${client.discoveredPeers.size} peer(s)`);

    client.disconnect();

    // Exit after cleanup
    setTimeout(() => process.exit(0), 1000);

  } catch (err) {
    console.error('\n❌ Test failed:', err.message);
    process.exit(1);
  }
}

// Run if executed directly
if (require.main === module) {
  runTest().catch(err => {
    console.error('Fatal error:', err);
    process.exit(1);
  });
}

module.exports = { PeerPigeonClient, generatePeerId };
