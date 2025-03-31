#!/bin/bash

set -eux

# Source .env without automatic exporting
. .env

# Export only the variables needed by the migration command
# Modify this list based on the actual requirements of your migration command
# Common database-related environment variables are included as examples
export DB_HOST
export DB_PORT
export DB_NAME
export DB_USER
export DB_PASSWORD
# Add other required variables from .env as needed

go run ./cmd/hatchet-migrate