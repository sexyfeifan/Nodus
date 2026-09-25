#!/usr/bin/env bash
# Nodus 迁移脚本：从 podux / frpc-hub / 原生 frpc 无损迁移到 Nodus
#
# 用法:
#   bash scripts/migrate-to-nodus.sh                 # 自动探测并迁移
#   bash scripts/migrate-to-nodus.sh --dry-run       # 只打印将执行的操作
#   bash scripts/migrate-to-nodus.sh --pb-data /path/to/pb_data
#   bash scripts/migrate-to-nodus.sh --frpc-conf /etc/frp/frpc.toml
#
# 说明:
#   1) podux / frpc-hub 与 Nodus 共用同一套 PocketBase 数据结构 (pb_data)，
#      直接拷贝 pb_data 即可 100% 保留服务器、代理、用户、指标等数据。
#   2) 原生 frpc 配置文件 (frpc.toml / frpc.ini) 通过 Nodus 导入接口写入，
#      不会丢失已有穿透节点。本脚本只导入，不删除原配置。
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; BLUE='\033[0;34m'; NC='\033[0m'
info()  { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }
step()  { printf "${BLUE}[STEP]${NC} %s\n" "$*"; }

DRY_RUN=0
SRC_PB_DATA=""
FRPC_CONF=()
NODUS_URL="${NODUS_URL:-http://127.0.0.1:8090}"
NODUS_DIR="${NODUS_DIR:-}"       # 目标 pb_data（Docker 部署时留空，用 volume）
IMAGE="${IMAGE:-sexyfeifan/nodus:latest}"
BACKUP_ROOT="${BACKUP_ROOT:-./nodus-migration-backup}"

usage() {
  sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)      DRY_RUN=1; shift ;;
    --pb-data)      SRC_PB_DATA="$2"; shift 2 ;;
    --frpc-conf)    FRPC_CONF+=("$2"); shift 2 ;;
    --nodus-url)    NODUS_URL="$2"; shift 2 ;;
    --nodus-dir)    NODUS_DIR="$2"; shift 2 ;;
    --backup-root)  BACKUP_ROOT="$2"; shift 2 ;;
    -h|--help)      usage ;;
    *) error "未知参数: $1"; usage ;;
  esac
done

run() {
  if [ "$DRY_RUN" = "1" ]; then
    echo "  [dry-run] $*"
  else
    eval "$@"
  fi
}

# ── 1. 探测旧版 pb_data（podux / frpc-hub）──────────────────
detect_pb_data() {
  if [ -n "$SRC_PB_DATA" ]; then
    echo "$SRC_PB_DATA"; return 0
  fi

  local candidates=()

  # Docker volume 中的常见位置
  if command -v docker >/dev/null 2>&1; then
    for vol in $(docker volume ls -q 2>/dev/null | grep -iE 'podux|frpc-hub|frpchub|nodus' || true); do
      local mp
      mp=$(docker volume inspect "$vol" --format '{{.Mountpoint}}' 2>/dev/null || true)
      if [ -n "$mp" ] && [ -d "$mp" ] && [ -f "$mp/data.db" -o -f "$mp/data.db-wal" ]; then
        candidates+=("$mp")
      fi
    done

    # 正在运行/已停止容器的挂载点
    for c in $(docker ps -a --format '{{.Names}}' | grep -iE 'podux|frpc-hub|frpchub|nodus' || true); do
      local mp
      mp=$(docker inspect "$c" --format '{{range .Mounts}}{{if eq .Destination "/app/pb_data"}}{{.Source}}{{end}}{{end}}' 2>/dev/null || true)
      if [ -n "$mp" ] && [ -d "$mp" ]; then
        candidates+=("$mp")
      fi
    done
  fi

  # 本机常见目录
  for d in \
    ./pb_data ./podux/pb_data ./frpc-hub/pb_data \
    /opt/podux/pb_data /opt/frpc-hub/pb_data /opt/nodus/pb_data \
    /var/lib/podux /var/lib/frpc-hub \
    "$HOME/podux/pb_data" "$HOME/frpc-hub/pb_data"; do
    if [ -d "$d" ]; then
      candidates+=("$d")
    fi
  done

  if [ "${#candidates[@]}" -eq 0 ]; then
    echo ""
    return 0
  fi

  # 返回第一个包含 data.db 的
  for c in "${candidates[@]}"; do
    if [ -f "$c/data.db" ] || [ -f "$c/data.db-wal" ]; then
      echo "$c"
      return 0
    fi
  done
  echo "${candidates[0]}"
}

# ── 2. 探测原生 frpc 配置 ───────────────────────────────────
detect_frpc_conf() {
  if [ "${#FRPC_CONF[@]}" -gt 0 ]; then
    printf '%s\n' "${FRPC_CONF[@]}"
    return 0
  fi
  local found=()
  for f in \
    /etc/frp/frpc.toml /etc/frp/frpc.ini \
    /etc/frpc.toml /etc/frpc.ini \
    /opt/frp/frpc.toml /opt/frp/frpc.ini \
    "$HOME/frpc.toml" "$HOME/frpc.ini" \
    "$HOME/.config/frp/frpc.toml" \
    ./frpc.toml ./frpc.ini; do
    if [ -f "$f" ]; then
      found+=("$f")
    fi
  done
  # systemd 服务中的 ExecStart 路径
  if command -v systemctl >/dev/null 2>&1; then
    local unit
    unit=$(systemctl cat frpc 2>/dev/null | grep -oE '(/[^ ]*frpc\.(toml|ini))' | head -1 || true)
    if [ -n "$unit" ] && [ -f "$unit" ]; then
      found+=("$unit")
    fi
  fi
  if [ "${#found[@]}" -gt 0 ]; then
    printf '%s\n' "${found[@]}"
  fi
}

# ── 3. 迁移 pb_data ─────────────────────────────────────────
migrate_pb_data() {
  local src="$1"

  if [ -z "$src" ] || [ ! -d "$src" ]; then
    warn "未找到旧版 pb_data（podux / frpc-hub），跳过数据迁移，按全新安装处理。"
    return 1
  fi

  step "发现旧版数据目录: $src"

  if [ ! -f "$src/data.db" ] && [ ! -f "$src/data.db-wal" ]; then
    warn "$src 中没有 data.db，不是有效的 PocketBase 数据目录。"
    return 1
  fi

  local ts backup
  ts=$(date +%Y%m%d-%H%M%S)
  backup="${BACKUP_ROOT}/pb_data-${ts}"
  step "备份到 ${backup}"
  run "mkdir -p '${BACKUP_ROOT}'"
  run "cp -a '${src}' '${backup}'"

  if [ -n "$NODUS_DIR" ]; then
    # 目标为本地目录（binary 部署）
    step "复制到 Nodus 数据目录 ${NODUS_DIR}"
    run "mkdir -p '${NODUS_DIR}'"
    if [ "$DRY_RUN" = "1" ]; then
      echo "  [dry-run] cp -a '${src}/.' '${NODUS_DIR}/'"
    else
      cp -a "${src}/." "${NODUS_DIR}/"
    fi
  else
    # 目标为 Docker volume
    if ! command -v docker >/dev/null 2>&1; then
      error "未安装 docker 且未指定 --nodus-dir，无法定位目标数据目录。"
      exit 1
    fi
    step "复制到 Docker volume Nodus-data"
    run "docker volume create Nodus-data >/dev/null"
    if [ "$DRY_RUN" = "1" ]; then
      echo "  [dry-run] docker run --rm -v '${src}':/from -v Nodus-data:/to alpine cp -a /from/. /to/"
    else
      docker run --rm -v "${src}":/from -v Nodus-data:/to alpine cp -a /from/. /to/
    fi
  fi

  info "pb_data 迁移完成（服务器、代理、用户、指标数据均已保留）。"
  return 0
}

# ── 4. 导入 frpc 配置 ──────────────────────────────────────
import_frpc_conf() {
  local confs
  confs=$(detect_frpc_conf || true)
  if [ -z "$confs" ]; then
    warn "未发现 frpc 配置文件，跳过配置导入。"
    return 0
  fi

  local admin_token="${NODUS_TOKEN:-}"
  if [ -z "$admin_token" ]; then
    warn "未设置 NODUS_TOKEN（管理员 API token），无法通过 API 导入 frpc 配置。"
    warn "请在 Nodus 界面：设置 → 导入 → 粘贴 frpc TOML 内容，或设置 NODUS_TOKEN 后重试。"
    while IFS= read -r f; do
      [ -n "$f" ] && echo "  待导入配置: $f"
    done <<< "$confs"
    return 0
  fi

  while IFS= read -r f; do
    [ -z "$f" ] && continue
    # INI 格式暂不支持直接导入，提示转换
    if [[ "$f" == *.ini ]]; then
      warn "INI 格式暂不支持自动导入: $f"
      warn "请手动在 Nodus 导入页粘贴内容，或先转为 TOML。"
      continue
    fi

    step "导入 frpc 配置: $f"
    if [ "$DRY_RUN" = "1" ]; then
      echo "  [dry-run] POST ${NODUS_URL}/api/import/execute  <- $f"
      continue
    fi

    local payload
    payload=$(python3 - "$f" <<'PY'
import json, sys
content = open(sys.argv[1], encoding="utf-8").read()
print(json.dumps({"tomlContent": content}))
PY
)
    local resp
    resp=$(curl -fsS -X POST "${NODUS_URL}/api/import/execute" \
      -H "Authorization: ${admin_token}" \
      -H "Content-Type: application/json" \
      -d "$payload" || true)

    if echo "$resp" | grep -q '"error"'; then
      error "导入失败: $resp"
    else
      info "导入结果: $resp"
    fi
  done <<< "$confs"
}

# ── 5. 启动 Nodus ───────────────────────────────────────────
start_nodus() {
  if ! command -v docker >/dev/null 2>&1; then
    warn "未检测到 docker。若以二进制方式部署，请自行启动: Nodus serve --http 0.0.0.0:8090"
    return 0
  fi

  if docker ps --format '{{.Names}}' | grep -qx 'Nodus'; then
    info "Nodus 容器已在运行。"
    return 0
  fi

  step "启动 Nodus 容器"
  run "docker rm -f Nodus >/dev/null 2>&1 || true"
  run "docker pull ${IMAGE}"
  run "docker run -d --name Nodus --restart unless-stopped -p 8090:8090 -v Nodus-data:/app/pb_data -e TZ=Asia/Shanghai ${IMAGE}"

  if [ "$DRY_RUN" = "1" ]; then
    return 0
  fi

  sleep 3
  if docker ps --format '{{.Names}}' | grep -qx 'Nodus'; then
    info "Nodus 已启动: ${NODUS_URL}"
  else
    error "Nodus 启动失败，请查看: docker logs Nodus"
    exit 1
  fi
}

# ── main ────────────────────────────────────────────────────
echo
info "Nodus 迁移工具（podux / frpc-hub / frpc → Nodus）"
[ "$DRY_RUN" = "1" ] && warn "dry-run 模式：只显示将执行的操作，不做任何修改"
echo

src_pb=$(detect_pb_data || true)
had_data=0
if migrate_pb_data "$src_pb"; then
  had_data=1
fi

start_nodus

if [ "$had_data" = "1" ]; then
  info "迁移完成：旧数据已导入，启动后迁移脚本会自动升级集合规则。"
  echo "  请打开 ${NODUS_URL} 确认服务器与代理列表完整。"
else
  info "未迁移旧数据（按全新安装处理）。"
  echo "  请打开 ${NODUS_URL} 完成管理员初始化。"
fi

import_frpc_conf

echo
info "迁移流程结束。备份目录: ${BACKUP_ROOT}"
