#!/bin/bash

# Test cross-hub peer discovery

echo "=== Starting Cross-Hub Peer Discovery Test ==="
echo "Hub-B peer will run for 35 seconds"
echo "Hub-C peer will start after 5 seconds and run for 25 seconds"
echo ""

# Start hub-B peer in background, redirect to file
/tmp/peer-client -hub wss://pigeonhub-b.fly.dev -listen 35s > /tmp/peer-b.log 2>&1 &
B_PID=$!
echo "Started Hub-B peer (PID: $B_PID)"

# Wait 5 seconds
sleep 5

# Start hub-C peer in background, redirect to file
/tmp/peer-client -hub wss://pigeonhub-c.fly.dev -listen 25s > /tmp/peer-c.log 2>&1 &
C_PID=$!
echo "Started Hub-C peer (PID: $C_PID)"

# Wait for both to finish
wait $B_PID $C_PID 2>/dev/null

echo ""
echo "=== Hub-B Peer Output ==="
cat /tmp/peer-b.log
echo ""
echo "=== Hub-C Peer Output ==="
cat /tmp/peer-c.log
echo ""
echo "=== Test Complete ==="
