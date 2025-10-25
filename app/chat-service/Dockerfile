# Dockerfile for chat-service (multi-stage, reproducible, small final image)
# Build with: docker build -t chat-service:latest .
# Optional: docker build --progress=plain --build-arg BUILD_DIR=./cmd/chat-service .
# Requires BuildKit for cache mounts (recommended).

# ---------- builder ----------
FROM golang:1.24-bullseye AS builder
ARG BUILD_DIR=./cmd/server
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0

WORKDIR /src

# Cache modules separately for faster rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the source
COPY . .
# RUN pwd && ls -la && sleep 100

# Build using Go build
# TODO: use BuildKit to speedup compilation by caching
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/chat-service ${BUILD_DIR}

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian11
ARG USER_ID=10001
ENV PORT=8080

EXPOSE ${PORT}

# Copy statically linked binary
COPY --from=builder /out/chat-service /usr/local/bin/chat-service

# Run as non-root UID
USER ${USER_ID}

ENTRYPOINT ["/usr/local/bin/chat-service"]