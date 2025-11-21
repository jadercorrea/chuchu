#!/usr/bin/env python3
import sys
import json
import pandas as pd
import numpy as np
from pathlib import Path
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import classification_report, confusion_matrix, accuracy_score

SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
MODEL_PATH = PROJECT_ROOT / "models" / "complexity_model.json"
DEFAULT_EVAL = PROJECT_ROOT / "data" / "eval.csv"
DEFAULT_TRAIN = PROJECT_ROOT / "data" / "training_data.csv"

def softmax(x):
    e = np.exp(x - np.max(x))
    return e / e.sum()

def load_model():
    with open(MODEL_PATH, "r") as f:
        return json.load(f)

def load_dataset(path: Path):
    if path is None:
        if DEFAULT_EVAL.exists():
            path = DEFAULT_EVAL
        else:
            path = DEFAULT_TRAIN
    df = pd.read_csv(path)
    return df

def predict_batch(model, texts):
    vec = TfidfVectorizer(vocabulary=model["tfidf"]["vocabulary"])   
    vec.idf_ = np.array(model["tfidf"]["idf_weights"])             
    X = vec.transform(texts).toarray()
    coefs = np.array(model["classifier"]["coefficients"])          
    intercepts = np.array(model["classifier"]["intercepts"])       
    classes = model["classifier"]["classes"]                        
    logits = X.dot(coefs.T) + intercepts
    probs = np.apply_along_axis(softmax, 1, logits)
    idx = probs.argmax(axis=1)
    preds = [classes[i] for i in idx]
    return preds

def main():
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else None
    model = load_model()
    df = load_dataset(path)
    texts = df["message"].astype(str).tolist()
    y_true = df["label"].tolist()
    label_map = {"simple":0, "complex":1, "multistep":2}
    y_pred_text = predict_batch(model, texts)
    y_pred = [label_map[x] for x in y_pred_text]
    acc = accuracy_score(y_true, y_pred)
    print(f"Accuracy: {acc:.3f}")
    print(classification_report(y_true, y_pred, target_names=["simple","complex","multistep"], digits=3))
    cm = confusion_matrix(y_true, y_pred)
    print("Confusion Matrix:")
    for row in cm:
        print(" ".join(str(int(v)) for v in row))

if __name__ == "__main__":
    main()
