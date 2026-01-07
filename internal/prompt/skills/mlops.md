
# MLOps

Guidelines for machine learning operations, model lifecycle management, and ML pipelines.

## When to Activate

- Deploying ML models
- Setting up experiment tracking
- Building ML pipelines
- Model monitoring and drift detection

## Experiment Tracking

### MLflow setup

```python
import mlflow

# Set tracking URI
mlflow.set_tracking_uri("http://mlflow.internal:5000")
mlflow.set_experiment("recommendation-model")

with mlflow.start_run(run_name="xgboost-v2"):
    # Log parameters
    mlflow.log_param("learning_rate", 0.01)
    mlflow.log_param("max_depth", 6)
    mlflow.log_param("n_estimators", 100)
    
    # Train model
    model = train_model(params)
    
    # Log metrics
    mlflow.log_metric("accuracy", 0.95)
    mlflow.log_metric("f1_score", 0.93)
    mlflow.log_metric("auc_roc", 0.97)
    
    # Log model artifact
    mlflow.sklearn.log_model(model, "model")
    
    # Log additional artifacts
    mlflow.log_artifact("confusion_matrix.png")
    mlflow.log_artifact("feature_importance.csv")
```

### Weights & Biases

```python
import wandb

wandb.init(
    project="recommendation-model",
    config={
        "learning_rate": 0.01,
        "epochs": 100,
        "batch_size": 32,
    }
)

for epoch in range(epochs):
    train_loss = train_epoch(model, train_loader)
    val_loss, val_accuracy = evaluate(model, val_loader)
    
    wandb.log({
        "epoch": epoch,
        "train_loss": train_loss,
        "val_loss": val_loss,
        "val_accuracy": val_accuracy,
    })

# Log model
wandb.save("model.pth")
wandb.finish()
```

## Model Serving

### FastAPI model server

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import joblib
import numpy as np

app = FastAPI(title="ML Model API")

# Load model at startup
model = None

@app.on_event("startup")
async def load_model():
    global model
    model = joblib.load("model.joblib")

class PredictionRequest(BaseModel):
    features: list[float]

class PredictionResponse(BaseModel):
    prediction: float
    confidence: float

@app.post("/predict", response_model=PredictionResponse)
async def predict(request: PredictionRequest):
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    
    features = np.array(request.features).reshape(1, -1)
    prediction = model.predict(features)[0]
    confidence = model.predict_proba(features).max()
    
    return PredictionResponse(
        prediction=float(prediction),
        confidence=float(confidence)
    )

@app.get("/health")
async def health():
    return {"status": "healthy", "model_loaded": model is not None}
```

### Docker for ML

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy model and code
COPY model.joblib .
COPY app.py .

# Non-root user
RUN useradd -m appuser
USER appuser

EXPOSE 8000
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Feature Store

### Feast configuration

```yaml
# feature_store.yaml
project: recommendation
registry: s3://feature-store/registry.db
provider: aws
online_store:
  type: redis
  connection_string: redis://redis:6379
offline_store:
  type: redshift
  cluster_id: feature-store
  region: us-east-1
  database: features
```

### Feature definitions

```python
from feast import Entity, Feature, FeatureView, FileSource
from feast.types import Float32, Int64
from datetime import timedelta

# Entity definition
user = Entity(
    name="user_id",
    join_keys=["user_id"],
    description="User identifier"
)

# Feature source
user_features_source = FileSource(
    path="s3://features/user_features.parquet",
    timestamp_field="event_timestamp",
)

# Feature view
user_features = FeatureView(
    name="user_features",
    entities=[user],
    ttl=timedelta(days=1),
    schema=[
        Feature(name="avg_purchase_amount", dtype=Float32),
        Feature(name="purchase_count", dtype=Int64),
        Feature(name="days_since_last_purchase", dtype=Int64),
    ],
    source=user_features_source,
)
```

## ML Pipeline

### Kubeflow pipeline

```python
from kfp import dsl
from kfp.components import create_component_from_func

@create_component_from_func
def preprocess_data(input_path: str, output_path: str):
    import pandas as pd
    
    df = pd.read_parquet(input_path)
    # Preprocessing logic
    df_processed = preprocess(df)
    df_processed.to_parquet(output_path)
    return output_path

@create_component_from_func
def train_model(data_path: str, model_path: str, params: dict):
    import joblib
    from sklearn.ensemble import RandomForestClassifier
    
    data = load_data(data_path)
    model = RandomForestClassifier(**params)
    model.fit(data.X, data.y)
    joblib.dump(model, model_path)
    return model_path

@create_component_from_func
def evaluate_model(model_path: str, test_data_path: str) -> float:
    import joblib
    
    model = joblib.load(model_path)
    test_data = load_data(test_data_path)
    accuracy = model.score(test_data.X, test_data.y)
    return accuracy

@dsl.pipeline(name="ML Training Pipeline")
def training_pipeline(input_data: str):
    preprocess_task = preprocess_data(input_data, "/tmp/processed")
    
    train_task = train_model(
        preprocess_task.output,
        "/tmp/model.joblib",
        {"n_estimators": 100, "max_depth": 10}
    )
    
    evaluate_task = evaluate_model(
        train_task.output,
        "/tmp/test_data"
    )
```

## Model Monitoring

### Drift detection

```python
from evidently import ColumnMapping
from evidently.report import Report
from evidently.metrics import (
    DataDriftTable,
    DatasetDriftMetric,
    ColumnDriftMetric,
)

def detect_drift(reference_data, current_data):
    report = Report(metrics=[
        DatasetDriftMetric(),
        DataDriftTable(),
        ColumnDriftMetric(column_name="feature_1"),
    ])
    
    report.run(
        reference_data=reference_data,
        current_data=current_data,
    )
    
    result = report.as_dict()
    
    if result["metrics"][0]["result"]["dataset_drift"]:
        alert_team("Data drift detected!")
        trigger_retraining()
    
    return result
```

### Performance monitoring

```python
import prometheus_client as prom

# Metrics
prediction_latency = prom.Histogram(
    'model_prediction_latency_seconds',
    'Model prediction latency',
    buckets=[.01, .025, .05, .1, .25, .5, 1.0]
)

prediction_count = prom.Counter(
    'model_prediction_total',
    'Total predictions',
    ['model_version', 'outcome']
)

feature_distribution = prom.Histogram(
    'model_feature_value',
    'Feature value distribution',
    ['feature_name'],
    buckets=[0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0]
)

@app.post("/predict")
async def predict(request: PredictionRequest):
    with prediction_latency.time():
        result = model.predict(request.features)
    
    prediction_count.labels(
        model_version="v2",
        outcome="success"
    ).inc()
    
    # Track feature distributions
    for i, value in enumerate(request.features):
        feature_distribution.labels(f"feature_{i}").observe(value)
    
    return result
```

## Model Registry

### MLflow model registry

```python
import mlflow
from mlflow.tracking import MlflowClient

client = MlflowClient()

# Register model
model_uri = f"runs:/{run_id}/model"
model_version = mlflow.register_model(model_uri, "recommendation-model")

# Transition to staging
client.transition_model_version_stage(
    name="recommendation-model",
    version=model_version.version,
    stage="Staging"
)

# After validation, promote to production
client.transition_model_version_stage(
    name="recommendation-model",
    version=model_version.version,
    stage="Production"
)

# Load production model
model = mlflow.pyfunc.load_model(
    model_uri="models:/recommendation-model/Production"
)
```

## A/B Testing Models

```python
import random

class ModelRouter:
    def __init__(self):
        self.models = {
            "control": load_model("v1"),
            "treatment": load_model("v2"),
        }
        self.traffic_split = {"control": 0.9, "treatment": 0.1}
    
    def route(self, user_id: str) -> str:
        # Deterministic routing based on user_id
        hash_val = hash(user_id) % 100
        cumulative = 0
        for variant, weight in self.traffic_split.items():
            cumulative += weight * 100
            if hash_val < cumulative:
                return variant
        return "control"
    
    def predict(self, user_id: str, features):
        variant = self.route(user_id)
        model = self.models[variant]
        
        prediction = model.predict(features)
        
        # Log for analysis
        log_experiment(user_id, variant, prediction)
        
        return prediction
```
