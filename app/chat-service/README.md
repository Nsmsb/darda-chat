# Chat Service

A WebSocket-based chat service for exchanging messages between clients.

## Overview
This service exposes a WebSocket endpoint to send and receive JSON messages between clients or broadcast messages.
Core logic lives in the `handler` package and message model in `model`.

## Repository layout (relevant paths)
- Configuration: `app/chat-service/internal/config/config.go` (type `config.Config`, function `config.Load`)
- Server entrypoint: `app/chat-service/cmd/server/main.go`
- WebSocket handler: `app/chat-service/internal/handler/message_handler.go`
- Message model: `app/chat-service/internal/model/message.go`

## Prerequisites
- Go (compatible version installed)
- Linux environment (development instructions below assume Linux)
- Redis server running

## Configs
Configuration is loaded from environment variables at startup. See app/chat-service/internal/config/config.go (type `config.Config`, function `config.Load`) for the concrete fields and parsing rules.

Used environment variables
- PORT (string) — server port the HTTP/WebSocket listener binds to. Default: `8080`.
- REDIS_ADDR (string) — Redis address. Default: `localhost:6379`.
- REDIS_PASS (string) — Redis password (if any). Default: empty.
- REDIS_DB (int) — Redis database index. Default: `0`.

## Run (development)
From the chat-service directory:
```bash
cd app/chat-service
# Run Redis if not running
docker run -d --name redis -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
# Run Chat-service
PORT=8080 REDIS_ADDR=localhost:6379 go run ./cmd/server/main.go
```

## WebSocket API
- Endpoint: GET /api/v1/ws?id={userId}
- Message JSON format (see `model.Message`):
  - sender: string
  - destination: string (user id or `"broadcast"`)
  - content: string