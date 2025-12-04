#!/bin/bash
set -e

echo "🚀 Deploying PeerPigeon-Go to Fly.io"
echo "====================================="
echo ""

export PATH="/Users/danraeder/.fly/bin:$PATH"

PIGEONHUB_B="${PIGEONHUB_B:-pigeonhub-b}"
PIGEONHUB_C="${PIGEONHUB_C:-pigeonhub-c}"

echo "Target apps:"
echo "  Hub B (Bootstrap): $PIGEONHUB_B"
echo "  Hub C (Secondary): $PIGEONHUB_C"
echo ""

echo "================================================"
echo "Step 1: Deploying Bootstrap Hub ($PIGEONHUB_B)"
echo "================================================"
echo ""

flyctl deploy --app "$PIGEONHUB_B" --env IS_HUB=true --env HOST=0.0.0.0 --env PORT=3000

echo ""
echo "✅ Hub B deployed!"
echo ""
echo "⏳ Waiting 10 seconds for bootstrap hub..."
sleep 10
echo ""

echo "================================================"
echo "Step 2: Deploying Secondary Hub ($PIGEONHUB_C)"
echo "================================================"
echo ""

BOOTSTRAP_URL="ws://$PIGEONHUB_B.fly.dev:3000"
echo "Bootstrap URL: $BOOTSTRAP_URL"
echo ""

flyctl deploy --app "$PIGEONHUB_C" --env IS_HUB=true --env HOST=0.0.0.0 --env PORT=3000 --env BOOTSTRAP_HUBS="$BOOTSTRAP_URL"

echo ""
echo "✅ Hub C deployed!"
echo ""
echo "================================================"
echo "✅ Deployment Complete!"
echo "================================================"
echo ""
echo "🌐 Endpoints:"
echo "   Hub B: https://$PIGEONHUB_B.fly.dev"
echo "   Hub C: https://$PIGEONHUB_C.fly.dev"
echo ""
echo "🧪 Test:"
echo "   cd examples/interop && npm install"
echo "   HUB1_URL=wss://$PIGEONHUB_B.fly.dev HUB2_URL=wss://$PIGEONHUB_C.fly.dev npm run test:dual"
echo ""
