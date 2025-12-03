#!/bin/bash

# Collect from curated list of known repos

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
DATA_DIR="$SCRIPT_DIR/../data"
TEMP_DIR="/tmp/chu_github_samples"
CHU_BIN="$PROJECT_ROOT/bin/chu"

mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

OUTPUT_CSV="$DATA_DIR/training_data_github.csv"
echo "language_count,primary_ratio,secondary_ratio,has_docs,has_tests,has_scripts,has_infrastructure,has_data,context" > "$OUTPUT_CSV"

# Curated repos by category
declare -A REPOS

# Pure code (Go projects)
REPOS[pure_code]="golang/go kubernetes/kubernetes prometheus/prometheus etcd-io/etcd hashicorp/terraform"

# Polyglot balanced (Java+Kotlin, Python+JS, etc)
REPOS[polyglot_balanced]="apache/kafka spring-projects/spring-boot elastic/elasticsearch"

# Polyglot scripted (main lang + scripts)
REPOS[polyglot_scripted]="ansible/ansible django/django rails/rails"

# Documentation sites
REPOS[documentation]="rust-lang/rust-by-example vuejs/docs python/cpython"

# Infrastructure
REPOS[infrastructure]="hashicorp/terraform-provider-aws kelseyhightower/kubernetes-the-hard-way"

# Data science
REPOS[data_science]="tensorflow/tensorflow scikit-learn/scikit-learn jupyter/notebook"

for context in "${!REPOS[@]}"; do
    echo "Collecting $context samples..."
    
    for repo in ${REPOS[$context]}; do
        echo "  Processing $repo..."
        
        repo_dir="$TEMP_DIR/$(echo $repo | tr '/' '_')"
        
        if [ -d "$repo_dir" ]; then
            rm -rf "$repo_dir"
        fi
        
        gh repo clone "$repo" "$repo_dir" -- --depth 1 --quiet 2>/dev/null || {
            echo "    Failed to clone, skipping..."
            continue
        }
        
        cd "$repo_dir"
        
        output=$("$CHU_BIN" detect-language 2>/dev/null || echo "")
        
        if [ -z "$output" ]; then
            echo "    No output from chu, skipping..."
            cd "$TEMP_DIR"
            continue
        fi
        
        lang_count=$(echo "$output" | grep -c "%" || echo "0")
        primary_ratio=$(echo "$output" | grep "%" | head -1 | awk '{print $2}' | tr -d '%' || echo "0")
        secondary_ratio=$(echo "$output" | grep "%" | head -2 | tail -1 | awk '{print $2}' | tr -d '%' || echo "0")
        
        primary_ratio=$(echo "scale=2; $primary_ratio / 100" | bc -l 2>/dev/null || echo "0")
        secondary_ratio=$(echo "scale=2; $secondary_ratio / 100" | bc -l 2>/dev/null || echo "0")
        
        has_docs=0
        has_tests=0
        has_scripts=0
        has_infra=0
        has_data=0
        
        [ -f "README.md" ] && has_docs=1
        find . -maxdepth 3 -name "*test*" 2>/dev/null | grep -q . && has_tests=1
        find . -maxdepth 2 -name "*.sh" -o -name "Makefile" 2>/dev/null | grep -q . && has_scripts=1
        find . -maxdepth 2 -name "Dockerfile" -o -name "*.tf" 2>/dev/null | grep -q . && has_infra=1
        find . -maxdepth 2 -name "*.csv" 2>/dev/null | head -5 | grep -q . && has_data=1
        
        echo "$lang_count,$primary_ratio,$secondary_ratio,$has_docs,$has_tests,$has_scripts,$has_infra,$has_data,$context" >> "$OUTPUT_CSV"
        echo "    Added sample"
        
        cd "$TEMP_DIR"
        rm -rf "$repo_dir"
    done
done

echo ""
echo "Collection complete!"
samples=$(tail -n +2 "$OUTPUT_CSV" | wc -l | tr -d ' ')
echo "Collected samples: $samples"
echo "Output: $OUTPUT_CSV"
