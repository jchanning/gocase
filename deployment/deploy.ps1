# deploy.ps1 — Deploy GoCaSE to Oracle Cloud ARM instance (gocase.fistraltech.com)
# Usage: .\deployment\deploy.ps1
#   Optional flags:
#     -Setup      — uploads and runs setup-server.sh (run ONCE before first deploy)
#     -DBOnly     — only restarts the database container
#
# Prerequisites (one-time):
#   1. DNS A record: gocase.fistraltech.com → 132.145.64.140
#   2. Run with -Setup flag to install Docker, nginx config, and SSL cert
#   3. Create /home/opc/gocase/.env.production on the server
#      (copy from deployment/env.production.example and fill in values)
#   4. Upload your OCI API key:
#      scp -i C:\Users\johnm\.ssh\oracle-wordai.key \path\to\oci_api_key.pem opc@132.145.64.140:/home/opc/gocase-data/oci_api_key.pem

param(
    [switch]$Setup,
    [switch]$DBOnly
)

$ErrorActionPreference = "Stop"

# ── Connection details ────────────────────────────────────────────────
$IP         = "132.145.64.140"
$KEY        = "C:\Users\johnm\.ssh\oracle-wordai.key"
$USER       = "opc"
$REMOTE_DIR = "/home/$USER/gocase"
$DATA_DIR   = "/home/$USER/gocase-data"

# ── Helpers ───────────────────────────────────────────────────────────
function Invoke-SSH {
    param([string]$Command)
    Write-Host "  > $Command" -ForegroundColor DarkGray
    ssh -i $KEY -o StrictHostKeyChecking=no "${USER}@${IP}" $Command
    if ($LASTEXITCODE -ne 0) { throw "SSH command failed: $Command" }
}

function Invoke-SCP {
    param([string]$Source, [string]$Dest)
    Write-Host "  SCP: $Source -> $Dest" -ForegroundColor DarkGray
    scp -i $KEY -o StrictHostKeyChecking=no -r $Source "${USER}@${IP}:${Dest}"
    if ($LASTEXITCODE -ne 0) { throw "SCP failed: $Source -> $Dest" }
}

# ── Step 0: One-time server setup ────────────────────────────────────
if ($Setup) {
    Write-Host "`n=== Running one-time server setup ===" -ForegroundColor Cyan
    $SetupScript = Join-Path $PSScriptRoot "setup-server.sh"
    if (-not (Test-Path $SetupScript)) {
        throw "setup-server.sh not found at $SetupScript"
    }
    $AdminEmail = Read-Host "Enter admin email for SSL certificate"
    Invoke-SCP $SetupScript "/tmp/setup-server.sh"
    Invoke-SSH "bash /tmp/setup-server.sh '$AdminEmail'"
    Write-Host "`nOne-time setup complete. Now:" -ForegroundColor Green
    Write-Host "  1. Create /home/opc/gocase/.env.production (see deployment/env.production.example)" -ForegroundColor Yellow
    Write-Host "  2. Upload your OCI API key to: $DATA_DIR/oci_api_key.pem" -ForegroundColor Yellow
    Write-Host "  3. Re-run deploy.ps1 (without -Setup) to deploy the application." -ForegroundColor Yellow
    exit 0
}

# ── Step 1: Verify env file exists on server ─────────────────────────
Write-Host "`n=== Checking production environment file ===" -ForegroundColor Cyan
$envCheck = ssh -i $KEY -o StrictHostKeyChecking=no "${USER}@${IP}" "test -f $REMOTE_DIR/.env.production && echo EXISTS || echo MISSING" 2>&1
if ($envCheck -match "MISSING") {
    Write-Host "  WARNING: $REMOTE_DIR/.env.production not found on server." -ForegroundColor Yellow
    Write-Host "  Create it from deployment/env.production.example before the app will work." -ForegroundColor Yellow
} else {
    Write-Host "  .env.production found." -ForegroundColor Green
}

# ── Step 2: Package the project ──────────────────────────────────────
Write-Host "`n=== Packaging project ===" -ForegroundColor Cyan
$PROJECT_ROOT = Split-Path -Parent $PSScriptRoot   # one level up from deployment/
Write-Host "  Project root: $PROJECT_ROOT"

$TMP_TAR = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "gocase-deploy.tar.gz")
Write-Host "  Creating archive: $TMP_TAR"

# Use git archive to read directly from git's object store, avoiding
# OneDrive Files-On-Demand permission issues with plain tar.
Push-Location $PROJECT_ROOT
try {
    git archive --format=tar.gz HEAD -o $TMP_TAR
    if ($LASTEXITCODE -ne 0) { throw "git archive failed" }
} finally {
    Pop-Location
}
Write-Host "  Archive created: $((Get-Item $TMP_TAR).Length / 1MB | ForEach-Object { [math]::Round($_, 1) }) MB"

# ── Step 2: Upload to server ─────────────────────────────────────────
Write-Host "`n=== Uploading to server ===" -ForegroundColor Cyan
Invoke-SSH "mkdir -p $REMOTE_DIR"
Invoke-SCP $TMP_TAR "${REMOTE_DIR}/gocase.tar.gz"
Remove-Item $TMP_TAR -Force

# Clean old files (preserve .env.production), fix permissions, then extract
Invoke-SSH "chmod -R u+rwX $REMOTE_DIR 2>/dev/null || true"
Invoke-SSH "find $REMOTE_DIR -mindepth 1 -not -name 'gocase.tar.gz' -not -name '.env.production' -delete 2>/dev/null || true"
Invoke-SSH "cd $REMOTE_DIR && tar -xzf gocase.tar.gz && rm -f gocase.tar.gz"

# Ensure uploads directory exists on server
Invoke-SSH "mkdir -p $REMOTE_DIR/uploads/notes"

# ── Step 3: Build & start containers ─────────────────────────────────
if ($DBOnly) {
    Write-Host "`n=== Restarting database only ===" -ForegroundColor Cyan
    Invoke-SSH "cd $REMOTE_DIR && docker compose -f docker-compose.prod.yml up -d db"
} else {
    Write-Host "`n=== Building and starting all services ===" -ForegroundColor Cyan
    Invoke-SSH "cd $REMOTE_DIR && docker compose -f docker-compose.prod.yml build --no-cache app"
    Invoke-SSH "cd $REMOTE_DIR && docker compose -f docker-compose.prod.yml --env-file .env.production up -d"
}

# ── Step 4: Health check ─────────────────────────────────────────────
Write-Host "`n=== Waiting for services ===" -ForegroundColor Cyan
Start-Sleep -Seconds 15
Invoke-SSH "cd $REMOTE_DIR && docker compose -f docker-compose.prod.yml ps"

Write-Host "`n=== Checking app health ===" -ForegroundColor Cyan
Invoke-SSH "curl -sf --max-time 10 http://127.0.0.1:8081/ > /dev/null && echo 'App: OK' || echo 'App: NOT READY - check: docker logs gocase-app'"

# ── Done ──────────────────────────────────────────────────────────────
Write-Host "`n=== Deployment complete ===" -ForegroundColor Green
Write-Host @"

  Application: https://gocase.fistraltech.com

  Useful commands (run on the server):
    docker logs gocase-app -f          # live app logs
    docker logs gocase-db  -f          # live db logs
    docker compose -f docker-compose.prod.yml ps    # container status
    sudo journalctl -u gocase -f       # systemd service logs

  OCI Budget Alert (do this once if not done):
    OCI Console > Billing & Cost Management > Budgets > Create Budget
    Set a monthly threshold (e.g. 10 GBP) with email alerts at 80% and 100%.

"@ -ForegroundColor Yellow
