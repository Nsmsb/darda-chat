#!/usr/bin/env bash

set -e

# Ensure GOPATH is set
export GOPATH="$(go env GOPATH)"
export PATH="$PATH:$GOPATH/bin"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_ROOT="$PROJECT_ROOT/app/message-reader-service/internal/api"

echo "Generating Go protobuf code..."

PROTO_FILES=$(find "$API_ROOT" -name "*.proto")

for proto in $PROTO_FILES; do
  proto_dir=$(dirname "$proto")
  proto_file=$(basename "$proto")
  gen_dir="$proto_dir/gen"

  echo "Processing: $proto_file in $proto_dir"

  mkdir -p "$gen_dir"

  # Run protoc inside the proto directory
  (
    cd "$proto_dir"

    protoc \
      --go_out=paths=source_relative:"$gen_dir" \
      --go-grpc_out=paths=source_relative:"$gen_dir" \
      -I . \
      "$proto_file"
  )
done

echo "Protobuf generation complete."
