#!/bin/bash

# Enable error checking and variable checking, but not command tracing
set -eu

# Warning about .env file security
echo "WARNING: This script loads environment variables from the .env file."
echo "Make sure the .env file is properly secured with restrictive file permissions."
echo "Sensitive variables might be exposed if the file is compromised."

# Load environment variables from .env file
if [ -f .env ]; then
    # Only export variables during the sourcing of .env
    set -a
    . .env
    set +a
else
    echo "Warning: .env file not found."
fi

# Start the development server
npx --yes nodemon --signal SIGINT --config nodemon.engine.json --exec go run ./cmd/hatchet-engine --no-graceful-shutdown