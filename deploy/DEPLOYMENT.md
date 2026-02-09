# Deployment Guide

This guide covers deploying clibot in production using systemd or supervisor.

## Platform Support

**Supported Platforms**:
- ✅ **Linux** - Fully supported and recommended for production
- ✅ **macOS** - Fully supported
- ⚠️ **Windows** - Only via WSL2 (not recommended for production)

**Windows users**:
- Use WSL2 for development/testing
- For production, deploy to a Linux server (VPS, cloud, etc.)
- See [Windows setup guide](../README.md#windows-setup-wsl2) for details

## Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment with systemd](#deployment-with-systemd)
- [Deployment with Supervisor](#deployment-with-supervisor)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## Prerequisites

1. **Install clibot**:
```bash
go install github.com/keepmind9/clibot@latest
```

**Note**: `go install` places the binary at `$GOPATH/bin/clibot` (usually `~/go/bin/clibot`).
Make sure `~/go/bin` is in your PATH:
```bash
# Add to PATH (add to ~/.bashrc or ~/.zshrc for persistence)
export PATH=$PATH:~/go/bin
```

2. **Install tmux** (required):
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux

# Fedora/CentOS/RHEL
sudo dnf install tmux
```

**Note**: You must specify the config file path with `--config`. Common options:
- `~/.config/clibot/config.yaml` - User-specific (recommended)
- `/etc/clibot/config.yaml` - System-wide (requires root)
- `./config.yaml` - Project directory

If you use a custom location, update the `--config` path in:
- systemd service file (`ExecStart=` line)
- supervisor config file (`command=` line)

3. **Configure clibot**:
```bash
# Create config directory (user-specific)
mkdir -p ~/.config/clibot

# Copy config template
cp configs/config.yaml ~/.config/clibot/config.yaml

# Edit configuration
nano ~/.config/clibot/config.yaml
```

**Alternative (system-wide)**:
```bash
# Create config directory (system-wide)
sudo mkdir -p /etc/clibot
sudo cp configs/config.yaml /etc/clibot/config.yaml
sudo nano /etc/clibot/config.yaml
```

**Important**: Fill in your bot credentials and whitelist users in the config file.

## Deployment with systemd

systemd is the init system for modern Linux distributions (Ubuntu 16.04+, CentOS 7+, etc.).

### Installation

1. **Create systemd user directory and copy the service file**:
```bash
mkdir -p ~/.config/systemd/user
cp deploy/clibot.service ~/.config/systemd/user/
```

**Customize paths** (optional):
If you're using a different config location or binary path, edit the service file:
```bash
nano ~/.config/systemd/user/clibot.service
```

Key lines to customize:
- `ExecStart=~/go/bin/clibot serve --config ~/.config/clibot/config.yaml`
  - `~/go/bin/clibot`: Binary path (change if `$GOPATH/bin` is different)
  - If `~/go/bin` is in your PATH, you can use `clibot` instead
  - `--config ~/.config/clibot/config.yaml`: Config file path (required)
- `WorkingDirectory=~/projects`
  - By default this is commented out (uses user's home directory)
  - Uncomment and change to set a specific working directory

2. **Reload systemd**:
```bash
systemctl --user daemon-reload
```

3. **Enable clibot** to start on login:
```bash
systemctl --user enable clibot
```

4. **Start clibot**:
```bash
systemctl --user start clibot
```

**Optional: Enable lingering** (start service on boot, not just login):
```bash
loginctl enable-linger $USER
```

### Management Commands

```bash
# Check status
systemctl --user status clibot

# Stop clibot
systemctl --user stop clibot

# Restart clibot
systemctl --user restart clibot

# View logs
journalctl --user -u clibot -f

# View logs since last boot
journalctl --user -u clibot -b

# View last 100 lines
journalctl --user -u clibot -n 100
```

### Log Rotation

systemd handles log rotation automatically via journald. To configure persistent logging:

```bash
# Create journal directory
sudo mkdir -p /var/log/journal

# Restart journald
sudo systemctl restart systemd-journald
```

## Deployment with Supervisor

**Note**: supervisor is a system-level service that requires root privileges to install and configure. For user-level service management, consider using [systemd user services](#deployment-with-systemd) instead.

Supervisor is a process control system for Unix-like operating systems.

### Installation

1. **Install supervisor**:
```bash
# Ubuntu/Debian
sudo apt-get install supervisor

# Fedora/CentOS/RHEL
sudo dnf install supervisor

# macOS
brew install supervisor
```

2. **Copy the configuration file**:
```bash
sudo cp deploy/clibot.conf /etc/supervisor/conf.d/clibot.conf
```

**Customize paths** (optional):
If you're using a different config location or binary path, edit the config file:
```bash
sudo nano /etc/supervisor/conf.d/clibot.conf
```

Key lines to customize:
- `command=/usr/local/bin/clibot serve`
  - By default uses `~/.config/clibot/config.yaml`
  - Add `--config /path/to/config.yaml` for custom location
  - Change binary path if installed elsewhere
- `user=clibot`
  - By default this is commented out (runs as current user)
  - Uncomment to run as a dedicated user
- `stdout_logfile=~/clibot.log`
  - By default uses user's home directory
  - Change to `/var/log/clibot/clibot.log` for system-wide logs

3. **Reread and update supervisor**:
```bash
sudo supervisorctl reread
sudo supervisorctl update
```

4. **Start clibot**:
```bash
sudo supervisorctl start clibot
```

### Management Commands

```bash
# Check status
sudo supervisorctl status clibot

# Stop clibot
sudo supervisorctl stop clibot

# Restart clibot
sudo supervisorctl restart clibot

# View logs
sudo supervisorctl tail -f clibot

# View stderr logs
sudo supervisorctl tail -f clibot stderr

# View stdout logs
sudo supervisorctl tail -f clibot stdout
```

### Log Rotation

Supervisor handles log rotation automatically based on the settings in `clibot.conf`:
- `stdout_logfile_maxbytes=50MB` - Rotate at 50MB
- `stdout_logfile_backups=10` - Keep 10 backup files

## Verification

### Verify clibot is running

```bash
# Check if tmux sessions exist
tmux list-sessions

# If running as dedicated user:
# sudo -u clibot tmux list-sessions

# Check if clibot is listening on port 8080
sudo netstat -tlnp | grep 8080

# Or using ss
sudo ss -tlnp | grep 8080
```

### Test from IM

Send a message to your bot:
```
status
```

You should receive a status response.

## Troubleshooting

### clibot won't start

1. **Check the service status**:
```bash
# systemd
sudo systemctl status clibot

# supervisor
sudo supervisorctl status clibot
```

2. **Check the logs**:
```bash
# systemd
sudo journalctl -u clibot -n 100

# supervisor
sudo tail -100 /var/log/clibot/clibot.log
```

3. **Common issues**:

   **Issue**: Permission denied
   **Solution**:
   ```bash
   # If running as current user:
   mkdir -p ~/.config/clibot
   chmod 700 ~/.config/clibot

   # If running as dedicated user:
   # sudo chown -R clibot:clibot /etc/clibot
   # sudo chown -R clibot:clibot /var/log/clibot
   ```

   **Issue**: Config file not found
   **Solution**:
   ```bash
   # If using user-specific config:
   mkdir -p ~/.config/clibot
   cp configs/config.yaml ~/.config/clibot/config.yaml

   # If using system-wide config:
   # Ensure /etc/clibot/config.yaml exists and is readable
   ```

   **Issue**: Port already in use
   **Solution**:
   ```bash
   # Find process using port 8080
   sudo lsof -i :8080
   # Change port in config.yaml
   ```

   **Issue**: tmux not found
   **Solution**:
   ```bash
   # Install tmux
   sudo apt-get install tmux  # Ubuntu/Debian
   ```

### Manual testing

Run clibot manually to debug issues:

```bash
# Run as current user (default)
/usr/local/bin/clibot serve

# Or with config specified
/usr/local/bin/clibot serve --config ~/.config/clibot/config.yaml

# Or run with debug logging
/usr/local/bin/clibot serve --config ~/.config/clibot/config.yaml --log-level debug

# If running as dedicated user:
# sudo -u clibot /usr/local/bin/clibot serve --config /etc/clibot/config.yaml
```

## Uninstallation

### systemd

```bash
# Stop and disable
systemctl --user stop clibot
systemctl --user disable clibot

# Remove service file
rm ~/.config/systemd/user/clibot.service
systemctl --user daemon-reload

# Remove user config (optional)
rm -rf ~/.config/clibot
```

### Supervisor

```bash
# Stop clibot
sudo supervisorctl stop clibot

# Remove config
sudo rm /etc/supervisor/conf.d/clibot.conf

# Reread and update
sudo supervisorctl reread
sudo supervisorctl update

# Remove user config (optional)
rm -rf ~/.config/clibot

# If running as dedicated user, remove user and directories:
# sudo userdel clibot
# sudo rm -rf /etc/clibot
# sudo rm -rf /var/log/clibot
```

## Production Tips

### Security

1. **Enable whitelist** in config.yaml:
```yaml
security:
  whitelist_enabled: true
  allowed_users:
    discord:
      - "your-user-id"
```

2. **Use environment variables for secrets**:
```yaml
bots:
  discord:
    token: "${DISCORD_TOKEN}"
```

Set via:
```bash
# For systemd user service
systemctl --user edit clibot
# Add: [Service] Environment=DISCORD_TOKEN=your_token

# For supervisor
sudo nano /etc/supervisor/conf.d/clibot.conf
# Add: environment=DISCORD_TOKEN="your_token"
```

### Performance

1. **Limit dynamic sessions**:
```yaml
session:
  max_dynamic_sessions: 50
```

2. **Adjust log levels** for production:
```yaml
logging:
  level: info  # or warn
  enable_stdout: false
```

3. **Monitor resources**:
```bash
# Check memory usage
ps aux | grep clibot

# Check tmux sessions
tmux list-sessions
```

### Backup

Back up your configuration:
```bash
cp ~/.config/clibot/config.yaml ~/.config/clibot/config.yaml.backup
```

## Additional Resources

- [README.md](../README.md) - Main documentation
- [SECURITY.md](../SECURITY.md) - Security best practices
- [Configuration Guide](../README.md#configuration) - Config file reference
