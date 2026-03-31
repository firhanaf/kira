# setup-local.ps1
# Jalankan: .\scripts\setup-local.ps1
# Setup Kira tanpa Docker — butuh PostgreSQL terinstall lokal

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "=== Kira Local Setup (Windows) ===" -ForegroundColor Cyan
Write-Host ""

# ── 1. Cek psql tersedia ──────────────────────────────────────────────────────
if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
    Write-Host "✗ psql tidak ditemukan." -ForegroundColor Red
    Write-Host ""
    Write-Host "Install PostgreSQL terlebih dahulu:" -ForegroundColor Yellow
    Write-Host "  https://www.postgresql.org/download/windows/"
    Write-Host ""
    Write-Host "Pastikan centang 'Add to PATH' saat instalasi," -ForegroundColor Yellow
    Write-Host "atau tambahkan manual: C:\Program Files\PostgreSQL\<versi>\bin" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ psql ditemukan: $(Get-Command psql | Select-Object -ExpandProperty Source)" -ForegroundColor Green

# ── 2. Copy .env ──────────────────────────────────────────────────────────────
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "✓ .env dibuat dari .env.example" -ForegroundColor Green
} else {
    Write-Host "  .env sudah ada, dilewati" -ForegroundColor Gray
}

# ── 3. Buat user & database PostgreSQL ───────────────────────────────────────
Write-Host ""
Write-Host "Membuat PostgreSQL user dan database..." -ForegroundColor Cyan
Write-Host "  (mungkin diminta password postgres superuser)" -ForegroundColor Gray
Write-Host ""

$sqlCommands = @"
DO `$`$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'kira') THEN
    CREATE USER kira WITH PASSWORD 'kira_secret';
  END IF;
END `$`$;

SELECT 'CREATE DATABASE kira_db OWNER kira'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'kira_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE kira_db TO kira;
"@

try {
    $sqlCommands | psql -U postgres
    Write-Host ""
    Write-Host "✓ User 'kira' dan database 'kira_db' siap" -ForegroundColor Green
} catch {
    Write-Host ""
    Write-Host "✗ Gagal membuat database: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Coba buat manual dengan psql:" -ForegroundColor Yellow
    Write-Host "  psql -U postgres"
    Write-Host "  CREATE USER kira WITH PASSWORD 'kira_secret';"
    Write-Host "  CREATE DATABASE kira_db OWNER kira;"
    exit 1
}

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
Write-Host "  2. Pastikan DATABASE_URL di .env:" -ForegroundColor White
Write-Host "       DATABASE_URL=postgres://kira:kira_secret@localhost:5432/kira_db?sslmode=disable" -ForegroundColor Gray
Write-Host ""
Write-Host "  3. Jalankan server:" -ForegroundColor White
Write-Host "       go run .\cmd\server\main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "  4. Cek server berjalan:" -ForegroundColor White
Write-Host "       curl http://localhost:8080/health" -ForegroundColor Gray
Write-Host ""
