# Nodus 迁移指南

从 **podux** / **frpc-hub** / **原生 frpc** 迁移到 Nodus，保留全部内网穿透节点配置。

Nodus 与 podux / frpc-hub 共用同一套 PocketBase 数据结构（`pb_data` 中的 `fh_servers`、`fh_proxies`、`fh_users`、`fh_settings`、`fh_metrics_*`），因此数据库迁移是**直接拷贝**，不会丢失任何服务器、代理、用户或监控数据。原生 frpc 的配置文件通过导入接口写入。

## 一键脚本

```bash
# 在服务器上执行（推荐先 --dry-run 预览）
bash scripts/migrate-to-nodus.sh --dry-run
bash scripts/migrate-to-nodus.sh
```

脚本会依次：

1. **探测旧版 `pb_data`** — Docker volume、容器挂载点、`/opt`、`$HOME` 下的 podux / frpc-hub 数据目录
2. **完整备份**到 `./nodus-migration-backup/`（含时间戳）
3. **拷贝数据**到 Nodus 的 Docker volume（`Nodus-data`）或指定目录（`--nodus-dir`）
4. **启动 Nodus** 容器；启动时 PocketBase 自动执行迁移（补 `role` 字段、收紧权限规则），无需手工操作
5. **导入 frpc 配置**（若有 `frpc.toml`）；设置 `NODUS_TOKEN` 后可走 API，否则提示在界面粘贴

常用参数：

| 参数 | 说明 |
| --- | --- |
| `--dry-run` | 只打印操作，不修改任何文件 |
| `--pb-data /path/to/pb_data` | 指定旧版数据目录 |
| `--frpc-conf /path/frpc.toml` | 指定 frpc 配置（可多次传入） |
| `--nodus-dir /path/to/pb_data` | 二进制部署时的目标数据目录 |
| `--nodus-url http://ip:8090` | Nodus 地址（导入接口用） |

## 手动迁移

### A. podux / frpc-hub → Nodus

```bash
# 1. 停止旧服务
docker stop podux frpc-hub 2>/dev/null || true

# 2. 备份旧数据（volume 名以实际为准，可用 docker volume ls 查看）
docker volume ls | grep -iE 'podux|frpc-hub'
docker run --rm -v <旧volume>:/from -v "$(pwd)":/to alpine \
  cp -a /from /to/pb_data-backup

# 3. 复制到 Nodus 数据卷
docker volume create Nodus-data
docker run --rm -v <旧volume>:/from -v Nodus-data:/to alpine cp -a /from/. /to/

# 4. 启动 Nodus
docker run -d --name Nodus --restart unless-stopped \
  -p 8090:8090 -v Nodus-data:/app/pb_data \
  sexyfeifan/nodus:latest
```

打开 `http://<服务器>:8090`，用**原有账号登录**，服务器与代理列表应完整显示。首次启动会自动执行数据迁移（新增 `role` 字段并把已有用户设为管理员）。

### B. 原生 frpc → Nodus

两种方式，任选其一：

**方式 1：界面导入（推荐）**

1. 安装并启动 Nodus，完成管理员初始化
2. 打开 **导入** 页面
3. 打开服务器上的 `frpc.toml`，把**全部内容**粘贴进去
4. 预览确认服务器地址与代理列表 → 执行导入
5. 检查 **服务器** / **代理** 页面，确认节点齐全
6. 确认无误后，再停掉原 `frpc` 进程（`systemctl stop frpc` 或 kill 进程），避免端口冲突

**方式 2：API 导入**

```bash
# 先在 Nodus 获取管理员认证 token（登录后从浏览器请求头 Authorization 获取）
export NODUS_TOKEN="<你的token>"

curl -X POST "http://127.0.0.1:8090/api/import/execute" \
  -H "Authorization: ${NODUS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"tomlContent\": $(jq -Rs . < /etc/frp/frpc.toml)}"
```

> 仅支持 **TOML** 格式（frpc 较新版本默认）。旧的 `frpc.ini` 请先手动转成 TOML，或在界面逐条添加代理。

## Agent 自动迁移指令

以下内容可直接发给在你服务器上运行的 AI Agent（如 Claude Code），让它自动完成无损迁移：

```text
请把这台服务器上的 frpc / podux / frpc-hub 无损迁移到 Nodus。要求：

1. 先探测现状，再动手：
   - 查找旧版 PocketBase 数据目录 pb_data（内含 data.db）：
     * docker volume ls / docker inspect 所有名字含 podux、frpc-hub、frpchub、nodus 的 volume 与容器
     * /opt、/var/lib、$HOME 下的 podux、frpc-hub 目录
   - 查找原生 frpc 配置：/etc/frp/frpc.toml、/etc/frpc.toml、$HOME/frpc.toml，
     以及 systemd 单元 frpc.service 的 ExecStart 参数里的配置路径
   - 记录：旧服务的运行方式（docker / systemd / 裸进程）、端口、volume 名

2. 迁移规则（必须遵守）：
   - 不要删除、不要覆盖、不要修改任何旧数据和旧配置文件
   - 复制 pb_data 前先完整备份到 ./nodus-migration-backup/<时间戳>/
   - Nodus 与 podux 共用同一套数据结构，pb_data 直接拷贝即可 100% 保留
     服务器、代理、用户、设置、监控指标；不要手工改 SQLite
   - 原生 frpc 的 frpc.toml 通过 Nodus 导入接口导入（POST /api/import/execute，
     body 为 {"tomlContent": "..."}），不要改写原文件
   - frpc.ini 不支持自动导入，报告给用户并列出需要手工转换的文件

3. 若找不到任何 pb_data 且没有 frpc 配置：按全新安装处理，
   只部署 Nodus 并提示用户完成管理员初始化，不要伪造迁移结果。

4. 部署 Nodus（优先 Docker）：
   - 镜像 sexyfeifan/nodus:latest，数据卷 Nodus-data 挂载到 /app/pb_data，端口 8090
   - 若迁移了 pb_data：先把备份目录/旧 volume 内容复制进 Nodus-data，再启动容器
   - 等待启动后检查容器日志无报错，curl http://127.0.0.1:8090/api/system/initialized
     应返回 JSON（已迁移数据时 initialized 为 true）

5. 切换与收尾：
   - 迁移数据场景：直接停用旧服务（systemctl stop / docker stop），
     保留其配置与数据卷作为回滚点，不要删除
   - 仅 frpc.toml 导入场景：先在 Nodus 启动对应节点并确认穿透正常，
     再停旧 frpc 进程，避免端口/连接冲突
   - 输出迁移报告：发现的旧数据路径、备份路径、导入的配置文件、
     服务器与代理数量、Nodus 访问地址、回滚方法（把备份拷回原路径并重启旧服务）

6. 验收标准：
   - 旧 pb_data 已备份且原文件未改动
   - Nodus 界面/接口中能看到迁移过来的全部服务器与代理
   - 需要时用户可按回滚方法恢复到迁移前状态
```

## 迁移后检查清单

- [ ] `http://<服务器>:8090` 能用原账号登录（迁移数据场景）
- [ ] **服务器** 页面节点数量与迁移前一致
- [ ] **代理** 页面规则数量与迁移前一致
- [ ] 对至少一个节点执行**启动**，确认连接正常
- [ ] 确认旧 frpc 进程已按计划停止或保留（避免双开冲突）
- [ ] 备份目录 `nodus-migration-backup/` 仍在

## 回滚

```bash
# Docker
docker stop Nodus
docker run --rm -v "$(pwd)/nodus-migration-backup/pb_data-<时间戳>":/from -v Nodus-data:/to alpine sh -c 'rm -rf /to/* && cp -a /from/. /to/'
# 然后启动旧的 podux / frpc-hub

# 或二进制：把备份目录拷回原 pb_data 路径后重启旧服务
```

## 常见问题

**Q: 迁移后账号密码变了？**  
不会。`fh_users` 整表保留，密码哈希原样拷贝。

**Q: 支持把多个 frpc.toml 合并进 Nodus 吗？**  
支持。每个 `frpc.toml` 的 `serverAddr` 对应一个 Nodus「服务器」节点，其 `proxies` 对应多条代理。可分多次导入，已有服务器/代理会标记为重复不会覆盖。

**Q: 指标和历史数据会丢吗？**  
不会。`fh_metrics_raw` / `fh_metrics_hourly` / `fh_metrics_daily` 随 `pb_data` 一起迁移。

**Q: 导入接口报 403？**  
需要管理员权限。用管理员账号登录后拿 `Authorization` token（v0.0.2 起写操作接口需管理员角色）。
