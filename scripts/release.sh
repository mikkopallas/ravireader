#!/usr/bin/env bash
set -e

# Determine the next semantic version
echo "Determining next version..."
NEXT_VERSION=$(go run github.com/caarlos0/svu/v3@latest next)
CURRENT_VERSION=$(go run github.com/caarlos0/svu/v3@latest current)

if [ "$NEXT_VERSION" = "$CURRENT_VERSION" ]; then
    echo "No significant changes found in commits to trigger a new release. Current version is $CURRENT_VERSION."
    exit 0
fi

echo "Bumping version: $CURRENT_VERSION -> $NEXT_VERSION"

# Generate Changelog
echo "Generating CHANGELOG.md..."
go run github.com/git-chglog/git-chglog/cmd/git-chglog@latest -o CHANGELOG.md --next-tag "$NEXT_VERSION"

# Provide instructions to the user
echo "================================================================"
echo "✅ CHANGELOG.md has been generated for release $NEXT_VERSION!"
echo "Please review the changes in CHANGELOG.md."
echo ""
echo "When you are ready to finalize the release, run the following:"
echo "  git commit -am \"chore(release): $NEXT_VERSION\""
echo "  git tag $NEXT_VERSION"
echo "  git push --follow-tags"
echo "================================================================"
