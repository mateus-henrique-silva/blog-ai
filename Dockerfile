# ─── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a static binary (CGO_ENABLED=0 works because modernc.org/sqlite is pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# ─── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/server .

# Copy templates and static assets
COPY --from=builder /app/web ./web

# Create data and uploads directories
RUN mkdir -p /app/data /app/web/static/uploads

EXPOSE 3000

CMD ["./server"]
