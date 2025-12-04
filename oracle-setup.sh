#!/bin/bash
set -e

# PeerPigeon-Go Oracle Cloud Deployment Script
# This script sets up and deploys PeerPigeon-Go on Oracle Cloud Always Free instances

echo "🚀 PeerPigeon-Go Oracle Cloud Setup"
echo "===================================="

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/peerpigeon"
SERVICE_NAME="peerpigeon"
USER="peerpigeon"

# Parse arguments
PORT="${PORT:-3000}"
IS_HUB="${IS_HUB:-false}"
BOOTSTRAP_HUBS="${BOOTSTRAP_HUBS:-}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
HUB_MESH_NAMESPACE="${HUB_MESH_NAMESPACE:-pigeonhub-mesh}"

show_usage() {
    echo "Usage: sudo ./oracle-setup.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  PORT=3000                    Port to listen on (default: 3000)"
    echo "  IS_HUB=true                  Enable hub mode (default: false)"
    echo "  BOOTSTRAP_HUBS='ws://...'    Comma-separated bootstrap hub URLs"
    echo "  AUTH_TOKEN='token'           Optional authentication token"
    echo "  HUB_MESH_NAMESPACE='ns'      Hub mesh namespace (default: pigeonhub-mesh)"
    echo ""
    echo "Examples:"
    echo "  # Bootstrap hub on first instance:"
    echo "  sudo PORT=3000 IS_HUB=true ./oracle-setup.sh"
    echo ""
    echo "  # Secondary hub connecting to bootstrap:"
    echo "  sudo PORT=3000 IS_HUB=true BOOTSTRAP_HUBS='ws://instance1-ip:3000' ./oracle-setup.sh"
}

if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    show_usage
    exit 0
fi

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
   exit 1
fi

echo -e "${GREEN}📋 Configuration:${NC}"
echo "  Port: $PORT"
echo "  Hub Mode: $IS_HUB"
echo "  Bootstrap Hubs: ${BOOTSTRAP_HUBS:-none}"
echo "  Namespace: $HUB_MESH_NAMESPACE"
echo ""

# Update system
echo -e "${YELLOW}📦 Updating system packages...${NC}"
apt-get update -qq
apt-get upgrade -y -qq

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
apt-get install -y -qq git curl wget build-essential

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}📦 Installing Go 1.22...${NC}"
    GO_VERSION="1.22.0"
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
else
    echo -e "${GREEN}✅ Go already installed: $(go version)${NC}"
fi

# Create service user
if ! id "$USER" &>/dev/null; then
    echo -e "${YELLOW}👤 Creating service user: $USER${NC}"
    useradd -r -s /bin/false -d $INSTALL_DIR $USER
else
    echo -e "${GREEN}✅ User $USER already exists${NC}"
fi

# Create installation directory
echo -e "${YELLOW}📁 Setting up installation directory: $INSTALL_DIR${NC}"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# Clone or update repository
if [ -d "$INSTALL_DIR/PeerPigeon-Go" ]; then
    echo -e "${YELLOW}🔄 Updating existing repository...${NC}"
    cd PeerPigeon-Go
    git pull
else
    echo -e "${YELLOW}📥 Cloning repository...${NC}"
    git clone https://github.com/draeder/PeerPigeon-Go.git
    cd PeerPigeon-Go
fi

# Build binary
echo -e "${YELLOW}🔨 Building PeerPigeon-Go...${NC}"
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o $INSTALL_DIR/peerpigeon ./cmd/peerpigeon

# Set permissions
chown -R $USER:$USER $INSTALL_DIR
chmod +x $INSTALL_DIR/peerpigeon

# Configure firewall
echo -e "${YELLOW}🔥 Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow $PORT/tcp
    ufw --force enable
    echo -e "${GREEN}✅ UFW configured${NC}"
fi

if command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=$PORT/tcp
    firewall-cmd --reload
    echo -e "${GREEN}✅ Firewalld configured${NC}"
fi

# Open Oracle Cloud firewall (iptables)
echo -e "${YELLOW}🔓 Opening Oracle Cloud fireables...${NC}"
iptables -I INPUT 6 -m state --state NEW -p tcp --dport $PORT -j ACCEPT
netfilter-persistent save || iptables-save > /etc/iptables/rules.v4

# Create systemd service
echo -e "${YELLOW}⚙️  Creating systemd service...${NC}"
cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=PeerPigeon WebSocket Signaling Server
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/peerpigeon
Restart=always
RestartSec=10
Environment="PORT=$PORT"
Environment="HOST=0.0.0.0"
Environment="MAX_CONNECTIONS=1000"
Environment="CORS_ORIGIN=*"
Environment="IS_HUB=$IS_HUB"
Environment="HUB_MESH_NAMESPACE=$HUB_MESH_NAMESPACE"
Environment="BOOTSTRAP_HUBS=$BOOTSTRAP_HUBS"
Environment="AUTH_TOKEN=$AUTH_TOKEN"

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
echo -e "${YELLOW}🔄 Starting service...${NC}"
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl restart $SERVICE_NAME

# Wait for service to start
sleep 3

# Check service status
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${GREEN}✅ Service started successfully!${NC}"
    systemctl status $SERVICE_NAME --no-pager -l
else
    echo -e "${RED}❌ Service failed to start${NC}"
    journalctl -u $SERVICE_NAME -n 50 --no-pager
    exit 1
fi

# Get public IP
echo ""
echo -e "${YELLOW}🌐 Detecting public IP...${NC}"
PUBLIC_IP=$(curl -s ifconfig.me || curl -s icanhazip.com || echo "unknown")

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Installation Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}Service Status:${NC}"
echo "  Command: sudo systemctl status $SERVICE_NAME"
echo "  Logs: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo -e "${GREEN}Endpoints:${NC}"
echo "  Health: http://$PUBLIC_IP:$PORT/health"
echo "  Stats: http://$PUBLIC_IP:$PORT/stats"
echo "  Hubs: http://$PUBLIC_IP:$PORT/hubs"
echo "  Hub Stats: http://$PUBLIC_IP:$PORT/hubstats"
echo "  WebSocket: ws://$PUBLIC_IP:$PORT/ws?peerId=<40-char-hex>"
echo ""
echo -e "${GREEN}Quick Test:${NC}"
echo "  curl http://$PUBLIC_IP:$PORT/health"
echo ""
echo -e "${YELLOW}⚠️  Don't forget to configure Oracle Cloud Security List:${NC}"
echo "  1. Go to Oracle Cloud Console"
echo "  2. Navigate to Networking > Virtual Cloud Networks"
echo "  3. Select your VCN > Security Lists"
echo "  4. Add Ingress Rule: TCP port $PORT from 0.0.0.0/0"
echo ""
