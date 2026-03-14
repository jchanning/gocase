# Oracle Cloud Infrastructure Deployment Checklist

**Project**: GoCaSE  
**Target**: OCI Free Tier (ARM64)  
**Date Started**: January 25, 2026  
**Status**: Planning Phase

---

## Current Status Assessment

### Existing OCI VM Analysis
**Run these commands on your existing OCI instance and fill in the results:**

```bash
# CPU & Memory allocation
nproc                    # Result: _____ cores
free -h                  # Total RAM: _____ GB, Used: _____ GB

# Current resource usage
top -bn1 | head -20      # CPU%: _____, Memory%: _____

# Docker resources (if applicable)
docker stats --no-stream # Docker using: _____ RAM

# Instance shape
curl -s -H "Authorization: Bearer Oracle" -L http://169.254.169.254/opc/v2/instance/ | grep shape
```

**Your Existing VM Details:**
- [ ] OS Distribution & Version: _______________________________
- [ ] OCI Shape: _______________________________
- [ ] Allocated OCPUs: _____ (of 4 total free tier)
- [ ] Allocated RAM: _____ GB (of 24 GB total free tier)
- [ ] Current CPU Usage: _____%
- [ ] Current RAM Usage: _____ GB / _____%
- [ ] Remaining Free Tier: _____ OCPU, _____ GB RAM

**Current Services Running:**
- [ ] Service 1: _______________________________ (RAM: _____ GB)
- [ ] Service 2: _______________________________ (RAM: _____ GB)
- [ ] Service 3: _______________________________ (RAM: _____ GB)
- [ ] Docker installed: YES / NO
- [ ] Nginx/web server installed: YES / NO

---

## Deployment Decision

**Based on your analysis above, choose ONE:**

### Option A: ✅ Deploy on Existing VM
- [ ] Selected
- **Reason**: _________________________________________________
- **Available resources**: _____ OCPU, _____ GB RAM for GoCaSE
- **Potential concerns**: _____________________________________

### Option B: ✅ Create New Separate VM
- [ ] Selected  
- **Reason**: _________________________________________________
- **New VM allocation**: _____ OCPU, _____ GB RAM
- **Remaining free tier after**: _____ OCPU, _____ GB RAM

---

## Infrastructure Details

### 1. OCI Instance Information
- [ ] Public IP Address: _______________________________
- [ ] SSH Key Name: _______________________________
- [ ] SSH Key Path (local): _______________________________
- [ ] OCI Username: _______________________________ (default: usually `ubuntu` or `opc`)
- [ ] Region: _______________________________ (e.g., us-ashburn-1)
- [ ] Availability Domain: _______________________________

### 2. Network & Security
- [ ] VCN Name: _______________________________
- [ ] Subnet: Public / Private
- [ ] Security List configured: YES / NO
- [ ] Ports currently open: _______________________________
- [ ] Need to open ports: 80 (HTTP), 443 (HTTPS), 22 (SSH)

### 3. Domain & SSL Configuration
- [ ] **Will use custom domain**: YES / NO
  - If YES:
    - [ ] Domain name: _______________________________
    - [ ] DNS provider: _______________________________
    - [ ] DNS A record pointing to OCI IP: DONE / TODO
- [ ] **SSL Certificate preference**:
  - [ ] Let's Encrypt (automatic, recommended)
  - [ ] Custom certificate (provide paths below)
  - [ ] No SSL (HTTP only - not recommended for production)

### 4. Database Configuration
- [ ] Production PostgreSQL password: _______________________________ (generate secure one)
- [ ] Database name: `gocase` (default) or _______________________________
- [ ] Backup frequency:
  - [ ] Daily at ___:___ (e.g., 02:00 AM)
  - [ ] Weekly on _______ at ___:___
  - [ ] Other: _______________________________
- [ ] Backup retention: _____ days
- [ ] Backup storage location: _______________________________

### 5. Application Settings
- [ ] Production port configuration:
  - [ ] Port 80/443 (nginx) → 8080 (app) - **Recommended**
  - [ ] Direct port 8080 (no nginx)
  - [ ] Custom: _______________________________
- [ ] Session secret key (will be generated): _______________________________
- [ ] Admin email for notifications: _______________________________
- [ ] Maximum upload file size: _____ MB (default: 10 MB)
- [ ] Log retention period: _____ days (default: 30)

### 6. Monitoring & Alerts (Optional)
- [ ] Set up email alerts: YES / NO
  - [ ] Alert email: _______________________________
  - [ ] Alert on: Disk >90%, RAM >90%, App down
- [ ] Set up log aggregation: YES / NO
- [ ] External monitoring service: _______________________________ (optional)

---

## Pre-Deployment Checklist

### Local Preparation
- [ ] Latest code committed to GitHub
- [ ] All tests passing locally
- [ ] Docker build successful for ARM64
- [ ] Environment variables documented
- [ ] Sample data/tests prepared

### OCI Prerequisites
- [ ] SSH access to instance confirmed
- [ ] Sudo/root access confirmed
- [ ] Internet connectivity verified
- [ ] Adequate disk space: _____ GB available (need ~10-20 GB)
- [ ] Security list rules reviewed

### Software Installation Needed
- [ ] Docker: Installed / Need to install
- [ ] Docker Compose: Installed / Need to install
- [ ] Git: Installed / Need to install
- [ ] Nginx: Installed / Need to install
- [ ] Certbot (for SSL): Installed / Need to install
- [ ] Other: _______________________________

---

## Deployment Strategy

### Deployment Method (Choose ONE):
- [ ] **Automated**: Single script deployment (recommended)
- [ ] **Manual**: Step-by-step following guide
- [ ] **Hybrid**: Script with manual verification steps

### Deployment Timeline:
- [ ] Immediate (same session)
- [ ] Scheduled for: _______________________________ (date/time)
- [ ] Flexible / when ready

### Rollback Plan:
- [ ] Keep existing app running during deployment
- [ ] Take snapshot before deployment
- [ ] Test deployment on clone first
- [ ] Direct deployment (accept risk)

---

## Files to Generate

**Once checklist complete, generate these files:**
- [ ] `docker-compose.prod.yml` - Production Docker Compose
- [ ] `deploy.sh` - Deployment automation script
- [ ] `nginx.conf` - Reverse proxy configuration
- [ ] `.env.production` - Production environment variables
- [ ] `backup.sh` - Database backup script
- [ ] `restore.sh` - Database restore script
- [ ] `update.sh` - Application update script
- [ ] `DEPLOYMENT.md` - Complete deployment guide
- [ ] `monitoring.sh` - Health check script (optional)

---

## Notes & Special Requirements

**Additional considerations:**
```
[Add any special requirements, constraints, or notes here]




```

---

## Next Steps

1. [ ] Fill out this checklist completely
2. [ ] Run diagnostic commands on existing OCI VM
3. [ ] Make deployment decision (existing vs new VM)
4. [ ] Provide completed checklist to Copilot
5. [ ] Review generated deployment files
6. [ ] Execute deployment
7. [ ] Verify deployment successful
8. [ ] Update DNS (if using custom domain)
9. [ ] Configure SSL
10. [ ] Test all functionality
11. [ ] Set up monitoring & backups
12. [ ] Document any custom configurations

---

**Status**: ⏸️ Awaiting information gathering

**Last Updated**: _______________________________
