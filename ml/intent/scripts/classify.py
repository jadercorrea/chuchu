#!/usr/bin/env python3
"""
Classify NL2Bash natural language requests into intent categories.

Based on the NL2Bash corpus:
Lin, Xi Victoria, et al. "NL2Bash: A Corpus and Semantic Parser for Natural Language 
Interface to the Linux Operating System." arXiv preprint arXiv:1802.08979 (2018).
https://victorialin.org/pubs/nl2bash.pdf
"""
import re
import csv
from pathlib import Path
from collections import Counter

SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
NL2BASH_DATA = PROJECT_ROOT / "data" / "nl-request-data.txt"
OUTPUT_CSV = PROJECT_ROOT / "data" / "training_data_expanded.csv"
MERGED_CSV = PROJECT_ROOT / "data" / "training_data_merged.csv"

def categorize_request(text):
    """
    Categorize NL2Bash requests into intent categories.
    
    Intent categories:
    - query: Execute command to see result (read-only operations)
    - editor: Modify files/system (write operations)
    - research: External information (N/A for shell commands)
    - review: Code analysis (N/A for shell commands)
    - router: Greetings/meta (N/A for shell commands)
    """
    text_lower = text.lower()
    
    # Editor: commands that modify files/system
    edit_patterns = [
        'add ', 'remove ', 'delete ', 'rename ', 'move ', 'copy ',
        'create ', 'modify ', 'change ', 'update ', 'replace ',
        'append ', 'prepend ', 'insert ', 'chmod ', 'chown ',
        'mkdir ', 'rmdir ', 'touch ', 'truncate ', 'write ',
        'extract ', 'compress ', 'archive and compress',
        'permission', 'install ', 'set variable', 'adjust '
    ]
    
    # Query: commands that read/display information
    query_patterns = [
        'display ', 'show ', 'list ', 'find ', 'search ',
        'count ', 'calculate ', 'monitor ', 'check ',
        'get ', 'retrieve ', 'output ', 'print ',
        'view ', 'read ', 'look for', 'grep ',
        'diff', 'status', 'log', 'history',
        'page through', 'interactively display', 'collect '
    ]
    
    # Check editor first (more specific)
    for pattern in edit_patterns:
        if pattern in text_lower:
            return 'editor'
    
    # Then check query
    for pattern in query_patterns:
        if pattern in text_lower:
            return 'query'
    
    # Archive operations without modification
    if 'archive ' in text_lower and 'compress' not in text_lower:
        if any(word in text_lower for word in ['to ', 'from ', 'preserve', 'skip']):
            return 'query'
    
    # Default: shell command descriptions are typically query
    return 'query'

def simplify_nl2bash_text(text):
    """
    Convert technical NL2Bash descriptions to natural user messages.
    """
    # Remove line number prefix
    text = re.sub(r'^\d+\|', '', text)
    
    # Remove OS-specific tags
    text = re.sub(r'\((?:GNU|BSD|Linux|Mac OSX)[\s-]specific\)\s*', '', text, flags=re.IGNORECASE)
    
    # Remove quotes
    text = re.sub(r'["\']([^"\']+)["\']', r'\1', text)
    
    # Simplify capitalization
    text = text.replace('Display ', 'display ')
    text = text.replace('Calculate ', 'calculate ')
    text = text.replace('Retrieve ', 'get ')
    text = text.replace('Monitor ', 'monitor ')
    text = text.replace('Find ', 'find ')
    
    return text.strip()

def process_nl2bash():
    """Process NL2Bash corpus and generate categorized training data."""
    print(f"📂 Loading NL2Bash corpus from {NL2BASH_DATA}")
    
    if not NL2BASH_DATA.exists():
        raise FileNotFoundError(f"NL2Bash data not found: {NL2BASH_DATA}")
    
    with open(NL2BASH_DATA, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    print(f"   Loaded {len(lines)} examples from NL2Bash")
    
    # Process and categorize
    processed = []
    for line in lines:
        line = line.strip()
        if not line:
            continue
        
        category = categorize_request(line)
        message = simplify_nl2bash_text(line)
        
        if message and len(message) > 10:  # Filter very short messages
            processed.append({
                'message': message,
                'label': category
            })
    
    print(f"   Processed {len(processed)} examples")
    
    # Count by category
    category_counts = Counter(item['label'] for item in processed)
    print(f"\n   Category distribution:")
    for cat, count in sorted(category_counts.items()):
        print(f"     {cat}: {count}")
    
    # Save expanded dataset
    OUTPUT_CSV.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT_CSV, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=['message', 'label'])
        writer.writeheader()
        writer.writerows(processed)
    
    print(f"\n✅ Expanded dataset saved to {OUTPUT_CSV}")
    return processed

def merge_with_existing(expanded_data):
    """Merge NL2Bash data with existing training data."""
    EXISTING_CSV = PROJECT_ROOT / "data" / "training_data.csv"
    
    print(f"\n📊 Merging datasets...")
    
    # Load existing
    existing = []
    if EXISTING_CSV.exists():
        with open(EXISTING_CSV, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            existing = list(reader)
        print(f"   Existing examples: {len(existing)}")
    
    # Sample from expanded to balance dataset
    # Add 500-1000 shell examples without overwhelming
    from random import sample, seed
    seed(42)  # Reproducible
    
    sample_size = min(1000, len(expanded_data))
    sampled = sample(expanded_data, sample_size)
    
    print(f"   Sampled from NL2Bash: {len(sampled)}")
    
    # Merge
    merged = existing + sampled
    
    # Save merged dataset
    with open(MERGED_CSV, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=['message', 'label'])
        writer.writeheader()
        writer.writerows(merged)
    
    # Stats
    merged_counts = Counter(item['label'] for item in merged)
    print(f"\n   Merged dataset stats:")
    print(f"     Total: {len(merged)} examples")
    for cat, count in sorted(merged_counts.items()):
        print(f"     {cat}: {count}")
    
    print(f"\n✅ Merged dataset saved to {MERGED_CSV}")

def main():
    print("🤖 NL2Bash Intent Classifier")
    print("=" * 60)
    print("Corpus: Lin et al. (2018) - NL2Bash")
    print("Paper: https://victorialin.org/pubs/nl2bash.pdf")
    print("=" * 60)
    
    try:
        expanded = process_nl2bash()
        merge_with_existing(expanded)
        
        print("\n" + "=" * 60)
        print("✨ Classification complete!")
        print("\nNext steps:")
        print("  1. Review training_data_merged.csv")
        print("  2. Run: python scripts/train.py")
        print("  3. Test: chu chat 'run git diff'")
        
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        exit(1)

if __name__ == "__main__":
    main()
