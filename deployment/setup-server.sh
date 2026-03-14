#!/bin/bash
# setup-server.sh — One-time server setup for GoCaSE on Oracle Linux 9.7
#
# Run this ONCE on the OCI instance before the first deployment.
# Usage:
#   scp -i ~/.ssh/oracle-wordai.key deployment/setup-server.sh opc@132.145.64.140:/tmp/
#   ssh -i ~/.ssh/oracle-wordai.key opc@132.145.64.140 "bash /tmp/setup-server.sh your@email.com"
#
# Prerequisites:
#   - DNS A record for gocase.fistraltech.com pointing to 132.145.64.140 must exist
#     and have propagated before certbot is run (check: nslookup gocase.fistraltech.com)

set -euo pipefail

DOMAIN="gocase.fistraltech.com"
APP_PORT="8081"
DEPLOY_DIR="/home/opc/gocase"
DATA_DIR="/home/opc/gocase-data"
ADMIN_EMAIL="${1:-}"

# ── Validate ──────────────────────────────────────────────────────────
if [ -z "$ADMIN_EMAIL" ]; then
    echo "ERROR: Admin email required."
    echo "Usage: bash setup-server.sh your@email.com"
    exit 1
fi

echo ""
echo "======================================================"
echo "  GoCaSE Server Setup"
echo "  Domain  : $DOMAIN"
echo "  App port: $APP_PORT (internal, nginx proxies)"
echo "  Email   : $ADMIN_EMAIL"
echo "======================================================"
echo ""

# ── Step 1: Install Docker CE ─────────────────────────────────────────
echo "=== [1/7] Installing Docker CE ==="
if command -v docker &>/dev/null; then
    echo "  Docker already installed: $(docker --version)"
else
    sudo dnf config-manager --add-repo https://download.docker.com/linux/rhel/docker-ce.repo
    sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    sudo systemctl enable --now docker
    sudo usermod -aG docker opc
    echo "  Docker installed. The 'docker' group change takes effect after re-login."
    echo "  For this session, using 'newgrp docker' or 'sudo docker' will be needed."
fi

# ── Step 2: Open firewall ports ───────────────────────────────────────
echo ""
echo "=== [2/7] Configuring firewall ==="
sudo firewall-cmd --permanent --add-service=http  2>/dev/null || true
sudo firewall-cmd --permanent --add-service=https 2>/dev/null || true
sudo firewall-cmd --reload 2>/dev/null || true
echo "  Ports 80 (HTTP) and 443 (HTTPS) opened."
echo "  NOTE: Also open these in the OCI Security List in the OCI Console:"
echo "  Networking > Virtual Cloud Networks > your-VCN > Security Lists > Add Ingress Rules"
echo "  Ingress: TCP port 80 from 0.0.0.0/0"
echo "  Ingress: TCP port 443 from 0.0.0.0/0"

# ── Step 3: Install certbot ───────────────────────────────────────────
echo ""
echo "=== [3/7] Installing certbot ==="
if command -v certbot &>/dev/null; then
    echo "  certbot already installed."
else
    sudo dnf install -y certbot python3-certbot-nginx
    echo "  certbot installed."
fi

# ── Step 4: Create data directories ──────────────────────────────────
echo ""
echo "=== [4/7] Creating data directories ==="
mkdir -p "$DATA_DIR/uploads/notes"
chmod 755 "$DATA_DIR"
echo "  Persistent data directory: $DATA_DIR"
echo "  Place your OCI API private key at: $DATA_DIR/oci_api_key.pem"

# ── Step 5: Install nginx config ─────────────────────────────────────
echo ""
echo "=== [5/7] Installing nginx config for $DOMAIN ==="
# Write initial HTTP-only config so certbot can complete its challenge
sudo tee /etc/nginx/conf.d/gocase.conf > /dev/null <<'NGINXEOF'
server {
    listen 80;
    listen [::]:80;
    server_name gocase.fistraltech.com;
    location / {
        return 301 https://$server_name$request_uri;
    }
}
NGINXEOF
sudo nginx -t && sudo systemctl reload nginx || { echo "  WARNING: nginx config test failed — check /etc/nginx/conf.d/gocase.conf"; }
echo "  Temporary HTTP config installed."

# ── Step 6: Obtain SSL certificate ───────────────────────────────────
echo ""
echo "=== [6/7] Obtaining SSL certificate for $DOMAIN ==="
echo "  Checking DNS resolution..."
RESOLVED_IP=$(dig +short "$DOMAIN" 2>/dev/null | tail -1 || nslookup "$DOMAIN" 2>/dev/null | awk '/^Address: /{print $2}' | tail -1 || echo "")
SERVER_IP=$(curl -4 -s --max-time 5 ifconfig.me 2>/dev/null || echo "unknown")

if [ "$RESOLVED_IP" != "$SERVER_IP" ]; then
    echo ""
    echo "  WARNING: DNS not yet propagated."
    echo "  $DOMAIN resolves to: ${RESOLVED_IP:-<not found>}"
    echo "  This server's IP is: $SERVER_IP"
    echo ""
    echo "  Skipping SSL certificate for now."
    echo "  Once DNS propagates, run this on the server to get the cert:"
    echo "    sudo certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email $ADMIN_EMAIL"
    SKIP_SSL=true
else
    echo "  DNS OK: $DOMAIN → $RESOLVED_IP"
    CERTBOT_BIN=$(command -v certbot 2>/dev/null || echo "/usr/local/bin/certbot")
    sudo "$CERTBOT_BIN" --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "$ADMIN_EMAIL"
    echo "  SSL certificate issued and nginx updated."
    SKIP_SSL=false
fi

# ── Step 7: Install systemd service ──────────────────────────────────
echo ""
echo "=== [7/7] Installing systemd auto-start service ==="
sudo tee /etc/systemd/system/gocase.service > /dev/null <<SVCEOF
[Unit]
Description=GoCaSE Application
Documentation=https://github.com/fistraltech/gocase
Requires=docker.service
After=docker.service network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$DEPLOY_DIR
EnvironmentFile=$DEPLOY_DIR/.env.production
ExecStart=/usr/bin/docker compose -f docker-compose.prod.yml up -d
ExecStop=/usr/bin/docker compose -f docker-compose.prod.yml down
TimeoutStartSec=120

[Install]
WantedBy=multi-user.target
SVCEOF

sudo systemctl daemon-reload
sudo systemctl enable gocase
echo "  systemd service 'gocase' installed and enabled (starts on boot)."

# ── Summary ───────────────────────────────────────────────────────────
echo ""
echo "======================================================"
echo "  Setup complete!"
echo "======================================================"
echo ""
echo "  NEXT STEPS:"
echo ""
echo "  1. OCI Console — add Security List ingress rules for TCP 80 and 443"
echo "     if you haven't already (ports are open in the OS firewall now)."
echo ""
echo "  2. Create the production env file on this server:"
echo "     nano $DEPLOY_DIR/.env.production"
echo "     (copy from env.production.example and fill in real values)"
echo ""
echo "  3. Upload your OCI API private key:"
echo "     scp -i oracle-wordai.key /path/to/oci_api_key.pem opc@$SERVER_IP:$DATA_DIR/oci_api_key.pem"
echo ""
if [ "${SKIP_SSL:-false}" = "true" ]; then
echo "  4. Fix DNS, then run to get SSL cert:"
echo "     sudo certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email $ADMIN_EMAIL"
echo ""
fi
echo "  5. From your Windows machine, run the deploy script:"
echo "     .\\deployment\\deploy.ps1"
echo ""
echo "  The app will be available at: https://$DOMAIN"
echo ""
