#!/bin/bash

set -eux

caddy start

# Security checks for .env file
if [ ! -f .env ]; then
  echo "Error: .env file does not exist. Please create it with appropriate configuration."
  exit 1
fi

# Check for group/world read permissions
perm_string=$(ls -l .env | awk '{print $1}')
if [[ "${perm_string:4:1}" == "r" || "${perm_string:7:1}" == "r" ]]; then
  echo "Warning: .env file has insecure permissions (readable by group or others)."
  echo "Current permissions: $(ls -l .env)"
  echo "Consider running: chmod 600 .env"
fi

# Source .env file
set -a
. .env
set +a

npx --yes nodemon --signal SIGINT --config nodemon.api.json --exec go run ./cmd/hatchet-lite