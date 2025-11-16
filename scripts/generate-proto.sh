#!/usr/bin/env bash

set -e

# Ensure GOPATH is set
export GOPATH="$(go env GOPATH)"
export PATH="$PATH:$GOPATH/bin"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# FIXED: No commas, proper bash array
API_ROOT=(
  "$PROJECT_ROOT/app/message-reader-service/internal/api"
  "$PROJECT_ROOT/app/chat-service/internal/api"
)

echo "Generating Go protobuf code..."

# Use proper array expansion
for dir in "${API_ROOT[@]}"; do
  if [ ! -d "$dir" ]; then
    echo "Directory $dir does not exist. Skipping..."
    continue
  fi

  # Find all .proto files in the directory
  PROTO_FILES=$(find "$dir" -name "*.proto")

  for proto in $PROTO_FILES; do
    proto_dir=$(dirname "$proto")
    proto_file=$(basename "$proto")
    gen_dir="$proto_dir/gen"

    echo "Processing: $proto_file in $proto_dir"

    mkdir -p "$gen_dir"

    (
      cd "$proto_dir"

      protoc \
        --go_out=paths=source_relative:"$gen_dir" \
        --go-grpc_out=paths=source_relative:"$gen_dir" \
        -I . \
        "$proto_file"
    )
  done
done

echo "Protobuf generation complete."
