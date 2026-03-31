#!/usr/bin/env bash
# setup-docker.sh
# Jalankan: bash scripts/setup-docker.sh
# Setup Kira menggunakan Docker untuk PostgreSQL
# Kompatibel dengan: Linux, macOS, Git Bash (Windows)

set -e

echo ""
echo "=== Kira Docker Setup ==="
echo ""

# ── 1. Cek Docker tersedia ────────────────────────────────────────────────────
if ! command -v docker &>/dev/null; then
    echo "✗ Docker tidak ditemukan."
    echo "  Install Docker Desktop: https://www.docker.com/products/docker-desktop/"
    exit 1
fi

echo "✓ Docker ditemukan"

# ── 2. Copy .env ──────────────────────────────────────────────────────────────
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "✓ .env dibuat dari .env.example"
else
    echo "  .env sudah ada, dilewati"
fi

# ── 3. Start PostgreSQL via Docker Compose ────────────────────────────────────
echo ""
echo "Menjalankan PostgreSQL via Docker..."
docker compose up -d postgres
echo ""
echo "Menunggu PostgreSQL siap..."
sleep 4

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
