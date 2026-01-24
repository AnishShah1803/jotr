#!/bin/bash

set -e

TYPE=${1:-patch}
DRY_RUN=false
BACKUP_DIR="/tmp/jotr-release-backup-$(date +%s)"

while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            TYPE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 --type (patch|minor|major) [--dry-run]"
            exit 1
            ;;
    esac
done

create_backup() {
    echo "Creating backup in $BACKUP_DIR..."
    mkdir -p "$BACKUP_DIR"

    git rev-parse HEAD > "$BACKUP_DIR/commit_hash"

    if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
        git stash push -m "Release script backup" --include-untracked
        echo "STASHED=true" > "$BACKUP_DIR/state"
    else
        echo "STASHED=false" > "$BACKUP_DIR/state"
    fi

    echo "Backup created"
}

restore_backup() {
    echo "Restoring from backup..."

    if [[ ! -d "$BACKUP_DIR" ]]; then
        echo "❌ No backup directory found"
        return 1
    fi

    if [[ $(cat "$BACKUP_DIR/state") == "true" ]]; then
        git stash pop || true
    fi

    echo "Backup restored"
}

LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [[ -n "$LATEST_TAG" ]]; then
    CURRENT_VERSION="$LATEST_TAG"
else
    CURRENT_YEAR=$(date +%Y)
    CURRENT_MONTH=$(date +%-m)
    YEAR_IN_DEV=$(expr $CURRENT_YEAR - 2025)
    CURRENT_VERSION="v${YEAR_IN_DEV}.${CURRENT_MONTH}.0"
fi

if [[ -z "$CURRENT_VERSION" ]]; then
    echo "❌ Could not determine current version"
    exit 1
fi

echo "📋 Current version: $CURRENT_VERSION"
echo "📈 Release type: $TYPE"

IFS='.' read -ra VERSION_PARTS <<< "${CURRENT_VERSION#v}"
YEAR=${VERSION_PARTS[0]:-1}
MONTH=${VERSION_PARTS[1]:-1}
PATCH=${VERSION_PARTS[2]:-0}

case $TYPE in
    "patch")
        NEW_VERSION="v${YEAR}.${MONTH}.$((PATCH + 1))"
        ;;
    "minor")  
        NEW_VERSION="v${YEAR}.$((MONTH + 1)).0"
        ;;
    "major")
        NEW_VERSION="v$((YEAR + 1)).1.0"
        ;;
    *)
        echo "❌ Invalid release type. Use: patch|minor|major"
        exit 1
        ;;
esac

echo "🆕 New version: $NEW_VERSION"

if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY RUN: Would release $CURRENT_VERSION -> $NEW_VERSION"
    exit 0
fi

echo ""
echo "Release summary:"
echo "  Current: $CURRENT_VERSION"
echo "  New:     $NEW_VERSION"
echo "  Type:    $TYPE"
echo ""

read -p "❓ Continue with release? (y/N): " -n 1 -r
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Release cancelled"
    exit 0
fi

echo "Starting release process..."

trap 'echo "❌ Release failed! Attempting to restore backup..."; restore_backup; exit 1' ERR

create_backup

echo "🧪 Running safety checks..."

if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
    echo "❌ Working directory not clean. Commit or stash changes first."
    git status --short
    exit 1
fi

if ! git rev-parse --verify HEAD >/dev/null 2>&1; then
    echo "❌ Not on a valid git commit"
    exit 1
fi

echo "Safety checks passed"

echo ""
if [[ "$LATEST_TAG" != "" ]]; then
    echo "📋 Changes since $CURRENT_VERSION:"
    git log --oneline --pretty=format:"%h %s" "${CURRENT_VERSION}..HEAD" || echo "No changes found"
    CHANGELOG=$(git log --oneline --pretty=format:"%h %s" "${CURRENT_VERSION}..HEAD" | head -20)
else
    echo "📋 This is the first release!"
    CHANGELOG="Initial release - Genesis of JOTR"
fi

echo ""
read -p "📝 Edit release notes above (press Enter to continue): "

echo "💾 Creating release commit (no version file changes - version set at build time)..."
git commit --allow-empty -m "Release $NEW_VERSION

Changes:
$CHANGELOG

Automated release script via ./scripts/release.sh --type $TYPE"

echo "Pushing commit and tag..."
git push origin HEAD
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
git push origin "$NEW_VERSION"

echo "Release complete!"
echo ""
echo "Next steps:"
echo "  1. GitHub Actions will build and create release automatically"
echo "  2. Check: https://github.com/AnishShah1803/jotr/actions"
echo "  3. Review draft release before publishing"
echo ""
echo "Cleanup: Removing backup directory $BACKUP_DIR"
rm -rf "$BACKUP_DIR"
echo ""
echo "Happy releasing!"