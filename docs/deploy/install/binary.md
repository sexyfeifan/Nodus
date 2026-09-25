# 二进制安装

::: warning 文档待完善
本页内容正在持续完善中，部分细节可能不够准确，请以实际情况为准。
:::

通过预编译的二进制文件安装 Nodus，适合直接在服务器或本地机器上运行。

## 下载

前往 [GitHub Releases](https://github.com/sexyfeifan/Nodus/releases/latest) 页面，根据你的系统选择对应的安装包：

| 系统 | 架构 | 文件名 |
| --- | --- | --- |
| Linux | x86_64 (amd64) | `Nodus-v0.0.3-linux-amd64.tar.gz` |
| Linux | ARM64 | `Nodus-v0.0.3-linux-arm64.tar.gz` |
| Linux | ARMv7 | `Nodus-v0.0.3-linux-arm.tar.gz` |
| macOS | Intel | `Nodus-v0.0.3-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `Nodus-v0.0.3-darwin-arm64.tar.gz` |
| Windows | x86_64 | `Nodus-v0.0.3-windows-amd64.zip` |

也可以通过命令行直接下载（以 Linux amd64 为例）：

```bash
curl -LO https://github.com/sexyfeifan/Nodus/releases/download/v0.0.3/Nodus-v0.0.3-linux-amd64.tar.gz
```

## 安装

### Linux / macOS

```bash
# 解压
tar -xzf Nodus-v0.0.3-linux-amd64.tar.gz

# 移动到系统路径（可选）
sudo mv Nodus /usr/local/bin/

# 赋予执行权限
sudo chmod +x /usr/local/bin/Nodus
```

### Windows

解压 `.zip` 文件到任意目录，在解压目录中打开 PowerShell 或命令提示符即可使用。

## 启动

```bash
Nodus serve --http 0.0.0.0:8090
```

启动后访问 `http://localhost:8090` 进入主界面，`http://localhost:8090/_/` 进入管理后台。

::: tip
首次启动会自动初始化数据库，数据默认存储在当前目录下的 `pb_data/` 文件夹中。建议固定工作目录，避免数据分散。
:::

## 配置开机自启（Linux）

推荐使用 systemd 管理 Nodus 进程。

**1. 创建专用用户（可选但推荐）**

```bash
sudo useradd -r -s /bin/false Nodus
```

**2. 创建数据目录**

```bash
sudo mkdir -p /var/lib/Nodus
sudo chown Nodus:Nodus /var/lib/Nodus
```

**3. 创建 systemd 服务文件**

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

**4. 启用并启动服务**

```bash
sudo systemctl daemon-reload
sudo systemctl enable Nodus
sudo systemctl start Nodus

# 查看运行状态
sudo systemctl status Nodus
```

## 查看日志

```bash
sudo journalctl -u Nodus -f
```
