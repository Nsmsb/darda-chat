# Dockerfile for message-reader (multi-stage, reproducible, small final image)
# Optional: docker build --progress=plain --build-arg BUILD_DIR=./cmd/message-reader .
# TODO: BuildKit for cache mounts.

# ---------- builder ----------
FROM golang:1.24-bullseye AS builder
ARG BUILD_DIR=./cmd/message-reader
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0

WORKDIR /src

# Cache modules separately for faster rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the source
COPY . .

# Build using Go build
# TODO: use BuildKit to speedup compilation by caching
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/message-reader ${BUILD_DIR}

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian11
ARG USER_ID=10001
# GRPC server port
ENV PORT=50051

EXPOSE ${PORT}

# Copy statically linked binary
COPY --from=builder /out/message-reader /usr/local/bin/message-reader

# Run as non-root UID
USER ${USER_ID}

ENTRYPOINT ["/usr/local/bin/message-reader"]