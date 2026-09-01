#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/Backend"

if [[ ! -f "$BACKEND/go.mod" ]]; then
  echo "ERROR: Backend/go.mod was not found." >&2
  exit 1
fi

cd "$BACKEND"
export GOTOOLCHAIN=local

echo "==> gofmt check"
UNFORMATTED="$(gofmt -l .)"
if [[ -n "$UNFORMATTED" ]]; then
  echo "The following Go files are not formatted:" >&2
  printf '%s\n' "$UNFORMATTED" >&2
  exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test -count=1 ./...

echo "==> quality gate passed"
