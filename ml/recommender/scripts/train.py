#!/usr/bin/env python3
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, classification_report
import json
import sys
from pathlib import Path

def load_data(data_path):
    df = pd.read_csv(data_path)
    print(f"Loaded {len(df)} training samples")
    print(f"Success rate: {df['success'].mean():.2%}")
    return df

def extract_features(df):
    le_action = LabelEncoder()
    le_language = LabelEncoder()
    le_complexity = LabelEncoder()
    
    df['action_encoded'] = le_action.fit_transform(df['action'])
    df['language_encoded'] = le_language.fit_transform(df['language'])
    df['complexity_encoded'] = le_complexity.fit_transform(df['complexity'])
    
    df['log_cost'] = np.log1p(df['cost_per_1m'])
    df['log_context'] = np.log1p(df['context_window'])
    
    features = [
        'action_encoded', 'language_encoded', 'complexity_encoded',
        'log_cost', 'log_context', 'has_coder_tag', 'has_instant_tag', 'model_size'
    ]
    
    X = df[features].values
    y = df['success'].values
    
    encoders = {
        'action': {v: i for i, v in enumerate(le_action.classes_)},
        'language': {v: i for i, v in enumerate(le_language.classes_)},
        'complexity': {v: i for i, v in enumerate(le_complexity.classes_)}
    }
    
    return X, y, features, encoders

def train_model(X, y):
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )
    
    model = LogisticRegression(
        max_iter=1000,
        class_weight='balanced',
        random_state=42
    )
    
    model.fit(X_train, y_train)
    
    y_pred = model.predict(X_test)
    accuracy = accuracy_score(y_test, y_pred)
    
    print(f"\nModel Performance:")
    print(f"Training samples: {len(X_train)}")
    print(f"Test samples: {len(X_test)}")
    print(f"Accuracy: {accuracy:.2%}")
    print("\nClassification Report:")
    print(classification_report(y_test, y_pred, target_names=['Failure', 'Success']))
    
    return model

def export_model(model, feature_names, encoders, output_path):
    model_data = {
        'features': feature_names,
        'coefficients': model.coef_.tolist(),
        'intercept': model.intercept_.tolist(),
        'encoders': encoders
    }
    
    with open(output_path, 'w') as f:
        json.dump(model_data, f, indent=2)
    
    print(f"\nModel exported to: {output_path}")
    print(f"Model size: {Path(output_path).stat().st_size / 1024:.2f} KB")

def main():
    script_dir = Path(__file__).parent.parent
    data_path = script_dir / 'data' / 'training_data.csv'
    output_path = script_dir / 'model.json'
    
    if not data_path.exists():
        print(f"Error: Training data not found at {data_path}")
        sys.exit(1)
    
    df = load_data(data_path)
    X, y, feature_names, encoders = extract_features(df)
    model = train_model(X, y)
    export_model(model, feature_names, encoders, output_path)
    
    print("\nTraining complete!")

if __name__ == '__main__':
    main()
