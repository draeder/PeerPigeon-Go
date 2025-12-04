#!/usr/bin/env node

/**
 * Generate random 40-character hex peer IDs for PeerPigeon
 * Usage: node generate-peer-ids.js [count]
 */

const crypto = require('crypto');

function generatePeerId() {
  return crypto.randomBytes(20).toString('hex');
}

const count = parseInt(process.argv[2]) || 1;

console.log(`🎲 Generating ${count} random peer ID(s):\n`);

const urls = [
  'ws://pigeonhub-b.fly.dev/ws',
  'wss://pigeonhub-b.fly.dev/ws',
  'ws://pigeonhub-c.fly.dev/ws',
  'wss://pigeonhub-c.fly.dev/ws'
];

for (let i = 0; i < count; i++) {
  const peerId = generatePeerId();
  console.log(`${i + 1}. ${peerId}`);
  urls.forEach(url => {
    console.log(`   ${url}?peerId=${peerId}`);
  });
  console.log();
}

console.log(`✅ ${count} peer ID(s) generated.\n`);
console.log('Copy and use any of these to connect:');
console.log('  NODE: const ws = new WebSocket(url);');
console.log('  BROWSER: const ws = new WebSocket(url);');
console.log('  WSCAT: wscat -c "ws://host:3000/ws?peerId=<id>"');
