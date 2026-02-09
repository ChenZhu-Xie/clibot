# 部署指南

本指南介绍如何使用 systemd 或 supervisor 在生产环境中部署 clibot。

## 平台支持

**支持的平台**：
- ✅ **Linux** - 完全支持，推荐用于生产环境
- ✅ **macOS** - 完全支持
- ⚠️ **Windows** - 仅通过 WSL2 支持（不推荐用于生产环境）

**Windows 用户**：
- 使用 WSL2 进行开发/测试
- 生产环境部署到 Linux 服务器（VPS、云服务器等）
- 详见 [Windows 安装指南](../README_zh.md#windows-安装指南-wsl2)

## 目录

- [前置要求](#前置要求)
- [使用 systemd 部署](#使用-systemd-部署)
- [使用 Supervisor 部署](#使用-supervisor-部署)
- [验证部署](#验证部署)
- [故障排查](#故障排查)
- [卸载](#卸载)

## 前置要求

1. **安装 clibot**：
```bash
go install github.com/keepmind9/clibot@latest
```

**注意**：`go install` 会将二进制文件安装到 `$GOPATH/bin/clibot`（通常是 `~/go/bin/clibot`）。
确保 `~/go/bin` 在您的 PATH 中：
```bash
# 添加到 PATH（添加到 ~/.bashrc 或 ~/.zshrc 以持久化）
export PATH=$PATH:~/go/bin
```

2. **安装 tmux**（必需）：
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux

# Fedora/CentOS/RHEL
sudo dnf install tmux
```

**注意**：必须使用 `--config` 指定配置文件路径。常见选项：
- `~/.config/clibot/config.yaml` - 用户级配置（推荐）
- `/etc/clibot/config.yaml` - 系统级配置（需要 root）
- `./config.yaml` - 项目目录

如果使用自定义位置，需要更新 `--config` 路径：
- systemd 服务文件中的 `ExecStart=` 行
- supervisor 配置文件中的 `command=` 行

3. **配置 clibot**：
```bash
# 创建配置目录（用户级）
mkdir -p ~/.config/clibot

# 复制配置模板
cp configs/config.yaml ~/.config/clibot/config.yaml

# 编辑配置文件
nano ~/.config/clibot/config.yaml
```

**重要提示**：在配置文件中填写您的 Bot 凭据和白名单用户。

## 使用 systemd 部署

systemd 是现代 Linux 发行版的初始化系统（Ubuntu 16.04+、CentOS 7+ 等）。

### 安装步骤

1. **创建 systemd 用户目录并复制服务文件**：
```bash
mkdir -p ~/.config/systemd/user
cp deploy/clibot.service ~/.config/systemd/user/
```

**自定义路径**（可选）：
如果使用不同的配置位置或二进制路径，编辑服务文件：
```bash
nano ~/.config/systemd/user/clibot.service
```

需要自定义的关键行：
- `ExecStart=/usr/local/bin/clibot serve --config /etc/clibot/config.yaml`
  - 更改二进制路径（如果安装在其他位置）
  - 更改 `--config` 路径到您的配置文件位置
- `User=clibot` 和 `Group=clibot`
  - 如果使用不同用户则更改
- `WorkingDirectory=/opt/clibot`
  - 更改到您首选的工作目录

2. **重新加载 systemd**：
```bash
systemctl --user daemon-reload
```

3. **启用 clibot** 登录时自启：
```bash
systemctl --user enable clibot
```

4. **启动 clibot**：
```bash
systemctl --user start clibot
```

**可选：启用 lingering**（在启动时自动启动服务，而不仅仅是登录时）：
```bash
loginctl enable-linger $USER
```

### 管理命令

```bash
# 查看状态
systemctl --user status clibot

# 停止 clibot
systemctl --user stop clibot

# 重启 clibot
systemctl --user restart clibot

# 查看日志
journalctl --user -u clibot -f

# 查看自上次启动以来的日志
journalctl --user -u clibot -b

# 查看最近 100 行
journalctl --user -u clibot -n 100
```

### 日志轮转

systemd 通过 journald 自动处理日志轮转。配置持久化日志：

```bash
# 创建 journal 目录
sudo mkdir -p /var/log/journal

# 重启 journald
sudo systemctl restart systemd-journald
```

## 使用 Supervisor 部署

**注意**：supervisor 是系统级服务，需要 root 权限来安装和配置。对于用户级服务管理，建议使用 [systemd 用户服务](#使用-systemd-部署)。

Supervisor 是一个类 Unix 操作系统的进程控制系统。

### 安装步骤

1. **安装 supervisor**：
```bash
# Ubuntu/Debian
sudo apt-get install supervisor

# Fedora/CentOS/RHEL
sudo dnf install supervisor

# macOS
brew install supervisor
```

2. **复制配置文件**：
```bash
sudo cp deploy/clibot.conf /etc/supervisor/conf.d/clibot.conf
```

**自定义路径**（可选）：
如果使用不同的配置位置或二进制路径，编辑配置文件：
```bash
sudo nano /etc/supervisor/conf.d/clibot.conf
```

需要自定义的关键行：
- `command=~/go/bin/clibot serve --config ~/.config/clibot/config.yaml`
  - `~/go/bin/clibot`：二进制路径（如果 `$GOPATH/bin` 不同则修改）
  - 如果 `~/go/bin` 在 PATH 中，可以用 `clibot` 代替
  - `--config ~/.config/clibot/config.yaml`：配置文件路径（必需）
- `user=clibot`
  - 默认此行被注释（以当前用户运行）
  - 取消注释以专用用户运行
- `stdout_logfile=/var/log/clibot/clibot.log`
  - 更改到您首选的日志位置

3. **重新读取并更新 supervisor**：
```bash
sudo supervisorctl reread
sudo supervisorctl update
```

4. **启动 clibot**：
```bash
sudo supervisorctl start clibot
```

### 管理命令

```bash
# 查看状态
sudo supervisorctl status clibot

# 停止 clibot
sudo supervisorctl stop clibot

# 重启 clibot
sudo supervisorctl restart clibot

# 查看日志
sudo supervisorctl tail -f clibot

# 查看 stderr 日志
sudo supervisorctl tail -f clibot stderr

# 查看 stdout 日志
sudo supervisorctl tail -f clibot stdout
```

### 日志轮转

Supervisor 根据 `clibot.conf` 中的设置自动处理日志轮转：
- `stdout_logfile_maxbytes=50MB` - 达到 50MB 时轮转
- `stdout_logfile_backups=10` - 保留 10 个备份文件

## 验证部署

### 验证 clibot 是否运行

```bash
# 检查 tmux 会话是否存在
tmux list-sessions

# 检查 clibot 是否监听 8080 端口
sudo netstat -tlnp | grep 8080

# 或使用 ss
sudo ss -tlnp | grep 8080
```

### 从 IM 测试

向您的 Bot 发送消息：
```
status
```

您应该会收到状态响应。

## 故障排查

### clibot 无法启动

1. **检查服务状态**：
```bash
# systemd
sudo systemctl status clibot

# supervisor
sudo supervisorctl status clibot
```

2. **检查日志**：
```bash
# systemd
sudo journalctl -u clibot -n 100

# supervisor
sudo tail -100 /var/log/clibot/clibot.log
```

3. **常见问题**：

   **问题**：权限被拒绝
   **解决方案**：
   ```bash
   sudo chown -R clibot:clibot /etc/clibot
   sudo chown -R clibot:clibot /var/log/clibot
   ```

   **问题**：找不到配置文件
   **解决方案**：确保 `/etc/clibot/config.yaml` 存在且可读

   **问题**：端口已被占用
   **解决方案**：
   ```bash
   # 查找占用 8080 端口的进程
   sudo lsof -i :8080
   # 在 config.yaml 中更改端口
   ```

   **问题**：找不到 tmux
   **解决方案**：
   ```bash
   # 安装 tmux
   sudo apt-get install tmux  # Ubuntu/Debian
   ```

### 手动测试

手动运行 clibot 以调试问题：

```bash
# 以当前用户身份运行（默认）
/usr/local/bin/clibot serve

# 或指定配置文件
/usr/local/bin/clibot serve --config ~/.config/clibot/config.yaml

# 或使用 debug 日志级别运行
/usr/local/bin/clibot serve --config ~/.config/clibot/config.yaml --log-level debug
```

## 卸载

### systemd

```bash
# 停止并禁用
systemctl --user stop clibot
systemctl --user disable clibot

# 删除服务文件
rm ~/.config/systemd/user/clibot.service
systemctl --user daemon-reload

# 删除用户配置（可选）
rm -rf ~/.config/clibot
```

### Supervisor

```bash
# 停止 clibot
sudo supervisorctl stop clibot

# 删除配置
sudo rm /etc/supervisor/conf.d/clibot.conf

# 重新读取并更新
sudo supervisorctl reread
sudo supervisorctl update

# 删除用户配置（可选）
rm -rf ~/.config/clibot
```

## 生产环境建议

### 安全性

1. **启用白名单** 在 config.yaml 中：
```yaml
security:
  whitelist_enabled: true
  allowed_users:
    discord:
      - "your-user-id"
```

2. **使用环境变量存储密钥**：
```yaml
bots:
  discord:
    token: "${DISCORD_TOKEN}"
```

设置方式：
```bash
# 对于 systemd 用户服务
systemctl --user edit clibot
# 添加: [Service] Environment=DISCORD_TOKEN=your_token

# 对于 supervisor
sudo nano /etc/supervisor/conf.d/clibot.conf
# 添加: environment=DISCORD_TOKEN="your_token"
```

### 性能

1. **限制动态会话数**：
```yaml
session:
  max_dynamic_sessions: 50
```

2. **调整日志级别** 用于生产环境：
```yaml
logging:
  level: info  # 或 warn
  enable_stdout: false
```

3. **监控资源**：
```bash
# 检查内存使用
ps aux | grep clibot

# 检查 tmux 会话
tmux list-sessions
```

### 备份

备份您的配置：
```bash
cp ~/.config/clibot/config.yaml ~/.config/clibot/config.yaml.backup
```

## 其他资源

- [README.md](../README_zh.md) - 主文档
- [SECURITY.md](../SECURITY.md) - 安全最佳实践
- [配置指南](../README_zh.md#配置) - 配置文件参考
