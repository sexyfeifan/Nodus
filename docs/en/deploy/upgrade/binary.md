# Binary Upgrade

::: warning Work in Progress
This page is still being improved. Some details may not be fully accurate — please refer to the actual behavior when in doubt.
:::

Before upgrading the Nodus binary, back up your data first, then replace the executable.

## Upgrade Steps

### 1. Back Up Data

Data is stored in `pb_data/` under the working directory. Back it up before upgrading:

```bash
cp -r /var/lib/Nodus/pb_data /var/lib/Nodus/pb_data.bak
```

### 2. Stop the Service

```bash
sudo systemctl stop Nodus
```

If not using systemd, kill the process manually:

```bash
pkill Nodus
```

### 3. Download the New Version

Visit [GitHub Releases](https://github.com/sexyfeifan/Nodus/releases/latest) for the latest version, or download via command line (Linux amd64 example):

```bash
curl -LO https://github.com/sexyfeifan/Nodus/releases/latest/download/Nodus-linux-amd64.tar.gz
tar -xzf Nodus-linux-amd64.tar.gz
```

### 4. Replace the Binary

```bash
sudo mv Nodus /usr/local/bin/Nodus
sudo chmod +x /usr/local/bin/Nodus
```

### 5. Restart the Service

```bash
sudo systemctl start Nodus

# Confirm service is running
sudo systemctl status Nodus
```

## Rollback

If issues occur after upgrading, restore the backup and reinstall the old version:

```bash
# Stop service
sudo systemctl stop Nodus

# Restore data backup
rm -rf /var/lib/Nodus/pb_data
cp -r /var/lib/Nodus/pb_data.bak /var/lib/Nodus/pb_data

# Download the old version from GitHub Releases, replace the binary, then start
sudo systemctl start Nodus
```
