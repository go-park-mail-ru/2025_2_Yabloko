#!/bin/sh
set -e

until tern migrate -c ./db/migrations/tern.conf -m ./db/migrations; do
  echo "Postgres is unavailable - sleeping"
  sleep 2
done

go test ./... -coverpkg=./... -v
go test ./... -coverpkg=./... -coverprofile=coverage.out
echo "=== TOTAL COVERAGE ==="
go tool cover -func=coverage.out | grep total
exec go run .
