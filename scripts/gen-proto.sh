#!/usr/bin/env bash
# Regenerates Go gRPC code from src/proto/*.proto.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc (go install ...@latest).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$ROOT_DIR/src"

# module=ticketflow makes protoc-gen-go place each file's output under the
# directory matching its go_package import path (e.g. proto/identitypb/),
# relative to --go_out — necessary because identity/event/booking each
# declare a distinct Go package and cannot share one directory.
cd "$SRC_DIR"
for proto_file in proto/*.proto; do
  echo "Generating $proto_file"
  protoc -I proto \
    --go_out=. --go_opt=module=ticketflow \
    --go-grpc_out=. --go-grpc_opt=module=ticketflow \
    "$proto_file"
done
