FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o watchme .

# ── Production Stage ──────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata libstdc++ libgcc

WORKDIR /app
COPY --from=builder /app/watchme .

# Default directories (will be overridden by docker-compose volumes)
RUN mkdir -p /data/movies /data/posters /data/config

EXPOSE 8000

CMD ["./watchme"]
