param(
  [string]$Proto = "proto/job/job.proto"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path $Proto)) {
  throw "Proto file not found: $Proto"
}

docker run --rm -v "${PWD}:/work" -w /work golang:1.22-alpine sh -c `
  "set -euo pipefail \
  && apk add --no-cache protobuf protobuf-dev >/dev/null \
  && /usr/local/go/bin/go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1 >/dev/null \
  && /usr/local/go/bin/go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 >/dev/null \
  && protoc -I proto -I /usr/include \
      --plugin=protoc-gen-go=/go/bin/protoc-gen-go \
      --plugin=protoc-gen-go-grpc=/go/bin/protoc-gen-go-grpc \
      --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      $Proto"

Write-Host "Generated Go protos for $Proto"

