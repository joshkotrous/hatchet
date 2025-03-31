#!/bin/bash
# This scripts generates a local API token.

set -eux

# Source the .env file
. .env

# SECURITY IMPROVEMENT:
# The original script exported all variables from .env to the environment,
# which could expose sensitive information. We now source the file without
# automatic exports.
#
# If the Go command requires environment variables from .env,
# uncomment and modify one of these approaches:
#
# 1. Export specific variables that are needed:
#    export VARIABLE1
#    export VARIABLE2
#
# 2. If you're sure all variables in .env are necessary and safe to export:
#    (This is less secure but maintains original functionality)
#    set -a
#    . .env
#    set +a

go run ./cmd/hatchet-admin token create --name "local" --tenant-id 707d0855-80ab-4e1f-a156-f1c4546cbf52