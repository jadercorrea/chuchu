
# SysOps

Guidelines for system administration, shell scripting, and infrastructure management.

## When to Activate

- Writing shell scripts
- Managing Linux systems
- Configuring services
- Performance tuning

## Shell Scripting

### Script template

```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# Description: What this script does
# Usage: ./script.sh <arg1> <arg2>

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="/var/log/$(basename "$0" .sh).log"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

error() {
    log "ERROR: $*" >&2
    exit 1
}

cleanup() {
    log "Cleaning up..."
    # Remove temp files, etc.
}
trap cleanup EXIT

main() {
    log "Starting script..."
    # Your logic here
    log "Script completed successfully"
}

main "$@"
```

### Input validation

```bash
# Require arguments
if [[ $# -lt 2 ]]; then
    echo "Usage: $0 <source> <destination>"
    exit 1
fi

# Validate file exists
if [[ ! -f "$1" ]]; then
    error "Source file does not exist: $1"
fi

# Validate directory
if [[ ! -d "$2" ]]; then
    error "Destination directory does not exist: $2"
fi
```

### Safe variable handling

```bash
# GOOD - quote variables
cp "$source" "$destination"
rm -rf "${temp_dir:?}/"

# BAD - unquoted (word splitting issues)
cp $source $destination
rm -rf $temp_dir/  # Dangerous!

# Default values
name="${1:-default_value}"
config_file="${CONFIG_FILE:-/etc/app/config.yaml}"
```

## Systemd Services

### Service unit file

```ini
# /etc/systemd/system/myapp.service
[Unit]
Description=My Application
Documentation=https://docs.example.com
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=appuser
Group=appgroup
WorkingDirectory=/opt/myapp

Environment=NODE_ENV=production
EnvironmentFile=/etc/myapp/env

ExecStart=/usr/bin/node /opt/myapp/server.js
ExecReload=/bin/kill -HUP $MAINPID

Restart=always
RestartSec=5
StartLimitIntervalSec=60
StartLimitBurst=3

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/lib/myapp

[Install]
WantedBy=multi-user.target
```

### Common systemctl commands

```bash
# Manage services
sudo systemctl start myapp
sudo systemctl stop myapp
sudo systemctl restart myapp
sudo systemctl reload myapp

# Enable/disable auto-start
sudo systemctl enable myapp
sudo systemctl disable myapp

# Check status
systemctl status myapp
journalctl -u myapp -f
```

## Log Management

### Logrotate configuration

```
# /etc/logrotate.d/myapp
/var/log/myapp/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 appuser appgroup
    sharedscripts
    postrotate
        systemctl reload myapp > /dev/null 2>&1 || true
    endscript
}
```

### Log analysis

```bash
# Find errors in logs
grep -E "ERROR|FATAL" /var/log/myapp/*.log

# Count requests per status code
awk '{print $9}' /var/log/nginx/access.log | sort | uniq -c | sort -rn

# Find slow requests (>1s)
awk '$NF > 1.0 {print}' /var/log/nginx/access.log

# Real-time monitoring
tail -f /var/log/myapp/app.log | grep --line-buffered ERROR
```

## Performance Monitoring

### System metrics

```bash
# CPU usage
top -b -n 1 | head -20
mpstat -P ALL 1 5

# Memory usage
free -h
vmstat 1 10

# Disk I/O
iostat -x 1 5
iotop -o

# Network
ss -tuln
nethogs
iftop
```

### Process investigation

```bash
# Find process by port
lsof -i :3000
ss -tlnp | grep 3000

# Process details
ps aux | grep myapp
pstree -p $(pgrep myapp)

# Open files by process
lsof -p $(pgrep myapp)

# System calls
strace -p $(pgrep myapp) -f -e trace=network
```

## Disk Management

### Check disk space

```bash
# Human-readable disk usage
df -h

# Find large files
find /var -type f -size +100M -exec ls -lh {} \;

# Directory sizes
du -sh /var/* | sort -rh | head -20

# Disk usage by filetype
find /var -type f -name "*.log" -exec du -ch {} + | tail -1
```

### Safe cleanup

```bash
# Clear old logs (keep 7 days)
find /var/log -name "*.log" -mtime +7 -delete

# Clear old temp files
find /tmp -type f -atime +3 -delete

# Clear package cache (Debian/Ubuntu)
apt-get clean
apt-get autoremove

# Clear journal logs older than 7 days
journalctl --vacuum-time=7d
```

## Cron Jobs

### Crontab best practices

```cron
# /etc/cron.d/myapp

# Always set PATH
PATH=/usr/local/bin:/usr/bin:/bin

# Always set MAILTO
MAILTO=ops@example.com

# Use descriptive comments
# Backup database daily at 2 AM
0 2 * * * appuser /opt/myapp/scripts/backup.sh >> /var/log/myapp/backup.log 2>&1

# Health check every 5 minutes
*/5 * * * * appuser /opt/myapp/scripts/healthcheck.sh > /dev/null 2>&1

# Cleanup weekly on Sunday at 3 AM
0 3 * * 0 appuser /opt/myapp/scripts/cleanup.sh >> /var/log/myapp/cleanup.log 2>&1
```

### Lock to prevent overlap

```bash
#!/usr/bin/env bash
# Use flock to prevent concurrent runs
exec 200>/var/lock/myapp-backup.lock
flock -n 200 || { echo "Already running"; exit 1; }

# Your backup logic here
```

## SSH Configuration

### Secure SSH config

```
# /etc/ssh/sshd_config
Port 22
Protocol 2
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
PermitEmptyPasswords no
ChallengeResponseAuthentication no
UsePAM yes
X11Forwarding no
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server
MaxAuthTries 3
LoginGraceTime 20
AllowUsers deploy@* admin@192.168.1.*
```

### SSH client config

```
# ~/.ssh/config
Host production
    HostName prod.example.com
    User deploy
    IdentityFile ~/.ssh/prod_key
    ForwardAgent no

Host staging
    HostName staging.example.com
    User deploy
    IdentityFile ~/.ssh/staging_key
    
Host *
    ServerAliveInterval 60
    ServerAliveCountMax 3
    AddKeysToAgent yes
```

## Firewall

### UFW (Ubuntu)

```bash
# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow specific services
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp

# Allow from specific IP
ufw allow from 192.168.1.0/24 to any port 5432

# Enable firewall
ufw enable
ufw status verbose
```
