<a name="readme-top"></a>

[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache License][license-shield]][license-url]

<div align="right">
  <b>简体中文</b> | <a href="README_en.md">English</a>
</div>

> **项目说明：** Nodus 前身为 **podux**（更早名为 frpc-hub），现由 [sexyfeifan/Nodus](https://github.com/sexyfeifan/Nodus) 独立维护。数据结构与 podux 兼容，可无损迁移，详见 [迁移指南](docs/guide/migration.md)。

<br />
<div align="center">
  <a href="https://github.com/sexyfeifan/Nodus">
    <img src="site/src/assets/logo.png" alt="Logo" width="140">
  </a>

<h3 align="center">Nodus</h3>

  <p align="center">
    frpc 的 Web 管理面板：集中管理 frp 客户端、内网穿透代理规则与连接状态。
    <br />
    <a href="docs/guide/getting-started.md"><strong>快速开始</strong></a>
    &nbsp;·&nbsp;
    <a href="docs/guide/migration.md"><strong>迁移指南</strong></a>
    &nbsp;·&nbsp;
    <a href="docs"><strong>文档</strong></a>
    &nbsp;·&nbsp;
    <a href="https://github.com/sexyfeifan/Nodus/issues">反馈问题</a>
  </p>

</div>

![界面截图](screenshot/1.png)

## 功能

基于当前代码实现，支持：

- **多节点管理** — 添加多个 frps 服务器（frpc 客户端节点），统一启动 / 停止 / 重载
- **代理规则管理** — 可视化配置 `tcp` / `udp` / `http` / `https` 代理（`stcp` / `xtcp` 在界面中已预留）
- **连接与日志** — 查看节点运行状态，实时输出 frpc 日志
- **网络监控** — 延迟探测与服务器地理位置信息
- **数据大盘** — 节点拓扑与连接统计
- **开机自启** — 节点配置 `autoConnection` 后随服务自动启动
- **配置导入** — 导入现有 `frpc.toml`，批量生成服务器与代理
- **权限控制** — 管理员 / 普通用户角色，写操作需管理员
- **版本检查** — 界面内查看当前版本与最新 Release

## 技术栈

| 组件 | 版本 |
| --- | --- |
| Go | 1.25 |
| PocketBase | 0.35.0 |
| frp | 0.68.0 |
| React / TypeScript / Vite | 19 / 5.9 / 7 |

## 快速部署

### 方式一：一键脚本

```bash
bash scripts/setup.sh
```

支持的环境变量：`PORT`（默认 8090）、`IMAGE`、`CONTAINER_NAME`、`VOLUME_NAME`、`MODE=docker|binary`。完整说明见 [搭建教程](docs/guide/getting-started.md)。

### 方式二：Docker

```bash
docker run -d \
  --name Nodus \
  --restart unless-stopped \
  -p 8090:8090 \
  -v Nodus-data:/app/pb_data \
  sexyfeifan/nodus:latest
```

或使用 [deploy/docker-compose.yml](deploy/docker-compose.yml)：

```bash
docker compose -f deploy/docker-compose.yml up -d
```

固定版本：

```bash
docker pull sexyfeifan/nodus:0.0.3
```

### 方式三：二进制

到 [Releases](https://github.com/sexyfeifan/Nodus/releases/latest) 下载对应平台压缩包（文件名含版本号，例如 `Nodus-v0.0.3-linux-amd64.tar.gz`）：

```bash
tar -xzf Nodus-v0.0.3-linux-amd64.tar.gz
./Nodus serve --http 0.0.0.0:8090
```

安装后访问 `http://<服务器>:8090`，按界面提示创建管理员账号。

## 从旧版本迁移

数据结构与 podux / frpc-hub 兼容，可直接拷贝 `pb_data` 无损迁移；原生 frpc 配置通过导入功能接入。

```bash
# 先预览，再执行
bash scripts/migrate-to-nodus.sh --dry-run
bash scripts/migrate-to-nodus.sh
```

脚本会自动备份、复制数据、启动 Nodus 并导入 `frpc.toml`。支持服务器上已有 **podux / frpc-hub / frpc** 三种来源；找不到旧数据时按全新安装处理。

详细步骤、手动迁移方法、**可直接发给 AI Agent 的自动迁移指令**、回滚方式见 [迁移指南](docs/guide/migration.md)。

## 文档

| 文档 | 内容 |
| --- | --- |
| [快速开始](docs/guide/getting-started.md) | Docker / Compose 安装与初始配置 |
| [迁移指南](docs/guide/migration.md) | podux / frpc-hub / frpc 迁移与 Agent 指令 |
| [配置说明](docs/guide/configuration.md) | 代理与节点配置项 |
| [安装部署](docs/deploy/install) | Docker / 二进制部署与升级 |
| [API 与开发](AGENT.md) | 本地开发、目录结构、技术细节 |

## 版本里程碑

- 2026-09-25: 发布 v0.0.1 — Nodus 首个版本（基于 podux v0.1.6）
- 2026-09-25: 发布 v0.0.2 — 安全加固与全面缺陷修复
- 2026-09-25: 发布 v0.0.3 — 修复角色提权漏洞与文档资产名

## License

[Apache 2.0](LICENSE)

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=sexyfeifan/Nodus&type=Date)](https://www.star-history.com/#sexyfeifan/Nodus&Date)

<!-- MARKDOWN LINKS & IMAGES -->

[forks-shield]: https://img.shields.io/github/forks/sexyfeifan/Nodus.svg?style=for-the-badge
[forks-url]: https://github.com/sexyfeifan/Nodus/network/members
[stars-shield]: https://img.shields.io/github/stars/sexyfeifan/Nodus.svg?style=for-the-badge
[stars-url]: https://github.com/sexyfeifan/Nodus/stargazers
[issues-shield]: https://img.shields.io/github/issues/sexyfeifan/Nodus.svg?style=for-the-badge
[issues-url]: https://github.com/sexyfeifan/Nodus/issues
[license-shield]: https://img.shields.io/github/license/sexyfeifan/Nodus.svg?style=for-the-badge
[license-url]: https://github.com/sexyfeifan/Nodus/blob/main/LICENSE
