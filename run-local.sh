#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

echo "╔══════════════════════════════════════╗"
echo "║     HushCircuits Pro - Local Dev     ║"
echo "╚══════════════════════════════════════╝"

# ── Prerequisites ──────────────────────────────
command -v go  >/dev/null 2>&1 || { echo "✗ Go required: https://go.dev/dl/";  exit 1; }
command -v node >/dev/null 2>&1 || { echo "✗ Node required: https://nodejs.org/"; exit 1; }
command -v docker >/dev/null 2>&1 || echo "⚠ Docker not found — Postgres/Redis won't start"

# ── Env ────────────────────────────────────────
[ -f .env ] && set -a && source .env && set +a

export PORT="${PORT:-8080}"
export DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5433/hushcircuits?sslmode=disable}"
export GENSMS_API_KEY="${GENSMS_API_KEY:-demo}"
export FEATHERLESS_API_KEY="${FEATHERLESS_API_KEY:-demo}"
export API_AUTH_TOKEN="${API_AUTH_TOKEN:-}"

# ── Docker services ────────────────────────────
echo ""
echo "[1/4] Starting Postgres & Redis…"
docker compose up -d postgres redis 2>/dev/null
for i in $(seq 1 30); do
  docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1 && break
  [ "$i" -eq 30 ] && { echo "✗ Postgres not ready"; exit 1; }
  sleep 1
done
echo "  ✓ Postgres :5433"
echo "  ✓ Redis    :6380"

# ── Build backend ──────────────────────────────
echo ""
echo "[2/4] Building backend…"
cd "$ROOT/backend"
go mod tidy 2>/dev/null
go build -o /tmp/hushcircuits-api ./cmd/api
echo "  ✓ Binary at /tmp/hushcircuits-api"

# ── Start backend ──────────────────────────────
echo ""
echo "[3/4] Starting backend on :$PORT…"
/tmp/hushcircuits-api &
BPID=$!
for i in $(seq 1 10); do
  sleep 1
  curl -sf http://localhost:$PORT/health >/dev/null && break
done
curl -sf http://localhost:$PORT/health >/dev/null \
  && echo "  ✓ Backend PID $BPID running" \
  || { echo "  ✗ Backend failed"; exit 1; }

# ── Start frontend ─────────────────────────────
echo ""
echo "[4/4] Starting frontend on :3000…"
cd "$ROOT/frontend"
npm install --silent 2>/dev/null
npx next dev -p 3000 > /tmp/frontend.log 2>&1 &
FPID=$!
sleep 5
curl -sf http://localhost:3000 >/dev/null \
  && echo "  ✓ Frontend PID $FPID running" \
  || echo "  ⚠ Frontend may still be compiling (check /tmp/frontend.log)"

# ── Done ───────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════╗"
echo "║  READY                              ║"
echo "║──────────────────────────────────────║"
echo "║  Frontend  http://localhost:3000     ║"
echo "║  Backend   http://localhost:$PORT/health ║"
echo "║──────────────────────────────────────║"
echo "║  Ctrl+C to stop all services        ║"
echo "╚══════════════════════════════════════╝"

trap "kill $BPID $FPID 2>/dev/null; docker compose stop >/dev/null; echo 'Stopped.'" SIGINT SIGTERM
wait
