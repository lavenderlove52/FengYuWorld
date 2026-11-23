#!/usr/bin/env bash
# 后端一键发布：拉代码 → 迁移 → 编译 → 重启 → 健康检查（失败自动回滚）
#
# 用法（在仓库根目录执行）：
#   ./deploy/release.sh              # 完整发布
#   ./deploy/release.sh --no-migrate # 跳过数据库迁移
#   ./deploy/release.sh --dry-run    # 只打印将要执行的操作
#
# 退出码：0 成功；非 0 表示中途失败（失败时不会留下半更新的服务）

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$REPO_DIR/backend"
SERVICE="${SERVICE_NAME:-myblog}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/api/health}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-15}"

DO_MIGRATE=1
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --no-migrate) DO_MIGRATE=0 ;;
    --dry-run)    DRY_RUN=1 ;;
    -h|--help)
      sed -n '2,10p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "未知参数：$arg（用 --help 查看用法）" >&2
      exit 2
      ;;
  esac
done

step() { printf '\n\033[1;34m==> %s\033[0m\n' "$*"; }
ok()   { printf '\033[1;32m  ✓ %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33m  ! %s\033[0m\n' "$*"; }
die()  { printf '\033[1;31m  ✗ %s\033[0m\n' "$*" >&2; exit 1; }

run() {
  if [ "$DRY_RUN" -eq 1 ]; then
    printf '  [dry-run] %s\n' "$*"
  else
    "$@"
  fi
}

step "检查工作区"
cd "$REPO_DIR"
if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  git status --short
  die "服务器上有未提交的改动，先处理干净再发布（或手动 stash）"
fi
ok "工作区干净，当前 $(git rev-parse --short HEAD) $(git log -1 --pretty=%s)"

step "拉取代码"
BEFORE="$(git rev-parse HEAD)"
run git pull --ff-only || die "git pull 失败（可能有本地提交或需要手动合并）"
AFTER="$(git rev-parse HEAD)"
if [ "$BEFORE" = "$AFTER" ]; then
  warn "代码没有变化，继续执行迁移与重启"
else
  ok "已更新到 $(git rev-parse --short HEAD)"
  git log --oneline "$BEFORE..$AFTER" | sed 's/^/    /'
fi

cd "$BACKEND_DIR"

if [ "$DO_MIGRATE" -eq 1 ]; then
  step "数据库迁移"
  # graph 迁移失败会直接退出，不会继续编译和重启
  run go run ./cmd/migrate || die "迁移失败，已中止（服务仍在运行旧版本）"
  ok "迁移完成"
else
  warn "已跳过数据库迁移（--no-migrate）"
fi

step "编译"
# 先编到临时文件：编译失败时线上二进制原样不动
run go build -o bin/server.new ./cmd/server || die "编译失败，已中止（服务仍在运行旧版本）"
ok "编译成功"

step "备份并替换二进制"
if [ "$DRY_RUN" -eq 0 ]; then
  if [ -f bin/server ]; then
    cp -p bin/server bin/server.bak
    ok "已备份旧版本到 bin/server.bak"
  else
    warn "没有找到旧二进制，跳过备份（首次部署？）"
  fi
  mv bin/server.new bin/server
  chmod +x bin/server
else
  printf '  [dry-run] cp -p bin/server bin/server.bak && mv bin/server.new bin/server\n'
fi

rollback() {
  if [ -f bin/server.bak ]; then
    warn "回滚到上一个版本"
    mv bin/server.bak bin/server
    systemctl restart "$SERVICE" || true
    warn "已回滚，请查看 journalctl -u $SERVICE -n 50"
  else
    warn "没有可回滚的备份"
  fi
}

step "重启服务"
run systemctl restart "$SERVICE" || { rollback; die "重启失败"; }

if [ "$DRY_RUN" -eq 1 ]; then
  printf '\n[dry-run] 到此为止，未实际重启与检查健康状态\n'
  exit 0
fi

step "健康检查"
healthy=0
for i in $(seq 1 "$HEALTH_TIMEOUT"); do
  if curl -sf -o /dev/null --max-time 2 "$HEALTH_URL"; then
    healthy=1
    break
  fi
  sleep 1
done

if [ "$healthy" -ne 1 ]; then
  warn "服务在 ${HEALTH_TIMEOUT}s 内没有通过健康检查"
  systemctl status "$SERVICE" --no-pager -l | tail -20 || true
  rollback
  die "发布失败，已回滚"
fi

ok "健康检查通过（$HEALTH_URL）"
step "完成"
ok "发布成功：$(git rev-parse --short HEAD)"
printf '  日志：journalctl -u %s -f\n' "$SERVICE"
