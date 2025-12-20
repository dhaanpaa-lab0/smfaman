#!/bin/bash

# release.sh
# Complete release automation script for smfaman
# This script will:
#   1. Create a new version tag (yyyy.mm.xxxxx format)
#   2. Run goreleaser to build distribution kits
#   3. Create GitHub release with distribution kits attached

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}✗ Error: $1${NC}" >&2
    exit 1
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# ============================================================================
# 1. Pre-flight Checks
# ============================================================================

echo ""
info "Running pre-flight checks..."
echo ""

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    error "goreleaser is not installed. Install from: https://goreleaser.com/install/"
fi
success "goreleaser is installed"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    error "gh CLI is not installed. Install from: https://cli.github.com/"
fi
success "gh CLI is installed"

# Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    error "gh CLI is not authenticated. Run: gh auth login"
fi
success "gh CLI is authenticated"

# Check if we're in a git repository
if ! git rev-parse --git-dir &> /dev/null; then
    error "Not in a git repository"
fi
success "In a git repository"

# Check if working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    warning "Working directory is not clean"
    git status --short
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Aborted due to dirty working directory"
    fi
else
    success "Working directory is clean"
fi

# Check if we're on main/master branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" ]] && [[ "$CURRENT_BRANCH" != "master" ]]; then
    warning "Not on main/master branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Aborted - not on main/master branch"
    fi
else
    success "On $CURRENT_BRANCH branch"
fi

# Check if GITHUB_TOKEN is set (required by goreleaser)
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    info "GITHUB_TOKEN not set, attempting to use gh CLI token..."
    export GITHUB_TOKEN=$(gh auth token)
    if [[ -z "$GITHUB_TOKEN" ]]; then
        error "Could not obtain GitHub token"
    fi
fi
success "GitHub token is available"

# Get GitHub user/org name
GITHUB_REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
if [[ -z "$GITHUB_REPO" ]]; then
    error "Could not determine GitHub repository"
fi
export GITHUB_USER=$(echo "$GITHUB_REPO" | cut -d'/' -f1)
success "GitHub repository: $GITHUB_REPO"

echo ""

# ============================================================================
# 2. Create Version Tag
# ============================================================================

info "Creating version tag..."
echo ""

# Get current year and month
YEAR=$(date +%Y)
MONTH=$(date +%m)

# Pattern for current month's tags
TAG_PREFIX="${YEAR}.${MONTH}."

# Fetch all tags from remote to ensure we have the latest
info "Fetching tags from remote..."
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
info "New tag will be: ${NEW_TAG}"
echo ""

# Allow custom tag if desired
read -p "Use this tag or enter custom tag (press Enter to use $NEW_TAG): " CUSTOM_TAG
if [[ -n "$CUSTOM_TAG" ]]; then
    NEW_TAG="$CUSTOM_TAG"
    info "Using custom tag: $NEW_TAG"
fi
echo ""

# Confirm before proceeding
read -p "Create tag $NEW_TAG and start release process? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    error "Aborted by user"
fi

# Create the tag
info "Creating tag ${NEW_TAG}..."
git tag -a "$NEW_TAG" -m "Release ${NEW_TAG}"
success "Tag created locally"

# Push the tag to origin
info "Pushing tag to origin..."
git push origin "$NEW_TAG"
success "Tag pushed to remote"

echo ""

# ============================================================================
# 3. Run GoReleaser
# ============================================================================

info "Running goreleaser..."
echo ""

# Navigate to project root
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Run goreleaser
# This will:
#   - Build binaries for all platforms
#   - Create distribution archives
#   - Generate checksums
#   - Create GitHub release
#   - Upload all artifacts to GitHub release
info "Building distribution kits and creating GitHub release..."
echo ""

if goreleaser release --clean; then
    success "goreleaser completed successfully"
else
    error "goreleaser failed. You may need to delete the tag: git tag -d $NEW_TAG && git push --delete origin $NEW_TAG"
fi

echo ""

# ============================================================================
# 4. Verify Release
# ============================================================================

info "Verifying release..."
echo ""

# Wait a moment for GitHub to process
sleep 2

# Check if release exists
if gh release view "$NEW_TAG" &> /dev/null; then
    success "GitHub release created successfully"

    # Show release info
    echo ""
    info "Release details:"
    gh release view "$NEW_TAG"

    echo ""
    success "Release $NEW_TAG completed successfully!"
    echo ""
    info "View release at: https://github.com/$GITHUB_REPO/releases/tag/$NEW_TAG"
    echo ""
else
    warning "Could not verify GitHub release"
    echo ""
    info "Please check manually at: https://github.com/$GITHUB_REPO/releases"
fi

# ============================================================================
# 5. Post-Release Summary
# ============================================================================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
success "Release Process Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "What was done:"
echo "  1. ✓ Created tag: $NEW_TAG"
echo "  2. ✓ Built distribution kits for all platforms"
echo "  3. ✓ Created GitHub release with artifacts"
echo ""
echo "Next steps:"
echo "  • Announce the release"
echo "  • Update documentation if needed"
echo "  • Monitor for issues"
echo ""
