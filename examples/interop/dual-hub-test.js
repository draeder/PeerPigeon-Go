#!/usr/bin/env node

/**
 * Dual Hub Cross-Connection Test
 * 
 * Tests peer discovery across two hub instances.
 * Connects one client to each hub and verifies they discover each other.
 */

const { PeerPigeonClient, generatePeerId } = require('./test-client');

// Configuration
const HUB1_URL = process.env.HUB1_URL || 'ws://localhost:3000';
const HUB2_URL = process.env.HUB2_URL || 'ws://localhost:3001';
const NETWORK_NAME = process.env.NETWORK_NAME || 'global';
const AUTH_TOKEN = process.env.AUTH_TOKEN || '';

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function runDualHubTest() {
  console.log('🧪 PeerPigeon-Go Dual Hub Test');
  console.log('===============================\n');
  console.log(`Hub 1: ${HUB1_URL}`);
  console.log(`Hub 2: ${HUB2_URL}`);
  console.log(`Network: ${NETWORK_NAME}\n`);

  const peerId1 = generatePeerId();
  const peerId2 = generatePeerId();

  console.log(`Client 1: ${peerId1.substring(0, 8)}... → Hub 1`);
  console.log(`Client 2: ${peerId2.substring(0, 8)}... → Hub 2\n`);

  const client1 = new PeerPigeonClient(HUB1_URL, peerId1, NETWORK_NAME);
  const client2 = new PeerPigeonClient(HUB2_URL, peerId2, NETWORK_NAME);

  let client1DiscoveredClient2 = false;
  let client2DiscoveredClient1 = false;

  // Set up discovery tracking
  client1.on('peer-discovered', (msg) => {
    if (msg.data?.peerId === peerId2) {
      client1DiscoveredClient2 = true;
      console.log('✅ Client 1 discovered Client 2 (cross-hub discovery working!)');
    }
  });

  client2.on('peer-discovered', (msg) => {
    if (msg.data?.peerId === peerId1) {
      client2DiscoveredClient1 = true;
      console.log('✅ Client 2 discovered Client 1 (cross-hub discovery working!)');
    }
  });

  try {
    // Connect both clients
    console.log('🔌 Connecting clients...\n');
    await Promise.all([
      client1.connect(),
      client2.connect()
    ]);

    await sleep(1000);

    // Announce both
    console.log('📢 Announcing clients...\n');
    client1.announce({ testClient: true, clientNum: 1 });
    client2.announce({ testClient: true, clientNum: 2 });

    // Wait for cross-hub discovery
    console.log('⏳ Waiting for cross-hub discovery (10 seconds)...\n');
    await sleep(10000);

    // Test signaling if discovered
    if (client1DiscoveredClient2) {
      console.log('📤 Client 1 → Client 2: Sending offer');
      client1.sendOffer(peerId2, 'test-offer-sdp');
      await sleep(500);
    }

    if (client2DiscoveredClient1) {
      console.log('📤 Client 2 → Client 1: Sending answer');
      client2.sendAnswer(peerId1, 'test-answer-sdp');
      await sleep(500);
    }

    // Results
    console.log('\n' + '='.repeat(50));
    console.log('📊 Test Results');
    console.log('='.repeat(50));
    console.log(`Client 1 peers discovered: ${client1.discoveredPeers.size}`);
    console.log(`Client 2 peers discovered: ${client2.discoveredPeers.size}`);
    console.log(`Cross-hub discovery C1→C2: ${client1DiscoveredClient2 ? '✅ SUCCESS' : '❌ FAILED'}`);
    console.log(`Cross-hub discovery C2→C1: ${client2DiscoveredClient1 ? '✅ SUCCESS' : '❌ FAILED'}`);

    if (client1DiscoveredClient2 && client2DiscoveredClient1) {
      console.log('\n✅ DUAL HUB TEST PASSED! Cross-hub discovery is working.');
    } else {
      console.log('\n⚠️  PARTIAL SUCCESS: Some discovery issues detected.');
      console.log('\nTroubleshooting:');
      console.log('1. Verify both hubs are running:');
      console.log(`   curl ${HUB1_URL.replace('ws://', 'http://').replace('/ws', '')}/health`);
      console.log(`   curl ${HUB2_URL.replace('ws://', 'http://').replace('/ws', '')}/health`);
      console.log('2. Check hub stats:');
      console.log(`   curl ${HUB1_URL.replace('ws://', 'http://').replace('/ws', '')}/hubstats`);
      console.log(`   curl ${HUB2_URL.replace('ws://', 'http://').replace('/ws', '')}/hubstats`);
      console.log('3. Verify BOOTSTRAP_HUBS is set on Hub 2');
      console.log('4. Check both hubs are in hub mode (IS_HUB=true)');
    }

    // Cleanup
    console.log('\n👋 Disconnecting clients...');
    client1.disconnect();
    client2.disconnect();

    await sleep(1000);

    process.exit(client1DiscoveredClient2 && client2DiscoveredClient1 ? 0 : 1);

  } catch (err) {
    console.error('\n❌ Test failed:', err.message);
    client1.disconnect();
    client2.disconnect();
    process.exit(1);
  }
}

// Run test
if (require.main === module) {
  runDualHubTest().catch(err => {
    console.error('Fatal error:', err);
    process.exit(1);
  });
}

module.exports = { runDualHubTest };
