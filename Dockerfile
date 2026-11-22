# Multi-stage Cloud-Native Dockerfile for Edeth
# Stage 1: Build binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and build essentials
RUN apk add --no-cache git gcc musl-dev

# Cache dependencies
COPY go.mod ./
# RUN go mod download # uncomment if external dependencies exist

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o edeth .

# Stage 2: Final lightweight runtime container with FFmpeg
FROM alpine:3.19

# Install runtime dependencies: FFmpeg, fontconfig, fonts for subtitles, and ca-certificates
RUN apk add --no-cache ffmpeg ca-certificates tzdata fontconfig ttf-dejavu \
    && addgroup -S edeth && adduser -S edeth -G edeth \
    && mkdir -p /app /workspace \
    && chown -R edeth:edeth /app /workspace \
    && chmod 1777 /tmp

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/edeth /app/edeth

# Set permissions
RUN chown -R edeth:edeth /app

USER edeth

# Standard container configuration
ENV PORT=8420
ENV HEADLESS=1
ENV DOCKER=1
EXPOSE 8420

ENTRYPOINT ["/app/edeth"]
