#!/bin/sh

set -e

until tern migrate -c ./db/migrations/tern.conf -m ./db/migrations; do
  echo "Postgres is unavailable - sleeping"
  sleep 2
done

exec go run .
