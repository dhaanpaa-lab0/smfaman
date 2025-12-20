#!/bin/bash

# create-release.sh
# Automatically creates a GitHub release for the latest tag using gh CLI
# Requires: gh CLI installed and authenticated (gh auth login)

set -e  # Exit on error

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: gh CLI is not installed."
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    echo "Error: gh CLI is not authenticated."
    echo "Run: gh auth login"
    exit 1
fi

# Fetch all tags from remote to ensure we have the latest
echo "Fetching tags from remote..."
git fetch --tags

# Get the latest tag
LATEST_TAG=$(git tag --sort=-version:refname | head -n 1)

if [ -z "$LATEST_TAG" ]; then
    echo "Error: No tags found in repository."
    exit 1
fi

echo "Latest tag: ${LATEST_TAG}"
echo ""

# Check if release already exists for this tag
if gh release view "$LATEST_TAG" &> /dev/null; then
    echo "Release for tag ${LATEST_TAG} already exists."
    read -p "Do you want to delete and recreate it? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Deleting existing release..."
        gh release delete "$LATEST_TAG" -y
    else
        echo "Aborted."
        exit 1
    fi
fi

# Confirm before proceeding
echo "This will create a GitHub release for tag: ${LATEST_TAG}"
read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create the release
# --generate-notes automatically creates release notes from commits
# --latest marks this as the latest release
echo "Creating GitHub release for ${LATEST_TAG}..."
gh release create "$LATEST_TAG" \
    --title "Release ${LATEST_TAG}" \
    --generate-notes \
    --latest

echo ""
echo "✓ Successfully created GitHub release: ${LATEST_TAG}"
echo ""
echo "View your release at:"
gh release view "$LATEST_TAG" --web --web-only=false 2>&1 | grep -o 'https://.*' || gh browse --repo $(gh repo view --json nameWithOwner -q .nameWithOwner) --no-browser
