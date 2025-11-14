# Dockerfile for message-writer (multi-stage, reproducible, small final image)
# Optional: docker build --progress=plain --build-arg BUILD_DIR=./cmd/message-writer .
# TODO: BuildKit for cache mounts.

# ---------- builder ----------
FROM golang:1.24-bullseye AS builder
ARG BUILD_DIR=./cmd/message-writer
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
    go build -trimpath -ldflags="-s -w" -o /out/message-writer ${BUILD_DIR}

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian11
ARG USER_ID=10001

# Copy statically linked binary
COPY --from=builder /out/message-writer /usr/local/bin/message-writer

# Run as non-root UID
USER ${USER_ID}

ENTRYPOINT ["/usr/local/bin/message-writer"]