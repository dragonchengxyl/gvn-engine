# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Cache dependency downloads before copying source
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Run pure-Go tests (packages that have no X11 / GPU dependency)
RUN go test \
    ./internal/compiler/ \
    ./internal/history/ \
    ./internal/ecs/ \
    ./internal/engine/ \
    -v -count=1

# Cross-compile Windows .exe (CGO disabled — Ebitengine v2 supports pure-Go)
RUN GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o gvn-engine.exe ./cmd/game

# Cross-compile Linux binary (headless CI runner)
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o gvn-engine-linux ./cmd/game

# ── Runtime stage (minimal, for headless CI validation) ─────────────────────
FROM debian:bookworm-slim AS runtime

WORKDIR /app
COPY --from=builder /app/gvn-engine-linux ./gvn-engine

# Default: run headless validation against the embedded demo script
ENTRYPOINT ["./gvn-engine", "-headless", "-script", "scripts/demo.json"]
