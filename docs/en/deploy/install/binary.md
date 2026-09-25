# Binary Installation

::: warning Work in Progress
This page is still being improved. Some details may not be fully accurate — please refer to the actual behavior when in doubt.
:::

Install Nodus using pre-compiled binaries — ideal for running directly on a server or local machine.

## Download

Go to the [GitHub Releases](https://github.com/sexyfeifan/Nodus/releases/latest) page and download the package for your platform:

| OS | Architecture | Filename |
| --- | --- | --- |
| Linux | x86_64 (amd64) | `Nodus-linux-amd64.tar.gz` |
| Linux | ARM64 | `Nodus-linux-arm64.tar.gz` |
| Linux | ARMv7 | `Nodus-linux-armv7.tar.gz` |
| macOS | Intel | `Nodus-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `Nodus-darwin-arm64.tar.gz` |
| Windows | x86_64 | `Nodus-windows-amd64.zip` |

Or download directly from the command line (Linux amd64 example):

```bash
curl -LO https://github.com/sexyfeifan/Nodus/releases/latest/download/Nodus-linux-amd64.tar.gz
```

## Install

### Linux / macOS

```bash
# Extract
tar -xzf Nodus-linux-amd64.tar.gz

# Move to system path (optional)
sudo mv Nodus /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/Nodus

# Verify installation
Nodus --version
```

### Windows

Extract the `.zip` file to any directory, then open PowerShell or Command Prompt inside that directory to use it.

## Start

```bash
Nodus serve --http 0.0.0.0:8090
```

After starting, visit `http://localhost:8090` for the main interface and `http://localhost:8090/_/` for the admin panel.

::: tip
On first launch, the database is initialized automatically. Data is stored in a `pb_data/` folder in the current working directory. Fix the working directory to avoid data scatter.
:::

## Auto-start on Boot (Linux)

Use systemd to manage the Nodus process.

**1. Create a dedicated user (optional but recommended)**

```bash
sudo useradd -r -s /bin/false Nodus
```

**2. Create the data directory**

```bash
sudo mkdir -p /var/lib/Nodus
sudo chown Nodus:Nodus /var/lib/Nodus
```

**3. Create the systemd service file**

```bash
sudo tee /etc/systemd/system/Nodus.service > /dev/null <<EOF
[Unit]
Description=Nodus Service
After=network.target

[Service]
Type=simple
User=Nodus
WorkingDirectory=/var/lib/Nodus
ExecStart=/usr/local/bin/Nodus serve --http 0.0.0.0:8090
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
```

**4. Enable and start the service**

```bash
sudo systemctl daemon-reload
sudo systemctl enable Nodus
sudo systemctl start Nodus

# Check status
sudo systemctl status Nodus
```

## View Logs

```bash
sudo journalctl -u Nodus -f
```

## Next Steps

- [Binary Upgrade](/en/deploy/upgrade/binary) — Upgrade to a new version
- [Getting Started](/en/guide/getting-started) — Add servers and proxies
