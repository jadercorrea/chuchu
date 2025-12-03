#!/bin/bash

# Collect real GitHub repos to train context detection model

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
DATA_DIR="$SCRIPT_DIR/../data"
TEMP_DIR="/tmp/chu_github_samples"
CHU_BIN="$PROJECT_ROOT/bin/chu"

mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

echo "Collecting GitHub repos for context training..."
echo "This will clone ~60 repos (10 per category)"
echo ""

# Check if chu is built
if [ ! -f "$CHU_BIN" ]; then
    echo "Error: chu binary not found. Run 'make build' first."
    exit 1
fi

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) not found. Install with: brew install gh"
    exit 1
fi

# Output CSV
OUTPUT_CSV="$DATA_DIR/training_data_github.csv"
echo "language_count,primary_ratio,secondary_ratio,has_docs,has_tests,has_scripts,has_infrastructure,has_data,context" > "$OUTPUT_CSV"

collect_category() {
    local category="$1"
    local search_query="$2"
    local samples="$3"
    
    echo "Collecting $category samples..."
    
    # Search GitHub repos
    repos=$(gh search repos "$search_query" --limit "$samples" --json fullName --jq '.[].fullName')
    
    for repo in $repos; do
        echo "  Processing $repo..."
        
        # Clone repo
        repo_dir="$TEMP_DIR/$(echo $repo | tr '/' '_')"
        if [ -d "$repo_dir" ]; then
            rm -rf "$repo_dir"
        fi
        
        gh repo clone "$repo" "$repo_dir" -- --depth 1 --quiet 2>/dev/null || continue
        
        # Run chu detect-language and extract features
        cd "$repo_dir"
        
        # Get language breakdown
        output=$("$CHU_BIN" detect-language 2>/dev/null || echo "")
        
        if [ -z "$output" ]; then
            cd "$TEMP_DIR"
            rm -rf "$repo_dir"
            continue
        fi
        
        # Parse output to extract features
        # This is a simplified parser - in production you'd want structured output
        lang_count=$(echo "$output" | grep -c "%" || echo "0")
        
        # Get primary and secondary ratios
        primary_ratio=$(echo "$output" | grep "%" | head -1 | awk '{print $2}' | tr -d '%' || echo "0")
        secondary_ratio=$(echo "$output" | grep "%" | head -2 | tail -1 | awk '{print $2}' | tr -d '%' || echo "0")
        
        # Convert to 0-1 range
        primary_ratio=$(echo "scale=2; $primary_ratio / 100" | bc -l)
        secondary_ratio=$(echo "scale=2; $secondary_ratio / 100" | bc -l)
        
        # Check for file types
        has_docs=0
        has_tests=0
        has_scripts=0
        has_infra=0
        has_data=0
        
        [ -f "README.md" ] || [ -f "CONTRIBUTING.md" ] && has_docs=1
        find . -name "*test*" -o -name "*spec*" | grep -q . && has_tests=1
        find . -name "*.sh" -o -name "Makefile" | grep -q . && has_scripts=1
        find . -name "Dockerfile" -o -name "*.tf" -o -name "docker-compose.yml" | grep -q . && has_infra=1
        find . -name "*.csv" -o -name "*.json" -o -name "*.yaml" | head -10 | grep -q . && has_data=1
        
        # Write to CSV
        echo "$lang_count,$primary_ratio,$secondary_ratio,$has_docs,$has_tests,$has_scripts,$has_infra,$has_data,$category" >> "$OUTPUT_CSV"
        
        cd "$TEMP_DIR"
        rm -rf "$repo_dir"
    done
}

# Collect pure_code repos (Go projects)
collect_category "pure_code" "language:go stars:>1000 NOT template NOT awesome" 10

# Collect polyglot_balanced repos (multi-language)
collect_category "polyglot_balanced" "language:java language:kotlin stars:>500" 10

# Collect polyglot_scripted repos (main lang + scripts)
collect_category "polyglot_scripted" "language:python language:shell stars:>500" 10

# Collect documentation repos
collect_category "documentation" "topic:documentation stars:>100 NOT awesome" 10

# Collect infrastructure repos
collect_category "infrastructure" "language:hcl OR dockerfile stars:>200" 10

# Collect data_science repos
collect_category "data_science" "topic:machine-learning language:python language:jupyter-notebook stars:>500" 10

echo ""
echo "Collection complete!"
echo "Collected samples: $(wc -l < "$OUTPUT_CSV")"
echo "Output: $OUTPUT_CSV"
echo ""
echo "Next steps:"
echo "  1. Review and clean data: vim $OUTPUT_CSV"
echo "  2. Merge with existing data"
echo "  3. Retrain: python3 $SCRIPT_DIR/train.py"
