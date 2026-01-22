# /status - Check Workflow Status

Display the current state of the workflow, features, and experiments.

## Usage

```
/status [feature-id]
```

If no feature-id is provided, shows overview of all features.

## What This Command Does

1. **Lists Features**: Shows all features in the workflow
2. **Shows Experiments**: Displays experiment status and results
3. **Model Performance**: Shows current model metrics
4. **Next Steps**: Suggests appropriate actions

## Process

### Overview Mode (no feature-id)

Scan `{{.DocsStateDir}}/` for all state files and display:

```
Workflow: {{.WorkflowName}}
Blueprint: python-data-dev
State Directory: {{.DocsStateDir}}

Active Features:
┌─────────────────┬──────────────────┬───────────────┬─────────────────┐
│ ID              │ Title            │ Status        │ Last Updated    │
├─────────────────┼──────────────────┼───────────────┼─────────────────┤
│ churn-model     │ Churn Prediction │ experimentation│ 2 hours ago    │
│ rec-system      │ Recommendations  │ design        │ 1 day ago       │
│ fraud-detect    │ Fraud Detection  │ review        │ 30 minutes ago  │
└─────────────────┴──────────────────┴───────────────┴─────────────────┘

Recent Experiments:
┌─────────────────┬────────────┬────────┬────────┬─────────┐
│ ID              │ Feature    │ Status │ F1     │ AUC     │
├─────────────────┼────────────┼────────┼────────┼─────────┤
│ exp-churn-v3    │ churn-model│ running│ -      │ -       │
│ exp-churn-v2    │ churn-model│ done   │ 0.82   │ 0.91    │
│ exp-fraud-v1    │ fraud-detect│ done  │ 0.95   │ 0.98    │
└─────────────────┴────────────┴────────┴────────┴─────────┘

Summary:
- Ideation: 0
- Design: 1
- Implementation: 0
- Experimentation: 1
- Review: 1
- Released: 0
```

### Feature Detail Mode (with feature-id)

Display detailed status for a specific feature:

```
Feature: churn-model
Title: Customer Churn Prediction
Status: experimentation
Created: 2024-01-15T10:30:00Z
Last Updated: 2024-01-16T14:22:00Z

Success Metrics:
┌──────────────┬─────────┬──────────┬─────────┐
│ Metric       │ Target  │ Current  │ Status  │
├──────────────┼─────────┼──────────┼─────────┤
│ Accuracy     │ > 85%   │ 87%      │ ✓ Met   │
│ F1 Score     │ > 0.80  │ 0.82     │ ✓ Met   │
│ AUC          │ > 0.85  │ 0.91     │ ✓ Met   │
└──────────────┴─────────┴──────────┴─────────┘

Experiments:
┌─────────────────┬────────────┬────────┬────────┬──────────────────┐
│ ID              │ Status     │ F1     │ AUC    │ Notes            │
├─────────────────┼────────────┼────────┼────────┼──────────────────┤
│ exp-churn-v1    │ completed  │ 0.78   │ 0.85   │ RF baseline      │
│ exp-churn-v2    │ completed  │ 0.82   │ 0.91   │ XGBoost + features│
│ exp-churn-v3    │ running    │ -      │ -      │ Neural network   │
└─────────────────┴────────────┴────────┴────────┴──────────────────┘

Best Model: exp-churn-v2 (XGBoost)

Timeline:
- Ideation: 2024-01-15T10:30:00Z
- Design Started: 2024-01-15T14:00:00Z
- Design Completed: 2024-01-15T17:30:00Z
- Implementation Started: 2024-01-16T09:00:00Z
- Experimentation Started: 2024-01-16T12:00:00Z

Files Changed:
- src/data/load.py
- src/features/transform.py
- src/models/train.py
- notebooks/exploration.ipynb

Design Doc: {{.DocsDesignDir}}/churn-model-design.md

Next Step: Complete exp-churn-v3 or run /review churn-model
```

### Experiment Detail

```
/status exp-churn-v2
```

Shows detailed experiment information:

```
Experiment: exp-churn-v2
Name: XGBoost with Engineered Features
Status: completed
Feature: churn-model

Hypothesis:
Adding interaction features will improve F1 by 5%

Parameters:
┌──────────────────┬────────────┐
│ Parameter        │ Value      │
├──────────────────┼────────────┤
│ model_type       │ xgboost    │
│ n_estimators     │ 100        │
│ max_depth        │ 6          │
│ learning_rate    │ 0.1        │
└──────────────────┴────────────┘

Metrics:
┌──────────────────┬────────────┐
│ Metric           │ Value      │
├──────────────────┼────────────┤
│ accuracy         │ 0.87       │
│ f1_score         │ 0.82       │
│ auc              │ 0.91       │
│ cv_mean          │ 0.86       │
│ cv_std           │ 0.02       │
└──────────────────┴────────────┘

Artifacts:
- models/xgb_v2.joblib
- plots/confusion_matrix.png
- plots/feature_importance.png

Conclusion:
Interaction features improved F1 by 3%, below hypothesis but significant.

Next Steps:
1. Try feature selection to reduce dimensionality
2. Experiment with different interaction terms
```

## Guidelines

- Run `/status` regularly to track progress
- Compare experiments to find best approach
- Address stuck features promptly
- Document learnings from experiments
