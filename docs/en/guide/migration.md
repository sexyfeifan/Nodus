# Migration Guide

Migrate from **podux** / **frpc-hub** / **native frpc** to Nodus without losing any tunnel nodes.

Nodus shares the same PocketBase data layout as podux / frpc-hub (`fh_servers`, `fh_proxies`, `fh_users`, `fh_settings`, `fh_metrics_*` inside `pb_data`), so the database migration is a straight copy. Native frpc config files are written in through the import API.

## One-shot script

```bash
# preview first
bash scripts/migrate-to-nodus.sh --dry-run
bash scripts/migrate-to-nodus.sh
```

The script will:

1. **Detect the old `pb_data`** — Docker volumes, container mounts, `podux` / `frpc-hub` dirs under `/opt` and `$HOME`
2. **Back up** to `./nodus-migration-backup/` (timestamped)
3. **Copy data** into Nodus's Docker volume (`Nodus-data`) or a chosen directory (`--nodus-dir`)
4. **Start Nodus** — PocketBase migrations run automatically (adds `role`, tightens rules)
5. **Import `frpc.toml`** if found; set `NODUS_TOKEN` to use the API, otherwise paste in the UI

| Flag | Meaning |
| --- | --- |
| `--dry-run` | print actions only |
| `--pb-data /path/to/pb_data` | old data directory |
| `--frpc-conf /path/frpc.toml` | frpc config (repeatable) |
| `--nodus-dir /path/to/pb_data` | target data dir for binary installs |
| `--nodus-url http://ip:8090` | Nodus base URL |

## Manual migration

### A. podux / frpc-hub → Nodus

```bash
docker stop podux frpc-hub 2>/dev/null || true

# back up the old volume
docker volume ls | grep -iE 'podux|frpc-hub'
docker run --rm -v <old-volume>:/from -v "$(pwd)":/to alpine cp -a /from /to/pb_data-backup

# copy into Nodus volume and start
docker volume create Nodus-data
docker run --rm -v <old-volume>:/from -v Nodus-data:/to alpine cp -a /from/. /to/
docker run -d --name Nodus --restart unless-stopped \
  -p 8090:8090 -v Nodus-data:/app/pb_data sexyfeifan/nodus:latest
```

Log in with the **existing account**. The first start applies migrations automatically.

### B. Native frpc → Nodus

1. Install and start Nodus, finish admin setup
2. Open the **Import** page
3. Paste the full contents of `frpc.toml`
4. Preview, then execute the import
5. Verify nodes under **Servers** / **Proxies**
6. Only then stop the old `frpc` process (`systemctl stop frpc`) to avoid port conflicts

API import:

```bash
export NODUS_TOKEN="<admin token>"
curl -X POST "http://127.0.0.1:8090/api/import/execute" \
  -H "Authorization: ${NODUS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"tomlContent\": $(jq -Rs . < /etc/frp/frpc.toml)}"
```

Only **TOML** is supported. Convert legacy `frpc.ini` first, or add proxies manually in the UI.

## Agent instructions (copy-paste)

Send this to an AI agent running on your server for an automated lossless migration:

```text
Migrate this server's frpc / podux / frpc-hub installation to Nodus losslessly.

1. Inspect before touching anything:
   - Find the old PocketBase data dir pb_data (contains data.db):
     * docker volume ls / docker inspect for names matching podux, frpc-hub, frpchub, nodus
     * podux / frpc-hub dirs under /opt, /var/lib, $HOME
   - Find native frpc configs: /etc/frp/frpc.toml, /etc/frpc.toml, $HOME/frpc.toml,
     and the config path in frpc.service's ExecStart
   - Record how the old service runs (docker / systemd / bare process), ports, volume names

2. Rules (must follow):
   - Never delete, overwrite or modify existing data or config files
   - Back up pb_data to ./nodus-migration-backup/<timestamp>/ before copying
   - Nodus shares podux's data layout: copying pb_data preserves servers, proxies,
     users, settings and metrics 100%. Do not hand-edit SQLite
   - Import frpc.toml via POST /api/import/execute with body {"tomlContent": "..."}
   - frpc.ini cannot be auto-imported; report those files to the user

3. If there is no pb_data and no frpc config: do a fresh install only and prompt
   the user to create the admin account. Do not fake a migration report.

4. Deploy Nodus (prefer Docker):
   - Image sexyfeifan/nodus:latest, volume Nodus-data mounted at /app/pb_data, port 8090
   - When migrating: copy the backup/old volume into Nodus-data first, then start
   - After start: check container logs, curl http://127.0.0.1:8090/api/system/initialized

5. Switch-over:
   - With migrated data: stop the old service, keep its volume/config as rollback point
   - With frpc.toml import only: start the node in Nodus and verify the tunnel,
     then stop the old frpc process to avoid double-binding ports
   - Report: found paths, backup path, imported files, server/proxy counts,
     Nodus URL, and how to roll back (copy backup back and restart the old service)

6. Done when:
   - old pb_data is backed up and untouched
   - every server and proxy is visible in Nodus
   - the user can restore the previous state using the rollback steps
```

## Post-migration checklist

- [ ] `http://<server>:8090` accepts the original account (data migration case)
- [ ] Server count matches the pre-migration count
- [ ] Proxy rule count matches the pre-migration count
- [ ] Start at least one node and confirm the tunnel works
- [ ] Old frpc process stopped or intentionally kept (no double binding)
- [ ] `nodus-migration-backup/` is still present

## Rollback

```bash
docker stop Nodus
docker run --rm -v "$(pwd)/nodus-migration-backup/pb_data-<ts>":/from -v Nodus-data:/to \
  alpine sh -c 'rm -rf /to/* && cp -a /from/. /to/'
# then start the old podux / frpc-hub again
```

## FAQ

**Will my account password change?**  
No. `fh_users` is copied as-is, including password hashes.

**Can multiple frpc.toml files be merged into Nodus?**  
Yes. Each `serverAddr` becomes one Nodus server node; its `proxies` become proxy rules. Existing entries are flagged as duplicates and not overwritten.

**Is metrics history lost?**  
No. `fh_metrics_*` tables travel with `pb_data`.

**Import returns 403?**  
Mutating endpoints require the admin role (since v0.0.2). Use an admin token in `Authorization`.
