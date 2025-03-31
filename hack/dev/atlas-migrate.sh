#!/bin/bash

MIGRATION_NAME=$1

# check if the first argument is empty
if [ -z "$MIGRATION_NAME" ]; then
  MIGRATION_NAME="temp"
else
  # Validate that the migration name only contains allowed characters
  if ! [[ "$MIGRATION_NAME" =~ ^[a-zA-Z0-9_.-]+$ ]]; then
    echo "Error: Migration name must only contain alphanumeric characters, underscores, dots, and hyphens."
    exit 1
  fi
fi

atlas migrate hash --dir "file://sql/atlas"

atlas migrate diff "$MIGRATION_NAME" \
  --dir "file://sql/atlas" \
  --to "file://sql/schema/v0.sql" \
  --dev-url "docker://postgres/15/dev?search_path=public"