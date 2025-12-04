# PeerPigeon-Go Deployment Guide

Deploy PeerPigeon-Go to Oracle Cloud Always Free tier or any Linux server.

## Quick Start (Oracle Cloud)

### Prerequisites
- Oracle Cloud account with Always Free tier
- 2 Ubuntu instances created (for hub setup)
- SSH access to both instances

### Deploy Bootstrap Hub (Instance 1)

```bash
# SSH into first instance
ssh ubuntu@instance1-ip

# Download and run setup script
curl -o oracle-setup.sh https://raw.githubusercontent.com/draeder/PeerPigeon-Go/main/deploy/oracle-setup.sh
chmod +x oracle-setup.sh

# Deploy as bootstrap hub
sudo PORT=3000 IS_HUB=true ./oracle-setup.sh
```

### Deploy Secondary Hub (Instance 2)

```bash
# SSH into second instance
ssh ubuntu@instance2-ip

# Download setup script
curl -o oracle-setup.sh https://raw.githubusercontent.com/draeder/PeerPigeon-Go/main/deploy/oracle-setup.sh
chmod +x oracle-setup.sh

# Deploy connecting to bootstrap (replace with instance1's public IP)
sudo PORT=3000 IS_HUB=true BOOTSTRAP_HUBS='ws://INSTANCE1_PUBLIC_IP:3000' ./oracle-setup.sh
```

### Configure Oracle Cloud Security Lists

**Important:** Oracle Cloud blocks all ports by default. You must add ingress rules:

1. Go to Oracle Cloud Console
2. Navigate to **Networking → Virtual Cloud Networks**
3. Select your VCN → **Security Lists**
4. Click your security list → **Add Ingress Rules**
5. Add rule:
   - **Source CIDR:** `0.0.0.0/0`
   - **IP Protocol:** `TCP`
   - **Destination Port Range:** `3000`
   - **Description:** `PeerPigeon WebSocket`

Repeat for both instances.

## Verification

### Check Service Status

```bash
# On each instance
sudo systemctl status peerpigeon
sudo journalctl -u peerpigeon -f
```

### Test Endpoints

```bash
# Replace with your public IPs
curl http://INSTANCE_IP:3000/health
curl http://INSTANCE_IP:3000/stats
curl http://INSTANCE_IP:3000/hubs
curl http://INSTANCE_IP:3000/hubstats
```

### Test WebSocket Connection

Install wscat:
```bash
npm install -g wscat
```

Connect to hub:
```bash
wscat -c "ws://INSTANCE_IP:3000/ws?peerId=0123456789abcdef0123456789abcdef01234567"
```

Send announce message:
```json
{"type":"announce","networkName":"global","data":{"isHub":false}}
```

## Configuration Options

The setup script accepts environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server port |
| `IS_HUB` | `false` | Enable hub mode |
| `BOOTSTRAP_HUBS` | empty | Comma-separated hub URLs |
| `AUTH_TOKEN` | empty | Optional authentication |
| `HUB_MESH_NAMESPACE` | `pigeonhub-mesh` | Hub network namespace |

### Examples

**Regular signaling server:**
```bash
sudo PORT=3000 ./oracle-setup.sh
```

**Bootstrap hub:**
```bash
sudo PORT=3000 IS_HUB=true ./oracle-setup.sh
```

**Secondary hub with auth:**
```bash
sudo PORT=3000 IS_HUB=true \
  BOOTSTRAP_HUBS='ws://hub1:3000,ws://hub2:3000' \
  AUTH_TOKEN='my-secret-token' \
  ./oracle-setup.sh
```

## Service Management

```bash
# Start service
sudo systemctl start peerpigeon

# Stop service
sudo systemctl stop peerpigeon

# Restart service
sudo systemctl restart peerpigeon

# View logs
sudo journalctl -u peerpigeon -f

# View recent errors
sudo journalctl -u peerpigeon -n 100 --no-pager
```

## Update Deployment

```bash
# SSH into instance
cd /opt/peerpigeon/PeerPigeon-Go
sudo git pull
sudo go build -o /opt/peerpigeon/peerpigeon ./cmd/peerpigeon
sudo systemctl restart peerpigeon
```

## Docker Deployment (Alternative)

Build and run with Docker:

```bash
# Build image
docker build -t peerpigeon .

# Run bootstrap hub
docker run -d \
  --name peerpigeon-hub1 \
  -p 3000:3000 \
  -e IS_HUB=true \
  peerpigeon

# Run secondary hub
docker run -d \
  --name peerpigeon-hub2 \
  -p 3001:3000 \
  -e PORT=3000 \
  -e IS_HUB=true \
  -e BOOTSTRAP_HUBS='ws://hub1-ip:3000' \
  peerpigeon
```

## Troubleshooting

### Service won't start
```bash
# Check logs
sudo journalctl -u peerpigeon -n 50

# Check binary
/opt/peerpigeon/peerpigeon --help

# Test manually
sudo -u peerpigeon /opt/peerpigeon/peerpigeon
```

### Can't connect from outside
1. Check Oracle Cloud Security List (most common issue)
2. Check UFW/iptables:
   ```bash
   sudo ufw status
   sudo iptables -L -n | grep 3000
   ```
3. Verify service is listening:
   ```bash
   sudo ss -tuln | grep 3000
   ```

### Hubs not connecting
1. Check bootstrap hub is running:
   ```bash
   curl http://BOOTSTRAP_IP:3000/health
   ```
2. Verify BOOTSTRAP_HUBS environment variable:
   ```bash
   sudo systemctl cat peerpigeon | grep BOOTSTRAP
   ```
3. Check hub stats:
   ```bash
   curl http://INSTANCE_IP:3000/hubstats
   ```

### High memory usage
Reduce MAX_CONNECTIONS in service file:
```bash
sudo systemctl edit peerpigeon
# Add: Environment="MAX_CONNECTIONS=500"
sudo systemctl restart peerpigeon
```

## Security Recommendations

1. **Enable authentication:**
   ```bash
   sudo PORT=3000 AUTH_TOKEN='generate-strong-token-here' ./oracle-setup.sh
   ```

2. **Restrict CORS** (if serving web clients):
   ```bash
   # Edit /etc/systemd/system/peerpigeon.service
   Environment="CORS_ORIGIN=https://yourdomain.com"
   ```

3. **Use a reverse proxy** (Nginx/Caddy) for TLS:
   ```bash
   sudo apt install nginx certbot python3-certbot-nginx
   # Configure Nginx to proxy WebSocket to port 3000
   sudo certbot --nginx -d your-domain.com
   ```

4. **Firewall rules:** Only allow required ports
   ```bash
   sudo ufw default deny incoming
   sudo ufw allow ssh
   sudo ufw allow 3000/tcp
   sudo ufw enable
   ```

## Cost

Oracle Cloud Always Free tier includes:
- 2 AMD VMs (1/8 OCPU, 1GB RAM each) - sufficient for PeerPigeon hubs
- 200 GB/month outbound data transfer
- **Cost: $0.00/month forever**

## Support

- GitHub Issues: https://github.com/draeder/PeerPigeon-Go/issues
- Original JS Project: https://github.com/PeerPigeon/PeerPigeon
