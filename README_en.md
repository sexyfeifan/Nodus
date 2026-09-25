<a name="readme-top"></a>

[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache License][license-shield]][license-url]

<div align="right">
  <a href="README.md">简体中文</a> | <b>English</b>
</div>

> **Note:** Nodus was formerly known as **podux** (and earlier as **frpc-hub**), now maintained independently at [sexyfeifan/Nodus](https://github.com/sexyfeifan/Nodus). The data layout is compatible with podux and can be migrated without loss — see the [Migration Guide](docs/guide/migration.md).

<br />
<div align="center">
  <a href="https://github.com/sexyfeifan/Nodus">
    <img src="site/src/assets/logo.png" alt="Logo" width="140">
  </a>

<h3 align="center">Nodus</h3>

  <p align="center">
    A web management panel for frpc: manage frp clients, tunnel proxy rules and connection status in one place.
    <br />
    <a href="docs/guide/getting-started.md"><strong>Getting Started</strong></a>
    &nbsp;·&nbsp;
    <a href="docs/guide/migration.md"><strong>Migration</strong></a>
    &nbsp;·&nbsp;
    <a href="docs"><strong>Docs</strong></a>
    &nbsp;·&nbsp;
    <a href="https://github.com/sexyfeifan/Nodus/issues">Issues</a>
  </p>

</div>

![Screenshot](screenshot/1.png)

## Features

What the current codebase actually supports:

- **Multi-node management** — add multiple frps servers (frpc client nodes) and start / stop / reload them
- **Proxy rules** — visual configuration of `tcp` / `udp` / `http` / `https` tunnels (`stcp` / `xtcp` placeholders in the UI)
- **Status & logs** — node connection status and live frpc log streaming
- **Network monitoring** — latency probes and server geolocation
- **Dashboard** — node topology and connection statistics
- **Auto-start** — nodes marked `autoConnection` start with the service
- **Config import** — import an existing `frpc.toml` to bulk-create servers and proxies
- **Access control** — admin / user roles; mutating operations require admin
- **Version check** — current version and latest release shown in the UI

## Tech Stack

| Component | Version |
| --- | --- |
| Go | 1.25 |
| PocketBase | 0.35.0 |
| frp | 0.68.0 |
| React / TypeScript / Vite | 19 / 5.9 / 7 |

## Quick Start

### Option 1: Setup script

```bash
bash scripts/setup.sh
```

Environment variables: `PORT` (default 8090), `IMAGE`, `CONTAINER_NAME`, `VOLUME_NAME`, `MODE=docker|binary`. See the [Getting Started guide](docs/guide/getting-started.md).

### Option 2: Docker

```bash
docker run -d \
  --name Nodus \
  --restart unless-stopped \
  -p 8090:8090 \
  -v Nodus-data:/app/pb_data \
  sexyfeifan/nodus:latest
```

Or with [deploy/docker-compose.yml](deploy/docker-compose.yml):

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Pinned version:

```bash
docker pull sexyfeifan/nodus:0.0.3
```

### Option 3: Binary

Download the archive for your platform from [Releases](https://github.com/sexyfeifan/Nodus/releases/latest) (the filename includes the version, e.g. `Nodus-v0.0.3-linux-amd64.tar.gz`):

```bash
tar -xzf Nodus-v0.0.3-linux-amd64.tar.gz
./Nodus serve --http 0.0.0.0:8090
```

Then open `http://<server>:8090` and create the admin account.

## Migrating from podux / frpc-hub / frpc

Nodus shares the same PocketBase data layout as podux / frpc-hub, so copying `pb_data` migrates everything without loss. Native frpc configs are imported through the import API.

```bash
# preview first, then run
bash scripts/migrate-to-nodus.sh --dry-run
bash scripts/migrate-to-nodus.sh
```

The script backs up, copies data, starts Nodus and imports `frpc.toml`. If no previous data is found it performs a fresh install.

Step-by-step instructions, manual migration, **copy-paste agent instructions for automated migration**, and rollback are in the [Migration Guide](docs/guide/migration.md).

## Documentation

| Document | Contents |
| --- | --- |
| [Getting Started](docs/guide/getting-started.md) | Docker / Compose install and initial setup |
| [Migration Guide](docs/guide/migration.md) | podux / frpc-hub / frpc migration + agent instructions |
| [Configuration](docs/guide/configuration.md) | Proxy and node options |
| [Deployment](docs/deploy/install) | Docker / binary install and upgrade |
| [Development](AGENT.md) | Local development, directory map, internals |

## Release Timeline

- 2026-09-25: v0.0.1 — first Nodus release (based on podux v0.1.6)
- 2026-09-25: v0.0.2 — security hardening and defect fixes
- 2026-09-25: v0.0.3 — role-escalation fix and docs asset names

## License

[Apache 2.0](LICENSE)

## Star History

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
