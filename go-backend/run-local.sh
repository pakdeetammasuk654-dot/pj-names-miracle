#!/usr/bin/env sh
# Starts the Go API with variables from the local, untracked .env file.
set -eu

if [ ! -f .env ]; then
  echo "Missing .env. Copy .env.example to .env and configure it." >&2
  exit 1
fi

set -a
. ./.env
set +a

exec go run .
