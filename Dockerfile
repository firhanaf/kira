# ── Stage 1: Build ────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../../AppData/Local/Temp/Rar$DRa16280.39516/kira-server .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/kira-server \
    ./cmd/server/main.go

# ── Stage 2: Runtime ──────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/kira-server .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
COPY --from=builder /app/templates ./templates

RUN mkdir -p ./storage/pdfs

EXPOSE 8080

ENTRYPOINT ["./kira-server"]
