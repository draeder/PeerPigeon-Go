# Quick Deployment Summary

## ✅ Created Files

### Deployment Infrastructure
- **`Dockerfile`** - Multi-stage Docker build for lightweight production image
- **`deploy/oracle-setup.sh`** - Automated Oracle Cloud deployment script
- **`deploy/README.md`** - Complete deployment guide with Oracle Cloud, Docker, and troubleshooting

### Interoperability Testing
- **`examples/interop/package.json`** - Node.js test dependencies
- **`examples/interop/test-client.js`** - Single server WebSocket test client
- **`examples/interop/dual-hub-test.js`** - Cross-hub discovery test
- **`examples/interop/README.md`** - Testing guide and protocol reference
- **`examples/interop/.gitignore`** - Node ignore patterns

### Documentation Updates
- **`README.md`** - Added deployment and testing sections

## 🚀 Deployment Steps (Oracle Cloud)

### Step 1: Create Two Oracle Cloud Instances
1. Sign up for Oracle Cloud (Always Free tier)
2. Create 2 Ubuntu instances in Always Free tier
3. Note both public IPs

### Step 2: Deploy Bootstrap Hub (Instance 1)
```bash
ssh ubuntu@INSTANCE1_IP
curl -o oracle-setup.sh https://raw.githubusercontent.com/draeder/PeerPigeon-Go/main/deploy/oracle-setup.sh
chmod +x oracle-setup.sh
sudo PORT=3000 IS_HUB=true ./oracle-setup.sh
```

### Step 3: Deploy Secondary Hub (Instance 2)
```bash
ssh ubuntu@INSTANCE2_IP
curl -o oracle-setup.sh https://raw.githubusercontent.com/draeder/PeerPigeon-Go/main/deploy/oracle-setup.sh
chmod +x oracle-setup.sh
sudo PORT=3000 IS_HUB=true BOOTSTRAP_HUBS='ws://INSTANCE1_IP:3000' ./oracle-setup.sh
```

### Step 4: Configure Oracle Cloud Security Lists
⚠️ **Critical:** Oracle blocks all ports by default!

1. Go to Oracle Cloud Console
2. Networking → Virtual Cloud Networks
3. Select VCN → Security Lists
4. Add Ingress Rule for **both instances**:
   - Source CIDR: `0.0.0.0/0`
   - Protocol: TCP
   - Port: 3000

### Step 5: Verify Deployment
```bash
# Check health (run from your machine)
curl http://INSTANCE1_IP:3000/health
curl http://INSTANCE2_IP:3000/health

# Check hub stats
curl http://INSTANCE1_IP:3000/hubstats
curl http://INSTANCE2_IP:3000/hubstats
```

## 🧪 Testing Cross-Hub Discovery

### Local Test Setup
```bash
cd examples/interop
npm install
```

### Test Both Hubs
```bash
HUB1_URL=ws://INSTANCE1_IP:3000 HUB2_URL=ws://INSTANCE2_IP:3000 npm run test:dual
```

### Expected Output (Success)
```
✅ Client 1 discovered Client 2 (cross-hub discovery working!)
✅ Client 2 discovered Client 1 (cross-hub discovery working!)
Cross-hub discovery C1→C2: ✅ SUCCESS
Cross-hub discovery C2→C1: ✅ SUCCESS

✅ DUAL HUB TEST PASSED! Cross-hub discovery is working.
```

## 📋 What the Scripts Do

### oracle-setup.sh
1. Updates system packages
2. Installs Go 1.22
3. Clones PeerPigeon-Go repository
4. Builds binary
5. Creates systemd service
6. Configures firewall (UFW + iptables)
7. Starts service automatically
8. Displays connection info

### test-client.js
- Connects to WebSocket server
- Generates random 40-char hex peer ID
- Sends announce message
- Listens for peer-discovered events
- Tests ping/pong
- Tests signaling (offer/answer/ICE)

### dual-hub-test.js
- Connects client A to hub 1
- Connects client B to hub 2
- Verifies A discovers B across hubs
- Verifies B discovers A across hubs
- Tests cross-hub signaling relay
- Reports success/failure with diagnostics

## 🔧 Service Management

```bash
# Check service status
sudo systemctl status peerpigeon

# View live logs
sudo journalctl -u peerpigeon -f

# Restart service
sudo systemctl restart peerpigeon

# Stop service
sudo systemctl stop peerpigeon
```

## 🌐 Your Endpoints (After Deployment)

### HTTP Endpoints
- Health: `http://INSTANCE_IP:3000/health`
- Stats: `http://INSTANCE_IP:3000/stats`
- Hubs: `http://INSTANCE_IP:3000/hubs`
- Hub Stats: `http://INSTANCE_IP:3000/hubstats`

### WebSocket Endpoint
- URL: `ws://INSTANCE_IP:3000/ws?peerId=<40-char-hex-id>`
- Must provide valid 40-character hex peer ID
- Example: `ws://203.0.113.12:3000/ws?peerId=0123456789abcdef0123456789abcdef01234567`

## 💰 Cost
**$0.00/month** on Oracle Cloud Always Free tier
- 2 AMD VMs (1/8 OCPU, 1GB RAM each)
- 200 GB/month outbound data transfer
- No time limits, no credit card expiration

## 🔐 Security Recommendations

1. **Add authentication:**
   ```bash
   sudo PORT=3000 IS_HUB=true AUTH_TOKEN='your-secret-token' ./oracle-setup.sh
   ```

2. **Use TLS/WSS in production:**
   - Set up Nginx or Caddy as reverse proxy
   - Get free SSL cert with Let's Encrypt
   - See `deploy/README.md` for details

3. **Restrict CORS if serving web clients:**
   ```bash
   # Edit /etc/systemd/system/peerpigeon.service
   Environment="CORS_ORIGIN=https://yourdomain.com"
   ```

## 📚 Next Steps

1. **Deploy:** Follow steps above to create your two Oracle instances
2. **Test:** Run interop tests to verify cross-hub discovery
3. **Integrate:** Connect your WebRTC application to the deployed hubs
4. **Monitor:** Use health/stats endpoints for monitoring
5. **Secure:** Add authentication and TLS for production

## 🆘 Troubleshooting

### Can't connect to server
- Check Oracle Security List has ingress rule for port 3000
- Verify firewall: `sudo iptables -L -n | grep 3000`
- Check service: `sudo systemctl status peerpigeon`

### Hubs not connecting
- Verify BOOTSTRAP_HUBS on second instance
- Check hubstats: `curl http://INSTANCE2_IP:3000/hubstats`
- View logs: `sudo journalctl -u peerpigeon -n 100`

### Cross-hub discovery fails
- Ensure both instances are in hub mode (IS_HUB=true)
- Verify same HUB_MESH_NAMESPACE on both
- Check both hubs can reach each other (network connectivity)

---

**All files are ready for deployment!** Commit and push to GitHub, then follow the deployment steps above.
