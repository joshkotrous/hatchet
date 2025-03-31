#!/bin/bash

set -eux

caddy start

# Temporarily disable command echo to avoid leaking sensitive information
set +x

# Security reminder: Ensure .env has restrictive permissions (chmod 600)
# and is never committed to version control
if [ -f .env ]; then
    set -a
    . .env
    set +a
else
    echo "WARNING: .env file not found. Environment variables may be missing."
fi

# Re-enable command echo
set -x

npx --yes nodemon --signal SIGINT --config nodemon.api.json --exec go run ./cmd/hatchet-api