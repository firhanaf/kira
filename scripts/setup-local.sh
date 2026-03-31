#!/usr/bin/env bash
# setup-local.sh
# Jalankan: bash scripts/setup-local.sh
# Setup Kira tanpa Docker — butuh PostgreSQL terinstall lokal
# Kompatibel dengan: Linux, macOS, Git Bash (Windows)

set -e

echo ""
echo "=== Kira Local Setup ==="
echo ""

# ── 1. Cek psql tersedia ──────────────────────────────────────────────────────
if ! command -v psql &>/dev/null; then
    echo "✗ psql tidak ditemukan. Install PostgreSQL terlebih dahulu."
    echo ""
    case "$(uname -s)" in
        Darwin)  echo "  macOS  : brew install postgresql@16 && brew services start postgresql@16" ;;
        Linux)   echo "  Linux  : sudo apt install postgresql  (atau dnf/pacman)" ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "  Windows: https://www.postgresql.org/download/windows/"
            echo "  Pastikan centang 'Add to PATH' saat instalasi,"
            echo "  atau tambahkan manual: C:/Program Files/PostgreSQL/<versi>/bin"
            ;;
    esac
    echo ""
    exit 1
fi

echo "✓ psql ditemukan: $(command -v psql)"

# ── 2. Copy .env ──────────────────────────────────────────────────────────────
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "✓ .env dibuat dari .env.example"
else
    echo "  .env sudah ada, dilewati"
fi

# ── 3. Buat user & database ───────────────────────────────────────────────────
echo ""
echo "Membuat PostgreSQL user dan database..."
echo "  (mungkin diminta password postgres superuser)"
echo ""

psql -U postgres <<-SQL
    DO \$\$ BEGIN
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'kira') THEN
            CREATE USER kira WITH PASSWORD 'kira_secret';
        END IF;
    END \$\$;
SQL

psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'kira_db'" \
    | grep -q 1 || psql -U postgres -c "CREATE DATABASE kira_db OWNER kira;"

psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE kira_db TO kira;"

echo ""
echo "✓ User 'kira' dan database 'kira_db' siap"

# ── 4. Cek Go tersedia ────────────────────────────────────────────────────────
if ! command -v go &>/dev/null; then
    echo ""
    echo "✗ Go tidak ditemukan. Install dari: https://go.dev/dl/"
    exit 1
fi

echo "✓ Go ditemukan: $(go version)"

# ── 5. Download dependencies ──────────────────────────────────────────────────
echo ""
echo "Download Go dependencies..."
go mod download
echo "✓ Dependencies siap"

# ── 6. Generate JWT_SECRET otomatis di .env jika masih placeholder ────────────
if grep -q "replace-with-a-very-long-random-secret" .env; then
    SECRET=$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 48)
    sed -i "s|replace-with-a-very-long-random-secret-at-least-256-bits|${SECRET}|" .env
    echo "✓ JWT_SECRET di-generate otomatis di .env"
fi

# ── 7. Instruksi akhir ────────────────────────────────────────────────────────
echo ""
echo "========================================"
echo "✓ Setup selesai!"
echo "========================================"
echo ""
echo "Jalankan server:"
echo "  go run ./cmd/server/main.go"
echo ""
echo "Cek server berjalan:"
echo "  curl http://localhost:8080/health"
echo ""
