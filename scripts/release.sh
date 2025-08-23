#!/bin/bash

# GoVTE Release Script
# Usage: ./scripts/release.sh v0.2.0

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Error: Version not specified"
    echo "Usage: $0 v0.2.0"
    exit 1
fi

# Remove 'v' prefix if present for version comparisons
VERSION_NUM=${VERSION#v}

echo "🚀 Preparing to release GoVTE ${VERSION}"
echo "================================"

# Step 1: Run tests
echo "📋 Running tests..."
if ! go test ./...; then
    echo "❌ Tests failed. Please fix before releasing."
    exit 1
fi
echo "✅ Tests passed"

# Step 2: Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  You have uncommitted changes:"
    git status --short
    echo ""
    read -p "Do you want to commit these changes? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git add .
        git commit -m "Prepare for ${VERSION} release"
    else
        echo "❌ Please commit or stash your changes before releasing."
        exit 1
    fi
fi

# Step 3: Create tag
echo "🏷️  Creating tag ${VERSION}..."
git tag -a "${VERSION}" -m "Release ${VERSION}

$(grep -A 50 "\[${VERSION_NUM}\]" CHANGELOG.md | sed '/^## \[/,$d' | sed '/^---/,$d')"

# Step 4: Push to remote
echo "📤 Pushing to remote..."
git push origin main
git push origin "${VERSION}"

echo ""
echo "✅ Release ${VERSION} created successfully!"
echo ""
echo "Next steps:"
echo "1. Check GitHub Actions: https://github.com/cliofy/govte/actions"
echo "2. View release: https://github.com/cliofy/govte/releases/tag/${VERSION}"
echo "3. Verify on pkg.go.dev: https://pkg.go.dev/github.com/cliofy/govte@${VERSION}"
echo ""
echo "To announce the release, you can use the template in RELEASE_CHECKLIST.md"