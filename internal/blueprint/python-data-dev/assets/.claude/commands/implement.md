# /implement - Implement a Feature

Implement data pipelines and ML models according to an approved design.

## Usage

```
/implement [feature-id]
```

## What This Command Does

1. **Loads Design**: Reads design doc from `{{.DocsDesignDir}}/`
2. **Implements Pipeline**: Creates data processing and model code
3. **Writes Tests**: Creates tests for pipelines and models
4. **Updates State**: Tracks implementation progress

## Prerequisites

Before using this command:
- Feature must have a design document
- Design should be approved
- Understand existing patterns in the codebase

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Review success metrics

### Step 2: Update State
```json
{
  "status": "implementation",
  "implementation_started_at": "<ISO timestamp>",
  "branch": "feature/<feature-id>"
}
```

### Step 3: Implement Data Pipeline
Follow the design document:

```python
# src/data/load.py
import pandas as pd
from pathlib import Path

def load_raw_data(path: Path) -> pd.DataFrame:
    """Load and validate raw data."""
    df = pd.read_csv(path)
    # Validation
    return df

# src/features/transform.py
from sklearn.base import BaseEstimator, TransformerMixin

class FeatureTransformer(BaseEstimator, TransformerMixin):
    """Custom feature transformer."""

    def fit(self, X, y=None):
        return self

    def transform(self, X):
        # Transform logic
        return X_transformed
```

### Step 4: Implement Model Training
```python
# src/models/train.py
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import cross_val_score
import mlflow

def train_model(X_train, y_train, params: dict):
    """Train model with experiment tracking."""
    with mlflow.start_run():
        mlflow.log_params(params)

        model = RandomForestClassifier(**params)
        cv_scores = cross_val_score(model, X_train, y_train, cv=5)

        mlflow.log_metric("cv_mean", cv_scores.mean())
        mlflow.log_metric("cv_std", cv_scores.std())

        model.fit(X_train, y_train)
        return model
```

### Step 5: Write Tests
```python
# tests/test_pipeline.py
import pytest

def test_load_data_valid():
    df = load_raw_data("test_data.csv")
    assert len(df) > 0

def test_feature_transformer():
    transformer = FeatureTransformer()
    X_out = transformer.fit_transform(X_test)
    assert X_out.shape[0] == X_test.shape[0]
```

### Step 6: Verify Implementation
- Run `pytest` to verify tests pass
- Run `ruff check .` for linting
- Run `mypy .` for type checking
- Verify pipeline reproducibility

### Step 7: Update State
```json
{
  "status": "experimentation",
  "implementation_completed_at": "<ISO timestamp>",
  "files_changed": ["list", "of", "files"],
  "tests_added": ["list", "of", "test", "files"]
}
```

### Step 8: Output Summary
- List of files changed
- Tests added
- Any deviations from design
- Suggested next step: `/experiment <feature-id>` or `/review <feature-id>`

## Guidelines

- Make small, focused commits
- Write tests alongside code
- Set random seeds for reproducibility
- Don't commit secrets or API keys
- Use type hints for functions
- Log important operations
- Track experiments systematically
