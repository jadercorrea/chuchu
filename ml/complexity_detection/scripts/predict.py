#!/usr/bin/env python3
"""
Test the trained model with interactive predictions.

Usage:
    python scripts/predict.py "implement user authentication"
    python scripts/predict.py  # Interactive mode
"""

import json
import sys
from pathlib import Path
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer

# Paths
SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
MODEL_PATH = PROJECT_ROOT / "models" / "complexity_model.json"

def softmax(x):
    """Compute softmax values for array x."""
    exp_x = np.exp(x - np.max(x))
    return exp_x / exp_x.sum()

def load_model():
    """Load model from JSON."""
    if not MODEL_PATH.exists():
        print(f"❌ Model not found at {MODEL_PATH}")
        print("   Run 'python scripts/train.py' first")
        sys.exit(1)
    
    with open(MODEL_PATH, 'r') as f:
        model = json.load(f)
    
    return model

def predict(model, message):
    vectorizer = TfidfVectorizer(vocabulary=model['tfidf']['vocabulary'])
    vectorizer.idf_ = np.array(model['tfidf']['idf_weights'])
    X = vectorizer.transform([message])
    X_array = X.toarray()[0]
    coefs = np.array(model['classifier']['coefficients'])
    intercepts = np.array(model['classifier']['intercepts'])
    classes = model['classifier']['classes']
    scores = np.dot(coefs, X_array) + intercepts
    text = f" {message.lower()} "
    idx_simple = classes.index('simple') if 'simple' in classes else 0
    idx_complex = classes.index('complex') if 'complex' in classes else 1
    idx_multistep = classes.index('multistep') if 'multistep' in classes else 2
    multistep_cues = [' then ', ' and then ', ', then ', '; then ', ' after ', ' followed by ', ' first ', ' second ', ' third ']
    complex_cues = ['oauth', 'oauth2', 'oidc', 'migrate', 'migration', 'deploy', 'docker', 'kubectl', 'k8s', 's3', 'pipeline', 'airflow', 'terraform', 'ansible', 'kafka', 'stripe', 'payment', 'upload', 'script', 'bash', 'nginx', 'logs']
    if any(c in text for c in multistep_cues):
        scores[idx_multistep] += 1.0
    if any(c in text for c in complex_cues):
        scores[idx_complex] += 1.5
    probs = softmax(scores)
    pred_idx = int(np.argmax(probs))
    return {
        'label': classes[pred_idx],
        'confidence': float(probs[pred_idx]),
        'probabilities': {classes[i]: float(probs[i]) for i in range(len(classes))}
    }

def interactive_mode(model):
    """Interactive prediction mode."""
    print("\n🤖 Task Complexity Predictor")
    print("=" * 50)
    print(f"Model version: {model['metadata']['version']}")
    print(f"Accuracy: {model['metadata']['accuracy']:.1%}")
    print("=" * 50)
    print("\nEnter task descriptions (or 'quit' to exit):\n")
    
    while True:
        try:
            message = input("➤ ").strip()
            
            if not message:
                continue
            
            if message.lower() in ['quit', 'exit', 'q']:
                print("\n👋 Goodbye!")
                break
            
            result = predict(model, message)
            
            print(f"\n   Prediction: {result['label'].upper()}")
            print(f"   Confidence: {result['confidence']:.1%}")
            print("\n   All probabilities:")
            for label, prob in sorted(result['probabilities'].items(), key=lambda x: x[1], reverse=True):
                bar = "█" * int(prob * 40)
                print(f"      {label:>10}: {prob:>6.1%} {bar}")
            print()
            
        except KeyboardInterrupt:
            print("\n\n👋 Goodbye!")
            break
        except Exception as e:
            print(f"❌ Error: {e}\n")

def main():
    """Main entry point."""
    model = load_model()
    
    if len(sys.argv) > 1:
        # Single prediction mode
        message = " ".join(sys.argv[1:])
        result = predict(model, message)
        
        print(f"\nMessage: {message}")
        print(f"Prediction: {result['label']}")
        print(f"Confidence: {result['confidence']:.1%}")
        print("\nProbabilities:")
        for label, prob in result['probabilities'].items():
            print(f"  {label}: {prob:.1%}")
    else:
        # Interactive mode
        interactive_mode(model)

if __name__ == "__main__":
    main()
