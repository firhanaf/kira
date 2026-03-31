# setup-docker.ps1
# Jalankan: .\scripts\setup-docker.ps1
# Setup Kira menggunakan Docker untuk PostgreSQL

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "=== Kira Docker Setup (Windows) ===" -ForegroundColor Cyan
Write-Host ""

# ── 1. Cek Docker tersedia ────────────────────────────────────────────────────
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "✗ Docker tidak ditemukan." -ForegroundColor Red
    Write-Host "  Install Docker Desktop: https://www.docker.com/products/docker-desktop/" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Docker ditemukan" -ForegroundColor Green

# ── 2. Copy .env ──────────────────────────────────────────────────────────────
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "✓ .env dibuat dari .env.example" -ForegroundColor Green
} else {
    Write-Host "  .env sudah ada, dilewati" -ForegroundColor Gray
}

# ── 3. Start PostgreSQL via Docker Compose ────────────────────────────────────
Write-Host ""
Write-Host "Menjalankan PostgreSQL via Docker..." -ForegroundColor Cyan
docker compose up -d postgres
Write-Host ""
Write-Host "Menunggu PostgreSQL siap..." -ForegroundColor Gray
Start-Sleep -Seconds 4

# ── 4. Cek Go tersedia ────────────────────────────────────────────────────────
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host ""
    Write-Host "✗ Go tidak ditemukan." -ForegroundColor Red
    Write-Host "  Install dari: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Go ditemukan: $(go version)" -ForegroundColor Green

# ── 5. Download dependencies ──────────────────────────────────────────────────
Write-Host ""
Write-Host "Download Go dependencies..." -ForegroundColor Cyan
go mod download
Write-Host "✓ Dependencies siap" -ForegroundColor Green

# ── 6. Instruksi akhir ────────────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✓ Setup selesai!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Langkah selanjutnya:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  1. Edit .env — isi JWT_SECRET (wajib, min 32 karakter):" -ForegroundColor White
Write-Host "       JWT_SECRET=$(New-Guid)$(New-Guid)" -ForegroundColor Gray
Write-Host ""
Write-Host "  2. Jalankan server:" -ForegroundColor White
Write-Host "       go run .\cmd\server\main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "  3. Cek server berjalan:" -ForegroundColor White
Write-Host "       curl http://localhost:8080/health" -ForegroundColor Gray
Write-Host ""
