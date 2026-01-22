# /experiment - Track ML Experiments

Track and manage ML experiments with parameters, metrics, and artifacts.

## Usage

```
/experiment [action] [experiment-id]
```

### Actions
- `new` - Create a new experiment
- `start` - Start running an experiment
- `log` - Log metrics and parameters
- `complete` - Mark experiment as complete
- `list` - List all experiments
- `compare` - Compare multiple experiments

## What This Command Does

1. **Creates Experiment**: Initializes experiment tracking
2. **Logs Parameters**: Records hyperparameters and configuration
3. **Tracks Metrics**: Logs evaluation metrics
4. **Saves Artifacts**: Records model files, plots, etc.
5. **Enables Comparison**: Facilitates experiment comparison

## Process

### Creating a New Experiment

```
/experiment new churn-model-v2
```

Creates `{{.DocsStateDir}}/experiments/exp-churn-model-v2.json`:

```json
{
  "id": "exp-churn-model-v2",
  "name": "Churn Model V2",
  "description": "Testing XGBoost with engineered features",
  "status": "planned",
  "hypothesis": "Adding interaction features will improve F1 by 5%",
  "feature_id": "churn-prediction",
  "created_at": "ISO timestamp",
  "parameters": {},
  "metrics": {},
  "artifacts": []
}
```

### Starting an Experiment

```
/experiment start exp-churn-model-v2
```

Updates status to "running" and records start time.

### Logging Parameters and Metrics

```
/experiment log exp-churn-model-v2
```

Prompts to log:
- **Parameters**: hyperparameters, configuration
- **Metrics**: accuracy, F1, loss, etc.
- **Artifacts**: model files, plots

Updated experiment file:
```json
{
  "status": "running",
  "started_at": "ISO timestamp",
  "parameters": {
    "model_type": "xgboost",
    "n_estimators": 100,
    "max_depth": 6,
    "learning_rate": 0.1
  },
  "metrics": {
    "accuracy": 0.87,
    "f1_score": 0.82,
    "auc": 0.91
  },
  "dataset": {
    "name": "customer_data_v2",
    "train_size": 80000,
    "test_size": 20000
  },
  "artifacts": [
    {"name": "model", "path": "models/xgb_v2.joblib", "type": "model"},
    {"name": "confusion_matrix", "path": "plots/cm.png", "type": "plot"}
  ]
}
```

### Completing an Experiment

```
/experiment complete exp-churn-model-v2
```

Prompts for:
- Conclusion
- Next steps

```json
{
  "status": "completed",
  "completed_at": "ISO timestamp",
  "conclusion": "Interaction features improved F1 by 3%, below hypothesis but significant",
  "next_steps": [
    "Try feature selection to reduce dimensionality",
    "Experiment with different interaction terms"
  ]
}
```

### Listing Experiments

```
/experiment list
```

Output:
```
Experiments for: churn-prediction
┌────────────────────┬─────────┬────────────┬──────────┬────────┐
│ ID                 │ Status  │ Created    │ F1       │ AUC    │
├────────────────────┼─────────┼────────────┼──────────┼────────┤
│ exp-churn-v1       │ completed│ 2024-01-10│ 0.78     │ 0.85   │
│ exp-churn-v2       │ completed│ 2024-01-15│ 0.82     │ 0.91   │
│ exp-churn-v3       │ running │ 2024-01-20│ -        │ -      │
└────────────────────┴─────────┴────────────┴──────────┴────────┘
```

### Comparing Experiments

```
/experiment compare exp-churn-v1 exp-churn-v2
```

Output:
```
Experiment Comparison
────────────────────────────────────────

Parameters:
┌──────────────────┬────────────┬────────────┐
│ Parameter        │ exp-v1     │ exp-v2     │
├──────────────────┼────────────┼────────────┤
│ model_type       │ rf         │ xgboost    │
│ n_estimators     │ 100        │ 100        │
│ max_depth        │ 10         │ 6          │
└──────────────────┴────────────┴────────────┘

Metrics:
┌──────────────────┬────────────┬────────────┬─────────┐
│ Metric           │ exp-v1     │ exp-v2     │ Change  │
├──────────────────┼────────────┼────────────┼─────────┤
│ accuracy         │ 0.84       │ 0.87       │ +3.6%   │
│ f1_score         │ 0.78       │ 0.82       │ +5.1%   │
│ auc              │ 0.85       │ 0.91       │ +7.1%   │
└──────────────────┴────────────┴────────────┴─────────┘

Winner: exp-churn-v2 (higher F1, AUC)
```

## Integration with MLflow

If MLflow is available, experiments are also logged there:

```python
import mlflow

with mlflow.start_run(run_name="exp-churn-v2"):
    mlflow.log_params(params)
    mlflow.log_metrics(metrics)
    mlflow.log_artifact("model.joblib")
```

## Guidelines

- Document hypothesis before running
- Log all relevant parameters
- Track data versions alongside experiments
- Record conclusions and learnings
- Use consistent metric names
- Save reproducible artifacts
