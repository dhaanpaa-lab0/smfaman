#!/bin/bash

# tag-version.sh
# Automatically creates and pushes version tags in format: yyyy.mm.xxxxx
# Where yyyy = year, mm = month, xxxxx = incremental number

set -e  # Exit on error

# Get current year and month
YEAR=$(date +%Y)
MONTH=$(date +%m)

# Pattern for current month's tags
TAG_PREFIX="${YEAR}.${MONTH}."

# Fetch all tags from remote to ensure we have the latest
echo "Fetching tags from remote..."
git fetch --tags

# Find all tags matching current month's pattern
EXISTING_TAGS=$(git tag -l "${TAG_PREFIX}*" | sort -V)

# Determine next version number
if [ -z "$EXISTING_TAGS" ]; then
    # No tags for this month yet, start at 1
    NEXT_NUM=1
else
    # Get the highest number from existing tags
    LAST_TAG=$(echo "$EXISTING_TAGS" | tail -n 1)
    LAST_NUM=$(echo "$LAST_TAG" | sed "s/${TAG_PREFIX}//")
    NEXT_NUM=$((LAST_NUM + 1))
fi

# Format the new tag (pad to 5 digits)
NEW_TAG=$(printf "%s.%s.%05d" "$YEAR" "$MONTH" "$NEXT_NUM")

echo "Current tags for ${YEAR}.${MONTH}:"
if [ -z "$EXISTING_TAGS" ]; then
    echo "  (none)"
else
    echo "$EXISTING_TAGS" | sed 's/^/  /'
fi
echo ""
echo "New tag will be: ${NEW_TAG}"
echo ""

# Confirm before proceeding
read -p "Create and push this tag? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Create the tag
echo "Creating tag ${NEW_TAG}..."
git tag -a "$NEW_TAG" -m "Release ${NEW_TAG}"

# Push the tag to origin
echo "Pushing tag to origin..."
git push origin "$NEW_TAG"

echo ""
echo "✓ Successfully created and pushed tag: ${NEW_TAG}"
