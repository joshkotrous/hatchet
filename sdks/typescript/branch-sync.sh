#!/bin/bash

# 1. Get the current branch name and sanitize it
raw_branch=$(echo "$GITHUB_HEAD_REF" | sed 's/refs\/heads\///')

if [ -z "$raw_branch" ]; then
    raw_branch=$(git rev-parse --abbrev-ref HEAD)
fi

# Sanitize the branch name to prevent command injection
# Only allow alphanumeric characters, hyphens, underscores, periods, and forward slashes (common in git branch names)
if [[ ! "$raw_branch" =~ ^[a-zA-Z0-9_\.\-\/]+$ ]]; then
    echo "Warning: Branch name '$raw_branch' contains invalid characters. Defaulting to 'main' branch."
    current_branch="main"
else
    current_branch="$raw_branch"
fi

# 2. Check a different repo and determine if a branch with the same name exists
git ls-remote --heads https://github.com/hatchet-dev/hatchet.git "$current_branch" | grep -q "refs/heads/$current_branch"
branch_exists=$?

# 3. If it does, update the .gitmodules to set `branch = {the branch name}`
if [ $branch_exists -eq 0 ]; then
    git config -f .gitmodules submodule.hatchet.branch "$current_branch"
    git add .gitmodules
    echo "Updated .gitmodules with branch '$current_branch'"
else
    echo "Branch '$current_branch' does not exist in the remote repository. Pulling main branch instead."
    git config -f .gitmodules submodule.hatchet.branch main
    git add .gitmodules
    echo "Updated .gitmodules with branch main"
fi

# 4. Initialize and update the submodule
git submodule init
git submodule update --remote --merge