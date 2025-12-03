#!/usr/bin/env python3
"""
Collect GitHub repos data and retrain context detection model
"""

import json
import os
import subprocess
import tempfile
from pathlib import Path

import pandas as pd

# Curated repos by category
REPOS = {
    "pure_code": [
        "golang/go",
        "kubernetes/kubernetes", 
        "prometheus/prometheus",
        "etcd-io/etcd",
    ],
    "polyglot_balanced": [
        "apache/kafka",
        "elastic/elasticsearch",
    ],
    "polyglot_scripted": [
        "django/django",
        "rails/rails",
    ],
    "documentation": [
        "vuejs/docs",
    ],
    "infrastructure": [
        "hashicorp/terraform-provider-aws",
    ],
    "data_science": [
        "scikit-learn/scikit-learn",
        "jupyter/notebook",
    ],
}

def run_chu_detect(repo_dir):
    """Run chu detect-language and parse output"""
    chu_bin = Path(__file__).parent.parent.parent.parent / "bin" / "chu"
    
    result = subprocess.run(
        [str(chu_bin), "detect-language"],
        cwd=repo_dir,
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0 or not result.stdout:
        return None
    
    # Parse output
    lines = result.stdout.strip().split("\n")
    lang_lines = [l for l in lines if "%" in l]
    
    if not lang_lines:
        return None
    
    # Extract features
    lang_count = len(lang_lines)
    
    # Parse percentages more carefully
    def parse_pct(line):
        parts = line.split()
        for part in parts:
            if "%" in part:
                try:
                    return float(part.rstrip("%")) / 100.0
                except ValueError:
                    continue
        return 0.0
    
    primary_pct = parse_pct(lang_lines[0])
    secondary_pct = 0.0
    if len(lang_lines) > 1:
        secondary_pct = parse_pct(lang_lines[1])
    
    # Check for file types
    repo_path = Path(repo_dir)
    has_docs = int((repo_path / "README.md").exists())
    
    has_tests = 0
    for p in repo_path.rglob("*test*"):
        if p.is_file():
            has_tests = 1
            break
    
    has_scripts = 0
    for ext in ["*.sh", "Makefile"]:
        if list(repo_path.glob(ext)):
            has_scripts = 1
            break
    
    has_infra = 0
    for pattern in ["Dockerfile", "*.tf"]:
        if list(repo_path.glob(pattern)):
            has_infra = 1
            break
    
    has_data = 0
    csvs = list(repo_path.rglob("*.csv"))
    if len(csvs) > 0:
        has_data = 1
    
    return {
        "language_count": lang_count,
        "primary_ratio": primary_pct,
        "secondary_ratio": secondary_pct,
        "has_docs": has_docs,
        "has_tests": has_tests,
        "has_scripts": has_scripts,
        "has_infrastructure": has_infra,
        "has_data": has_data,
    }

def collect_samples():
    """Collect samples from GitHub repos"""
    samples = []
    
    with tempfile.TemporaryDirectory() as tmpdir:
        for context, repos in REPOS.items():
            print(f"Collecting {context} samples...")
            
            for repo in repos:
                print(f"  Processing {repo}...")
                
                repo_name = repo.replace("/", "_")
                repo_dir = Path(tmpdir) / repo_name
                
                # Clone repo
                result = subprocess.run(
                    ["gh", "repo", "clone", repo, str(repo_dir), "--", "--depth", "1"],
                    capture_output=True
                )
                
                if result.returncode != 0:
                    print(f"    Failed to clone")
                    continue
                
                # Extract features
                features = run_chu_detect(repo_dir)
                
                if features is None:
                    print(f"    Failed to extract features")
                    continue
                
                features["context"] = context
                samples.append(features)
                print(f"    Added sample")
    
    return samples

def main():
    script_dir = Path(__file__).parent
    data_dir = script_dir.parent / "data"
    
    print("Collecting GitHub samples...")
    print()
    
    samples = collect_samples()
    
    print()
    print(f"Collected {len(samples)} samples")
    
    if len(samples) < 6:
        print("Not enough samples, keeping synthetic data")
        return
    
    # Save to CSV
    df = pd.DataFrame(samples)
    output_path = data_dir / "training_data_github.csv"
    df.to_csv(output_path, index=False)
    
    print(f"Saved to: {output_path}")
    print()
    print("Sample distribution:")
    print(df["context"].value_counts())
    print()
    
    # Merge with existing synthetic data
    synthetic_path = data_dir / "training_data.csv"
    if synthetic_path.exists():
        df_synthetic = pd.read_csv(synthetic_path)
        df_combined = pd.concat([df_synthetic, df], ignore_index=True)
        combined_path = data_dir / "training_data_combined.csv"
        df_combined.to_csv(combined_path, index=False)
        print(f"Combined with synthetic: {combined_path}")
        print(f"Total samples: {len(df_combined)}")
    
    print()
    print("Retraining model...")
    
    # Import and run training
    import sys
    sys.path.insert(0, str(script_dir))
    from train import train_model
    
    model_dir = script_dir.parent / "models"
    
    # Use combined data if available
    train_data_path = combined_path if 'combined_path' in locals() else output_path
    
    # Load and train
    df_train = pd.read_csv(train_data_path)
    print(f"Training on {len(df_train)} samples")
    
    # Save temporarily for train.py
    temp_train_path = data_dir / "training_data.csv"
    df_train.to_csv(temp_train_path, index=False)
    
    train_model(str(data_dir), str(model_dir))
    
    print()
    print("Done! Model retrained with real GitHub data")

if __name__ == "__main__":
    main()
