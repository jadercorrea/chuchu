#!/usr/bin/env python3
import json
import numpy as np
import sys
from pathlib import Path

def load_model(model_path):
    with open(model_path, 'r') as f:
        return json.load(f)

def extract_model_features(model_id):
    model_lower = model_id.lower()
    
    has_coder = 1 if 'coder' in model_lower or 'code' in model_lower else 0
    has_instant = 1 if 'instant' in model_lower or 'flash' in model_lower else 0
    
    model_size = 0
    for size in ['405b', '120b', '70b', '32b', '33b', '22b', '9b', '8b', '3b']:
        if size in model_lower:
            model_size = int(size.replace('b', ''))
            break
    
    return has_coder, has_instant, model_size

def predict(model_data, model_id, action, language, complexity, context_window, cost_per_1m):
    encoders = model_data['encoders']
    
    if action not in encoders['action']:
        print(f"Warning: Unknown action '{action}', using default")
        action_encoded = 0
    else:
        action_encoded = encoders['action'][action]
    
    if language not in encoders['language']:
        print(f"Warning: Unknown language '{language}', using default")
        language_encoded = 0
    else:
        language_encoded = encoders['language'][language]
    
    if complexity not in encoders['complexity']:
        print(f"Warning: Unknown complexity '{complexity}', using default")
        complexity_encoded = 0
    else:
        complexity_encoded = encoders['complexity'][complexity]
    
    has_coder, has_instant, model_size = extract_model_features(model_id)
    
    log_cost = np.log1p(cost_per_1m)
    log_context = np.log1p(context_window)
    
    features = np.array([
        action_encoded, language_encoded, complexity_encoded,
        log_cost, log_context, has_coder, has_instant, model_size
    ])
    
    coef = np.array(model_data['coefficients'][0])
    intercept = model_data['intercept'][0]
    
    logit = np.dot(features, coef) + intercept
    prob = 1 / (1 + np.exp(-logit))
    
    return prob

def main():
    script_dir = Path(__file__).parent.parent
    model_path = script_dir / 'model.json'
    
    if not model_path.exists():
        print(f"Error: Model not found at {model_path}")
        print("Run 'gptcode ml train recommender' first")
        sys.exit(1)
    
    model_data = load_model(model_path)
    
    if len(sys.argv) < 6:
        print("Usage: predict.py <model_id> <action> <language> <complexity> <context_window> <cost_per_1m>")
        print("\nExample:")
        print("  predict.py 'qwen-2.5-coder-32b' edit go complex 32000 0.14")
        print("\nInteractive mode:")
        
        model_id = input("Model ID: ")
        action = input("Action (edit/review/plan/research): ")
        language = input("Language (go/python/typescript): ")
        complexity = input("Complexity (simple/complex/multistep): ")
        context_window = int(input("Context window: "))
        cost_per_1m = float(input("Cost per 1M tokens: "))
    else:
        model_id = sys.argv[1]
        action = sys.argv[2]
        language = sys.argv[3]
        complexity = sys.argv[4]
        context_window = int(sys.argv[5])
        cost_per_1m = float(sys.argv[6])
    
    prob = predict(model_data, model_id, action, language, complexity, context_window, cost_per_1m)
    
    print(f"\nPrediction for:")
    print(f"  Model: {model_id}")
    print(f"  Task: {action} {language} code ({complexity})")
    print(f"  Context: {context_window:,} tokens")
    print(f"  Cost: ${cost_per_1m:.2f}/1M")
    print(f"\nSuccess probability: {prob:.2%}")
    print(f"Recommendation: {'✓ Good choice' if prob > 0.7 else '⚠ Consider alternatives' if prob > 0.5 else '✗ Not recommended'}")

if __name__ == '__main__':
    main()
