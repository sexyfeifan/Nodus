#!/usr/bin/env bash
# Nodus 一键部署脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/sexyfeifan/Nodus/main/scripts/setup.sh | bash
# 或本地执行: bash scripts/setup.sh
set -euo pipefail

IMAGE="${IMAGE:-sexyfeifan/nodus:latest}"
CONTAINER_NAME="${CONTAINER_NAME:-Nodus}"
PORT="${PORT:-8090}"
VOLUME_NAME="${VOLUME_NAME:-Nodus-data}"
DATA_DIR="${DATA_DIR:-}" # 非 Docker 部署时的 pb_data 路径（留空则使用 Docker volume）

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }

usage() {
  cat <<'EOF'
Nodus 部署脚本

环境变量（均可选）:
  IMAGE            镜像名（默认 sexyfeifan/nodus:latest）
  PORT             宿主机端口（默认 8090）
  CONTAINER_NAME   容器名（默认 Nodus）
  VOLUME_NAME      Docker volume 名（默认 Nodus-data）
  MODE             docker（默认）| binary
  DATA_DIR         MODE=binary 时 pb_data 目录（默认 ./pb_data）

示例:
  bash scripts/setup.sh
  PORT=9000 bash scripts/setup.sh
  MODE=binary bash scripts/setup.sh
EOF
}

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
  usage
  exit 0
fi

MODE="${MODE:-docker}"

# ── Docker 模式 ─────────────────────────────────────────────
if [ "$MODE" = "docker" ]; then
  if ! command -v docker >/dev/null 2>&1; then
    error "未检测到 docker。请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
  fi

  if ! docker info >/dev/null 2>&1; then
    error "Docker 守护进程未运行或无权限访问。"
    exit 1
  fi

  info "拉取镜像 ${IMAGE} ..."
  docker pull "${IMAGE}"

  if docker ps -a --format '{{.Names}}' | grep -qx "${CONTAINER_NAME}"; then
    warn "容器 ${CONTAINER_NAME} 已存在。"
    read -r -p "是否重建并覆盖运行？数据卷 ${VOLUME_NAME} 不会被删除。[y/N] " ans
    case "$ans" in
      y|Y|yes|YES) docker rm -f "${CONTAINER_NAME}" >/dev/null ;;
      *) info "已取消。"; exit 0 ;;
    esac
  fi

  info "启动容器 ${CONTAINER_NAME}（端口 ${PORT}，数据卷 ${VOLUME_NAME}）..."
  docker run -d \
    --name "${CONTAINER_NAME}" \
    --restart unless-stopped \
    -p "${PORT}:8090" \
    -v "${VOLUME_NAME}:/app/pb_data" \
    -e TZ="${TZ:-Asia/Shanghai}" \
    "${IMAGE}"

  sleep 2
  if ! docker ps --format '{{.Names}}' | grep -qx "${CONTAINER_NAME}"; then
    error "容器启动失败。查看日志: docker logs ${CONTAINER_NAME}"
    exit 1
  fi

  info "部署完成。"
  echo
  echo "  管理面板:  http://localhost:${PORT}"
  echo "  PocketBase 管理: http://localhost:${PORT}/_/"
  echo
  echo "下一步:"
  echo "  1. 打开管理面板，完成管理员账号初始化"
  echo "  2. 添加 frp 服务器（frps）节点"
  echo "  3. 添加代理规则，或从 frpc TOML 配置导入"
  echo "  4. 从旧版本迁移: 见 docs/guide/migration.md"
  echo
  echo "  查看日志: docker logs -f ${CONTAINER_NAME}"
  echo "  停止服务: docker stop ${CONTAINER_NAME}"
  exit 0
fi

# ── binary 模式 ─────────────────────────────────────────────
if [ "$MODE" = "binary" ]; then
  DATA_DIR="${DATA_DIR:-./pb_data}"
  BIN_PATH="${BIN_PATH:-./Nodus}"

  if [ ! -x "${BIN_PATH}" ]; then
    error "未找到可执行文件 ${BIN_PATH}。"
    echo "请先从 https://github.com/sexyfeifan/Nodus/releases 下载对应平台压缩包并解压，"
    echo "或将 Nodus 二进制放到当前目录后重试。也可设置 BIN_PATH=/path/to/Nodus"
    exit 1
  fi

  mkdir -p "${DATA_DIR}"
  info "启动 Nodus（数据目录 ${DATA_DIR}，端口 ${PORT}）..."
  nohup "${BIN_PATH}" serve --http "0.0.0.0:${PORT}" --dir "${DATA_DIR}" \
    > nodus.log 2>&1 &
  echo $! > nodus.pid

  sleep 2
  if kill -0 "$(cat nodus.pid)" 2>/dev/null; then
    info "部署完成。PID 已写入 nodus.pid，日志见 nodus.log"
    echo "  管理面板: http://localhost:${PORT}"
  else
    error "启动失败，查看 nodus.log"
    exit 1
  fi
  exit 0
fi

error "未知 MODE=${MODE}（支持 docker / binary）"
usage
exit 1
