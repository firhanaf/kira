.PHONY: dev build test migrate migrate-down migrate-create lint docker-build clean

# ─── Variables ─────────────────────────────────────────
BINARY_NAME=kira-server
BUILD_DIR=./bin
MAIN=./cmd/server/main.go
DB_URL?=$(shell grep DATABASE_URL .env | cut -d '=' -f2-)

# ─── Development ───────────────────────────────────────
dev:
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air -c .air.toml

run:
	go run $(MAIN)

# ─── Build ─────────────────────────────────────────────
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN)
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)"

# ─── Database Migrations ───────────────────────────────
migrate:
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -database "$(DB_URL)" -path ./internal/db/migrations up
	@echo "✓ Migrations applied"

migrate-down:
	migrate -database "$(DB_URL)" -path ./internal/db/migrations down 1

migrate-status:
	migrate -database "$(DB_URL)" -path ./internal/db/migrations version

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir ./internal/db/migrations -seq $$name
	@echo "✓ Migration files created"

# ─── Testing ───────────────────────────────────────────
test:
	go test -race -cover ./...

test-verbose:
	go test -v -race -cover ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# ─── Code Quality ──────────────────────────────────────
lint:
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

vet:
	go vet ./...

# ─── Docker ────────────────────────────────────────────
docker-build:
	docker build -t kira-server:latest .

docker-up:
	docker-compose up -d
	@echo "✓ Postgres running on :5432"

docker-down:
	docker-compose down

# ─── Cleanup ───────────────────────────────────────────
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

# ─── Setup (first time) ────────────────────────────────
setup:
	cp -n .env.example .env || true
	docker-compose up -d postgres
	@echo "Waiting for postgres..."
	@sleep 3
	$(MAKE) migrate
	@echo ""
	@echo "✓ Kira backend ready!"
	@echo "  Run: make dev"
