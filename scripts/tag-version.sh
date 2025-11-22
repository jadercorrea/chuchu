#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$REPO_ROOT"

get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

parse_version() {
    local version="$1"
    version="${version#v}"
    
    IFS='.' read -r major minor patch <<< "$version"
    echo "$major $minor $patch"
}

bump_version() {
    local bump_type="$1"
    local current_tag="$(get_latest_tag)"
    
    read -r major minor patch <<< "$(parse_version "$current_tag")"
    
    case "$bump_type" in
        major)
            ((major++))
            minor=0
            patch=0
            ;;
        minor)
            ((minor++))
            patch=0
            ;;
        patch)
            ((patch++))
            ;;
        *)
            echo "Error: Invalid bump type. Use 'major', 'minor', or 'patch'."
            exit 1
            ;;
    esac
    
    echo "v${major}.${minor}.${patch}"
}

show_usage() {
    cat <<EOF
Usage: $0 <major|minor|patch> [options]

Bump version and create a git tag with semantic versioning.

Arguments:
  major       Bump major version (X.0.0)
  minor       Bump minor version (x.X.0)
  patch       Bump patch version (x.x.X)

Options:
  -m MESSAGE  Custom tag message
  -p          Push tag to remote after creation
  -h          Show this help message

Examples:
  $0 patch              # Create v0.0.1, v0.0.2, etc.
  $0 minor              # Create v0.1.0, v0.2.0, etc.
  $0 major              # Create v1.0.0, v2.0.0, etc.
  $0 patch -p           # Create and push
  $0 minor -m "New feature release"

Current version: $(get_latest_tag)
EOF
}

main() {
    if [ $# -eq 0 ]; then
        show_usage
        exit 1
    fi
    
    local bump_type="$1"
    shift
    
    local message=""
    local push_tag=false
    
    while getopts "m:ph" opt; do
        case "$opt" in
            m)
                message="$OPTARG"
                ;;
            p)
                push_tag=true
                ;;
            h)
                show_usage
                exit 0
                ;;
            *)
                show_usage
                exit 1
                ;;
        esac
    done
    
    if ! git diff-index --quiet HEAD --; then
        echo "Error: You have uncommitted changes. Please commit or stash them first."
        exit 1
    fi
    
    local current_tag="$(get_latest_tag)"
    local new_tag="$(bump_version "$bump_type")"
    
    echo "Current version: $current_tag"
    echo "New version: $new_tag"
    echo
    
    if [ -z "$message" ]; then
        message="Release $new_tag"
    fi
    
    read -p "Create tag $new_tag? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
    
    git tag -a "$new_tag" -m "$message"
    echo "Created tag: $new_tag"
    
    if [ "$push_tag" = true ]; then
        echo "Pushing tag to origin..."
        git push origin "$new_tag"
        echo "Tag pushed successfully!"
    else
        echo
        echo "Tag created locally. To push to remote, run:"
        echo "  git push origin $new_tag"
    fi
    
    echo
    echo "Release $new_tag created successfully!"
    echo
    echo "This will trigger:"
    echo "  - GitHub Actions CD workflow"
    echo "  - Multi-platform binary builds"
    echo "  - GitHub Release creation"
}

main "$@"
