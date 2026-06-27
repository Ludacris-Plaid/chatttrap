#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════════╗
# ║  HushCircuits Pro v2 — Cyberpunk TUI Startup                    ║
# ║  "The Ethical Hacker's Voice"                                   ║
# ╚══════════════════════════════════════════════════════════════════╝
set -o pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

# ─── ANSI Pallete (matches globals.css + tailwind.config.ts) ───
if [[ -t 1 ]]; then
  C_RESET='\033[0m'
  C_BOLD='\033[1m'     C_DIM='\033[2m'       C_ITALIC='\033[3m'
  C_RED='\033[38;5;160m'    C_RED_BG='\033[48;5;160m'    C_RED_DIM='\033[38;5;124m'
  C_GREEN='\033[38;5;42m'   C_GREEN_BG='\033[48;5;42m'
  C_WHITE='\033[38;5;255m'  C_GRAY='\033[38;5;243m'      C_DGRAY='\033[38;5;239m'
  C_PURPLE='\033[38;5;99m'  C_AMBER='\033[38;5;214m'
  C_HIDE_CURSOR='\033[?25l'  C_SHOW_CURSOR='\033[?25h'
  C_CLEAR_LINE='\033[2K'     C_UP='\033[1A'
else
  C_RESET= C_BOLD= C_DIM= C_ITALIC= C_RED= C_RED_BG= C_RED_DIM=
  C_GREEN= C_GREEN_BG= C_WHITE= C_GRAY= C_DGRAY=
  C_PURPLE= C_AMBER= C_HIDE_CURSOR= C_SHOW_CURSOR= C_CLEAR_LINE= C_UP=
fi

# ─── Kill switch ────────────────────────────────────────────────
_cleanup() {
  if [[ -n "$BPID" ]]; then kill "$BPID" 2>/dev/null; fi
  if [[ -n "$FPID" ]]; then kill "$FPID" 2>/dev/null; fi
  echo -e "\n${C_RED}[−]${C_RESET} All services stopped."
}
trap '_cleanup; echo -e "$C_SHOW_CURSOR"; exit 0' SIGINT SIGTERM EXIT

# ─── Spinner frames ─────────────────────────────────────────────
SPIN=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
spin_i=0

_spin() { printf "${C_GRAY}%s${C_RESET}" "${SPIN[$spin_i]}"; spin_i=$(( (spin_i + 1) % ${#SPIN[@]} )); }

# ─── Progress bar ────────────────────────────────────────────────
_bar() {
  local pct=$1 w=${2:-30} filled=$(( pct * w / 100 )) empty=$(( w - filled ))
  local bar="${C_RED}"
  for ((i=0; i<filled; i++)); do bar+='█'; done
  bar+="${C_DGRAY}"
  for ((i=0; i<empty;  i++)); do bar+='░'; done
  bar+="${C_RESET}"
  printf "%s %3d%%" "$bar" "$pct"
}

# ─── Fancy header ────────────────────────────────────────────────
_header() {
  clear
  echo -e "${C_RESET}"
  echo -e "${C_DIM}╔══════════════════════════════════════════════════════════════════╗${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}                                                                  ${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_RED}██${C_DIM}╗${C_RESET}  ${C_RED}██${C_DIM}╗${C_RESET} ${C_WHITE}██${C_DIM}╗${C_RESET}   ${C_WHITE}██${C_DIM}╗${C_RESET} ${C_WHITE}███████${C_DIM}╗${C_RESET} ${C_RED}██${C_DIM}╗${C_RESET}  ${C_RED}██${C_DIM}╗${C_RESET} ${C_DIM}██████╗${C_RESET}  ${C_WHITE}██${C_DIM}╗${C_RESET}${C_WHITE}██████╗${C_RESET}${C_WHITE}██${C_DIM}╗${C_RESET}${C_WHITE}██${C_DIM}╗${C_RESET}${C_WHITE}███████╗${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_RED}██${C_DIM}║${C_RESET}  ${C_RED}██${C_DIM}║${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}   ${C_WHITE}██${C_DIM}║${C_RESET} ${C_WHITE}██${C_DIM}╔════╝${C_RESET} ${C_RED}██${C_DIM}║${C_RESET} ${C_RED}██${C_DIM}╔╝${C_RESET}${C_DIM}██${C_DIM}╔════╝${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}██${C_DIM}╔══${C_RED}██${C_DIM}╗${C_RESET}${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}██╔════╝${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_RED}███████${C_DIM}║${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}   ${C_WHITE}██${C_DIM}║${C_RESET} ${C_WHITE}███████${C_DIM}╗${C_RESET} ${C_RED}█████${C_DIM}╔╝${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}      ${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}██████${C_DIM}╔╝${C_RESET}${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}███████╗${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_RED}██${C_DIM}╔══${C_RED}██${C_DIM}║${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}   ${C_WHITE}██${C_DIM}║${C_RESET} ${C_DIM}╚════${C_WHITE}██${C_DIM}║${C_RESET} ${C_RED}██${C_DIM}╔${C_RED}██${C_DIM}╗${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}      ${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}██${C_DIM}╔═══╝${C_RESET} ${C_WHITE}██${C_DIM}║╚════${C_WHITE}██${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_RED}██${C_DIM}║${C_RESET}  ${C_RED}██${C_DIM}║${C_RESET} ${C_DIM}╚${C_WHITE}██████╔╝${C_RESET} ${C_WHITE}███████${C_DIM}║${C_RESET} ${C_RED}██${C_DIM}║${C_RESET}  ${C_RED}██${C_DIM}╗╚${C_WHITE}██████╗${C_RESET}${C_DIM}╚${C_WHITE}██████╔╝${C_RESET} ${C_WHITE}██${C_DIM}║${C_RESET}${C_WHITE}███████║${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}   ${C_DIM}╚═╝${C_RESET}  ${C_DIM}╚═╝${C_RESET}  ${C_DIM}╚═════╝${C_RESET} ${C_DIM}╚══════╝${C_RESET} ${C_DIM}╚═╝${C_RESET}  ${C_DIM}╚═╝${C_RESET} ${C_DIM}╚═════╝${C_RESET}  ${C_DIM}╚═════╝${C_RESET} ${C_DIM}╚═╝${C_RESET}${C_DIM}╚══════╝${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}                                                                  ${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}║${C_RESET}    ${C_RED}▸${C_RESET} ${C_BOLD}${C_WHITE}PRO v2.0${C_RESET}${C_DIM} — The Ethical Hacker's Voice${C_RESET}                         ${C_DIM}║${C_RESET}"
  echo -e "${C_DIM}╚══════════════════════════════════════════════════════════════════╝${C_RESET}"
  echo
}

# ─── Step printer ────────────────────────────────────────────────
_step() {
  local icon="$1" label="$2" detail="${3:-}"
  printf "${C_DGRAY}[${C_RESET}${icon}${C_DGRAY}]${C_RESET} ${C_BOLD}%-38s${C_RESET}" "$label"
  [[ -n "$detail" ]] && printf " ${C_DGRAY}%s${C_RESET}" "$detail"
}

_step_ok()   { echo -e "  ${C_GREEN}✓${C_RESET}"; }
_step_warn() { echo -e "  ${C_AMBER}⚠${C_RESET}  $1"; }
_step_fail() { echo -e "  ${C_BOLD}${C_RED}✗${C_RESET}  $1"; }

# ─── Fancy status line (used by spinner coroutine) ────────────────
_spin_status() {
  local msg="$1"
  printf "\r${C_CLEAR_LINE}  %s ${C_DGRAY}%s${C_RESET}" "$(_spin)" "$msg"
}

# ─── Animated wait with spinner ───────────────────────────────────
_wait_with_spin() {
  local timeout_secs=$1 msg="$2" i=0
  while [[ $i -lt $timeout_secs ]]; do
    _spin_status "$msg ($(( timeout_secs - i ))s)"
    sleep 1
    ((i++))
  done
  printf "\r${C_CLEAR_LINE}"
}

# ─── Section banner ───────────────────────────────────────────────
_section() {
  echo
  echo -e "${C_RED}▐${C_DIM}━━━${C_RESET} ${C_BOLD}${C_WHITE}$1${C_RESET} ${C_DIM}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 0 — Boot Screen & Environment
# ═══════════════════════════════════════════════════════════════════
_boot_screen() {
  _header
  echo -e "$C_HIDE_CURSOR"

  # Flicker effect
  for i in {1..3}; do
    echo -e "  ${C_RED}▌${C_RESET} ${C_DGRAY}INITIALIZING HUSHCIRCUITS KERNEL...${C_RESET}"
    sleep 0.03
    echo -e "\r${C_CLEAR_LINE}  ${C_RED}▌${C_RESET} ${C_DGRAY}INITIALIZING HUSHCIRCUITS KERNEL..${C_RESET}"
    sleep 0.03
    echo -e "\r${C_CLEAR_LINE}"
  done

  # Fake boot messages
  local boot_msgs=(
    "${C_DGRAY}  ├─ memory mapped [0x0000..0xFFFF]${C_RESET}"
    "${C_DGRAY}  ├─ entropy pool seeded${C_RESET}"
    "${C_DGRAY}  ├─ signal handlers installed${C_RESET}"
    "${C_GREEN}  ├─ secure enclave verified${C_RESET}"
    "${C_GREEN}  ├─ TPM attestation: OK${C_RESET}"
    "${C_RED}  └─ REDCORE engine online${C_RESET}"
  )
  for msg in "${boot_msgs[@]}"; do
    echo -e "$msg"
    sleep 0.05
  done
  sleep 0.3
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 1 — Prerequisite Scan
# ═══════════════════════════════════════════════════════════════════
_scan_prereqs() {
  _section "SYSTEM SCAN"
  local all_ok=true

  # go
  _step "🔍" "Scanning Go toolchain..."
  if command -v go &>/dev/null; then
    local gover; gover=$(go version 2>/dev/null | grep -oP 'go\K[0-9.]+')
    _step_ok
    echo -e "  ${C_DGRAY}    └─${C_RESET} go ${C_GREEN}${gover}${C_RESET} ${C_DGRAY}@${C_RESET} $(which go)"
  else
    _step_fail "Go not found — install from https://go.dev/dl/"
    all_ok=false
  fi

  # node
  _step "🔍" "Scanning Node.js runtime..."
  if command -v node &>/dev/null; then
    local nver; nver=$(node --version)
    _step_ok
    echo -e "  ${C_DGRAY}    └─${C_RESET} node ${C_GREEN}${nver}${C_RESET} ${C_DGRAY}@${C_RESET} $(which node)"
  else
    _step_fail "Node.js not found — install from https://nodejs.org/"
    all_ok=false
  fi

  # docker
  _step "🔍" "Scanning Docker daemon..."
  if docker info &>/dev/null 2>&1; then
    _step_ok
    echo -e "  ${C_DGRAY}    └─${C_RESET} docker ${C_GREEN}available${C_RESET} (containers: $(docker ps -q 2>/dev/null | wc -l) running)"
  else
    _step_warn "Docker not found — Postgres/Redis will need manual setup"
  fi

  # psql
  _step "🔍" "Scanning PostgreSQL client..."
  if command -v psql &>/dev/null; then
    _step_ok
  else
    _step_warn "psql not found — cannot run migrations automatically"
  fi

  # curl
  _step "🔍" "Scanning curl..."
  command -v curl &>/dev/null && _step_ok || { _step_warn "curl not found — health checks will be skipped"; }

  # ports
  _step "🔍" "Scanning port availability..."
  local ports_ok=true
  if ss -tln | grep -q ':8080\b'; then
    _step_warn "Port 8080 in use — backend may fail"
    ports_ok=false
  fi
  if ss -tln | grep -q ':3000\b'; then
    _step_warn "Port 3000 in use — frontend may fail"
    ports_ok=false
  fi
  $ports_ok && _step_ok

  $all_ok || { echo; echo -e "  ${C_BOLD}${C_RED}✗ Missing prerequisites — install them and re-run.${C_RESET}"; exit 1; }
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 2 — Interactive Configuration
# ═══════════════════════════════════════════════════════════════════
_prompt_config() {
  _section "CONFIGURATION"

  # Detect terminal width
  local tw; tw=$(tput cols 2>/dev/null || echo 80)

  # ─── Helper: prompt with default ───
  _ask() {
    local prompt="$1" default="$2" var="$3"
    printf "  ${C_RED}▸${C_RESET} ${C_WHITE}%s${C_RESET}" "$prompt"
    if [[ -n "$default" ]]; then
      printf " ${C_DGRAY}[%s]${C_RESET}: " "$default"
    else
      printf ": "
    fi
    read -r REPLY
    if [[ -z "$REPLY" && -n "$default" ]]; then REPLY="$default"; fi
    printf -v "$var" "%s" "$REPLY"
  }

  _ask_yn() {
    local prompt="$1" default="$2" var="$3"
    local yn_hint
    [[ "$default" == "y" ]] && yn_hint="Y/n" || yn_hint="y/N"
    printf "  ${C_RED}▸${C_RESET} ${C_WHITE}%s${C_RESET} ${C_DGRAY}[%s]${C_RESET}: " "$prompt" "$yn_hint"
    read -r REPLY
    REPLY="${REPLY,,}"
    [[ -z "$REPLY" ]] && REPLY="$default"
    [[ "$REPLY" == "y" || "$REPLY" == "yes" ]] && printf -v "$var" "true" || printf -v "$var" "false"
  }

  # Load existing .env as defaults
  local cfg_db_url="postgres://postgres:postgres@localhost:5433/hushcircuits?sslmode=disable"
  local cfg_redis_url="redis://localhost:6380/0"
  local cfg_port="8080"
  local cfg_admin_email="admin@hushcircuits.io"

  if [[ -f "$ROOT/.env" ]]; then
    # Extract values safely (handle = in values)
    local v; v=$(grep -E '^DATABASE_URL=' "$ROOT/.env" 2>/dev/null | head -1 | sed 's/^DATABASE_URL=//')
    [[ -n "$v" ]] && cfg_db_url="$v"
    v=$(grep -E '^REDIS_URL=' "$ROOT/.env" 2>/dev/null | head -1 | sed 's/^REDIS_URL=//')
    [[ -n "$v" ]] && cfg_redis_url="$v"
    v=$(grep -E '^PORT=' "$ROOT/.env" 2>/dev/null | head -1 | sed 's/^PORT=//')
    [[ -n "$v" ]] && cfg_port="$v"
    v=$(grep -E '^ADMIN_EMAIL=' "$ROOT/.env" 2>/dev/null | head -1 | sed 's/^ADMIN_EMAIL=//')
    [[ -n "$v" ]] && cfg_admin_email="$v"
  fi

  echo -e "  ${C_DGRAY}┌─ Quick Setup ──────────────────────────────────────────────┐${C_RESET}"
  echo

  _ask "Admin email (for admin panel access)" "$cfg_admin_email" cfg_admin_email
  _ask "Backend port" "$cfg_port" cfg_port
  _ask "Database URL" "$cfg_db_url" cfg_db_url
  _ask "Redis URL" "$cfg_redis_url" cfg_redis_url
  _ask_yn "Use Docker for Postgres/Redis?" "y" cfg_use_docker
  _ask_yn "Start frontend (Next.js dev server)?" "y" cfg_start_frontend
  _ask_yn "Enable demo mode (mock external APIs)?" "y" cfg_demo_mode
  _ask_yn "Connect SIP telephony trunk?" "n" cfg_enable_sip

  echo
  echo -e "  ${C_DGRAY}└───────────────────────────────────────────────────────────┘${C_RESET}"

  # Export for later phases
  CFG_ADMIN_EMAIL="$cfg_admin_email"
  CFG_PORT="$cfg_port"
  CFG_DB_URL="$cfg_db_url"
  CFG_REDIS_URL="$cfg_redis_url"
  CFG_USE_DOCKER="$cfg_use_docker"
  CFG_START_FRONTEND="$cfg_start_frontend"
  CFG_DEMO_MODE="$cfg_demo_mode"
  CFG_ENABLE_SIP="$cfg_enable_sip"
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 3 — Infrastructure
# ═══════════════════════════════════════════════════════════════════
_spin_infrastructure() {
  _section "INFRASTRUCTURE"

  local db_ok=false redis_ok=false

  if [[ "$CFG_USE_DOCKER" == "true" ]]; then
    # ─── Docker ───
    _step "🐳" "Launching Docker containers..."
    docker compose up -d postgres redis &>/dev/null
    _step_ok

    # ─── Postgres health ───
    _step "🐘" "Waiting for PostgreSQL..."
    for i in $(seq 1 30); do
      _spin_status "PostgreSQL health check ($i/30)"
      if docker compose exec -T postgres pg_isready -U postgres &>/dev/null 2>&1; then
        printf "\r${C_CLEAR_LINE}  ${C_GREEN}✔${C_RESET} ${C_BOLD}%-36s${C_RESET}  ${C_GREEN}✓${C_RESET} %s\n" "PostgreSQL" "healthy (port 5433)"
        db_ok=true
        break
      fi
      sleep 1
    done
    if ! $db_ok; then
      printf "\r${C_CLEAR_LINE}"
      _step_fail "PostgreSQL failed to start"
    fi

    # ─── Redis health ───
    _step "🔴" "Waiting for Redis..."
    for i in $(seq 1 15); do
      _spin_status "Redis health check ($i/15)"
      if docker compose exec -T redis redis-cli ping &>/dev/null 2>&1; then
        printf "\r${C_CLEAR_LINE}  ${C_GREEN}✔${C_RESET} ${C_BOLD}%-36s${C_RESET}  ${C_GREEN}✓${C_RESET} %s\n" "Redis" "healthy (port 6380)"
        redis_ok=true
        break
      fi
      sleep 1
    done
    if ! $redis_ok; then
      printf "\r${C_CLEAR_LINE}"
      _step_fail "Redis failed to start"
    fi
  else
    _step "🔌" "Skipping Docker — assuming external Postgres & Redis"
    _step_ok
    db_ok=true; redis_ok=true
  fi

  if ! $db_ok || ! $redis_ok; then
    echo; echo -e "  ${C_BOLD}${C_RED}✗ Infrastructure not ready — cannot continue.${C_RESET}"; exit 1
  fi
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 4 — Database Migrations
# ═══════════════════════════════════════════════════════════════════
_spin_migrations() {
  _section "DATABASE"

  if ! command -v psql &>/dev/null; then
    _step "🗄️ " "Skipping migrations — psql not available"
    _step_warn "Run backend/migrations/001_schema.sql manually"
    return
  fi

  _step "🗄️ " "Applying schema migration..."
  _spin_status "Running 001_schema.sql..."

  local mig_out
  # Extract password from DATABASE_URL if present
  local pg_pass="postgres"
  if [[ "$CFG_DB_URL" =~ :([^:@]+)@ ]]; then
    pg_pass="${BASH_REMATCH[1]}"
  fi
  mig_out=$(PGPASSWORD="$pg_pass" psql "$CFG_DB_URL" -f "$ROOT/backend/migrations/001_schema.sql" 2>&1)
  local mig_rc=$?

  printf "\r${C_CLEAR_LINE}"
  if [[ $mig_rc -eq 0 ]]; then
    local tables; tables=$(echo "$mig_out" | grep -c "CREATE TABLE")
    _step "🗄️ " "Schema applied ($tables tables)"
    _step_ok
  else
    _step_fail "Migration failed — check database connectivity"
    echo -e "  ${C_DGRAY}    └─${C_RESET} $mig_out"
    exit 1
  fi
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 5 — Build
# ═══════════════════════════════════════════════════════════════════
_spin_build() {
  _section "COMPILATION"

  # ─── Backend ───
  _step "🔧" "Building Go backend..."
  _spin_status "go build ./cmd/api ..."
  (
    cd "$ROOT/backend"
    go mod tidy &>/dev/null
    go build -o /tmp/hushcircuits-api ./cmd/api 2>/tmp/hush-build-err.log
  )
  local rc=$?
  printf "\r${C_CLEAR_LINE}"
  if [[ $rc -eq 0 ]]; then
    _step "🔧" "Backend binary"
    _step_ok
    echo -e "  ${C_DGRAY}    └─${C_RESET} /tmp/hushcircuits-api ${C_GREEN}$(du -h /tmp/hushcircuits-api | cut -f1)${C_RESET}"
  else
    _step_fail "Go build failed"
    echo -e "  ${C_DGRAY}    └─${C_RESET} $(cat /tmp/hush-build-err.log 2>/dev/null)"
    exit 1
  fi

  if [[ "$CFG_START_FRONTEND" == "true" ]]; then
    # ─── Frontend ───
    _step "🎨" "Installing frontend dependencies..."
    _spin_status "npm install ..."
    (
      cd "$ROOT/frontend"
      npm install --silent &>/tmp/hush-npm.log
    )
    local frc=$?
    printf "\r${C_CLEAR_LINE}"
    if [[ $frc -eq 0 ]]; then
      _step "🎨" "Frontend dependencies"
      _step_ok
    else
      _step_warn "npm install had warnings — may still work"
    fi

    # TypeScript check
    _step "🔬" "TypeScript type-check..."
    _spin_status "tsc --noEmit ..."
    (
      cd "$ROOT/frontend"
      npx tsc --noEmit &>/tmp/hush-tsc.log
    )
    local tsc_rc=$?
    printf "\r${C_CLEAR_LINE}"
    if [[ $tsc_rc -eq 0 ]]; then
      _step "🔬" "TypeScript"
      _step_ok
    else
      _step_warn "TypeScript warnings — frontend may still compile"
    fi
  fi
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 6 — Launch Services
# ═══════════════════════════════════════════════════════════════════
_spin_launch() {
  _section "IGNITION"

  # Kill any previous instance on our port
  fuser -k "${CFG_PORT}/tcp" &>/dev/null 2>&1
  sleep 0.5

  # Build env
  local env_vars=(
    "DATABASE_URL=$CFG_DB_URL"
    "REDIS_URL=$CFG_REDIS_URL"
    "PORT=$CFG_PORT"
    "ADMIN_EMAIL=$CFG_ADMIN_EMAIL"
    "ALLOWED_ORIGINS=http://localhost:3000"
  )

  # Load API keys from .env if present (handle = signs in values)
  if [[ -f "$ROOT/.env" ]]; then
    while IFS='=' read -r k v; do
      [[ -z "$k" || "$k" =~ ^[[:space:]]*# ]] && continue
      env_vars+=("$k=$v")
    done < <(grep -E '^(FEATHERLESS|GENSMS|CNAM|NOWPAYMENTS|JWT_SECRET|SUPABASE|SIP_)' "$ROOT/.env" 2>/dev/null)
  fi

  # Override with demo mode if selected
  if [[ "$CFG_DEMO_MODE" == "true" ]]; then
    env_vars+=("NOWPAYMENTS_API_KEY=")
    env_vars+=("NOWPAYMENTS_IPN_SECRET=")
  fi

  # Disable SIP if not requested
  if [[ "$CFG_ENABLE_SIP" != "true" ]]; then
    env_vars+=("SIP_USERNAME=")
    env_vars+=("SIP_PASSWORD=")
    env_vars+=("SIP_HOST=")
  fi

  # ─── Launch backend ───
  _step "🚀" "Launching backend..."
  env "${env_vars[@]}" nohup /tmp/hushcircuits-api &>/tmp/hushcircuits-api.log &
  BPID=$!
  sleep 2

  # Health check
  local be_ok=false
  for i in $(seq 1 10); do
    _spin_status "Backend health check ($i/10)"
    if curl -sf "http://localhost:$CFG_PORT/health" &>/dev/null; then
      printf "\r${C_CLEAR_LINE}"
      _step "🚀" "Backend  → http://localhost:$CFG_PORT"
      _step_ok
      echo -e "  ${C_DGRAY}    └─${C_RESET} PID ${C_GREEN}$BPID${C_RESET} ${C_DGRAY}·${C_RESET} v${C_GREEN}$(curl -s "http://localhost:$CFG_PORT/health" | grep -oP '"version":"\K[^"]+')${C_RESET}"
      be_ok=true
      break
    fi
    sleep 1
  done
  if ! $be_ok; then
    printf "\r${C_CLEAR_LINE}"
    _step_fail "Backend failed to start — check /tmp/hushcircuits-api.log"
    exit 1
  fi

  # ─── Launch frontend ───
  if [[ "$CFG_START_FRONTEND" == "true" ]]; then
    _step "🌐" "Launching frontend..."
    (
      cd "$ROOT/frontend"
      nohup npx next dev -p 3000 &>/tmp/hush-frontend.log &
      echo $! > /tmp/hush-frontend.pid
    )
    FPID=$(cat /tmp/hush-frontend.pid 2>/dev/null || echo "")
    local fe_ok=false
    for i in $(seq 1 20); do
      _spin_status "Frontend compiling ($i/20)"
      if curl -sf "http://localhost:3000" &>/dev/null; then
        printf "\r${C_CLEAR_LINE}"
        _step "🌐" "Frontend → http://localhost:3000"
        _step_ok
        echo -e "  ${C_DGRAY}    └─${C_RESET} PID ${C_GREEN}$FPID${C_RESET} ${C_DGRAY}·${C_RESET} Next.js dev server"
        fe_ok=true
        break
      fi
      sleep 2
    done
    if ! $fe_ok; then
      printf "\r${C_CLEAR_LINE}"
      _step "🌐" "Frontend → http://localhost:3000"
      _step_warn "Still compiling — check /tmp/hush-frontend.log"
      echo -e "  ${C_DGRAY}    └─${C_RESET} May take 30-60s on first run"
    fi
  fi
}

# ═══════════════════════════════════════════════════════════════════
#  PHASE 7 — Grand Finale
# ═══════════════════════════════════════════════════════════════════
_grand_finale() {
  sleep 0.5
  clear
  local tw; tw=$(tput cols 2>/dev/null || echo 80)

  echo
  echo -e "${C_RED}         ╔══════════════════════════════════════════════╗"
  echo -e "${C_RED}   ▄▄▄▄▄▄▄║${C_RESET}  ${C_BOLD}${C_WHITE}HUSHCIRCUITS PRO v2.0  —  ONLINE${C_RESET}           ${C_RED}║"
  echo -e "${C_RED}   █${C_WHITE}▓▒░ SYSTEM ░▒▓${C_RED}█║${C_RESET}                                              ${C_RED}║"
  echo -e "${C_RED}   ▀▀▀▀▀▀▀║${C_RESET}  ${C_GREEN}●${C_RESET} Backend   ${C_DIM}→${C_RESET}  http://localhost:${CFG_PORT}${C_RESET}        ${C_RED}║"

  if [[ "$CFG_START_FRONTEND" == "true" ]]; then
    echo -e "${C_RED}           ║${C_RESET}  ${C_GREEN}●${C_RESET} Frontend  ${C_DIM}→${C_RESET}  http://localhost:3000${C_RESET}        ${C_RED}║"
  fi

  echo -e "${C_RED}           ║${C_RESET}  ${C_GREEN}●${C_RESET} PostgreSQL ${C_DIM}→${C_RESET} localhost:5433${C_RESET}          ${C_RED}║"
  echo -e "${C_RED}           ║${C_RESET}  ${C_GREEN}●${C_RESET} Redis     ${C_DIM}→${C_RESET}  localhost:6380${C_RESET}           ${C_RED}║"
  echo -e "${C_RED}           ║${C_RESET}                                              ${C_RED}║"
  echo -e "${C_RED}           ║${C_RESET}  Login with any email (auto-creates profile)${C_RED}║"
  echo -e "${C_RED}           ║${C_RESET}  Admin: ${C_AMBER}${CFG_ADMIN_EMAIL}${C_RESET}                       ${C_RED}║"
  echo -e "${C_RED}           ╚══════════════════════════════════════════════╝${C_RESET}"

  echo
  echo -e "  ${C_DGRAY}╭─ Hotkeys ─────────────────────────────────────────────╮${C_RESET}"
  echo -e "  ${C_DGRAY}│${C_RESET}  ${C_RED}Ctrl+C${C_RESET}   Shutdown all services                      ${C_DGRAY}│${C_RESET}"
  echo -e "  ${C_DGRAY}│${C_RESET}  ${C_RED}h${C_RESET}         Health check endpoint                     ${C_DGRAY}│${C_RESET}"
  echo -e "  ${C_DGRAY}│${C_RESET}  ${C_RED}l${C_RESET}         Tail backend logs                          ${C_DGRAY}│${C_RESET}"
  echo -e "  ${C_DGRAY}│${C_RESET}  ${C_RED}f${C_RESET}         Tail frontend logs                         ${C_DGRAY}│${C_RESET}"
  echo -e "  ${C_DGRAY}│${C_RESET}  ${C_RED}s${C_RESET}         Service status                             ${C_DGRAY}│${C_RESET}"
  echo -e "  ${C_DGRAY}╰──────────────────────────────────────────────────────╯${C_RESET}"
  echo
  echo -e "  ${C_GREEN}■${C_RESET} System operational. ${C_RED}REDCORE${C_RESET} engine online."
  echo

  # Interactive hotkey loop
  echo -e "$C_SHOW_CURSOR"
  while true; do
    printf "${C_DGRAY}HUSHCIRCUITS${C_RESET}${C_RED}»${C_RESET} "
    read -r -n 1 -t 1 key
    case "$key" in
      h|H) echo; curl -s "http://localhost:$CFG_PORT/health" | python3 -m json.tool 2>/dev/null || echo -e "${C_RED}Backend unreachable${C_RESET}"; echo ;;
      l|L) echo; echo -e "${C_DGRAY}─── Backend Logs (tail) ───${C_RESET}"; tail -30 /tmp/hushcircuits-api.log 2>/dev/null; echo ;;
      f|F) echo; echo -e "${C_DGRAY}─── Frontend Logs (tail) ───${C_RESET}"; tail -30 /tmp/hush-frontend.log 2>/dev/null; echo ;;
      s|S) echo; echo -e "  Backend:  $(curl -sf "http://localhost:$CFG_PORT/health" &>/dev/null && echo "${C_GREEN}ONLINE${C_RESET}" || echo "${C_RED}OFFLINE${C_RESET}")"
           echo -e "  Frontend: $(curl -sf http://localhost:3000 &>/dev/null && echo "${C_GREEN}ONLINE${C_RESET}" || echo "${C_RED}OFFLINE${C_RESET}")"
           echo -e "  Postgres: $(docker compose exec -T postgres pg_isready -U postgres &>/dev/null 2>&1 && echo "${C_GREEN}ONLINE${C_RESET}" || echo "${C_RED}OFFLINE${C_RESET}")"
           echo -e "  Redis:    $(docker compose exec -T redis redis-cli ping &>/dev/null 2>&1 && echo "${C_GREEN}ONLINE${C_RESET}" || echo "${C_RED}OFFLINE${C_RESET}")"
           ;;
    esac
  done
}

# ═══════════════════════════════════════════════════════════════════
#  MAIN ENTRY
# ═══════════════════════════════════════════════════════════════════
_boot_screen
_scan_prereqs
_prompt_config

clear
_header
echo -e "  ${C_DGRAY}Configuration locked. Building HushCircuits Pro...${C_RESET}"
sleep 0.3

_spin_infrastructure
_spin_migrations
_spin_build
_spin_launch
_grand_finale
