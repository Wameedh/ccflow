# /release - Release a Feature

Prepare and execute a data science model release.

## Usage

```
/release [feature-id]
```

## What This Command Does

1. **Verifies Gates**: Ensures all quality gates pass
2. **Packages Model**: Creates deployment artifacts
3. **Deployment Steps**: Provides deployment checklist
4. **Updates State**: Marks feature as released

## Prerequisites

Before using this command:
- Feature must be in "approved" status
- All tests must pass
- Model performance verified

## Process

### Step 1: Verify Release Gates

{{if .GatesEnabled}}
**Gates are ENABLED for this workflow.**

Required gates:
- [ ] All tests passing
- [ ] Model performance meets thresholds
- [ ] Code review approved
- [ ] No data quality issues
- [ ] Documentation complete
{{else}}
**Gates are DISABLED for this workflow.**
Proceeding without gate verification.
{{end}}

### Step 2: Pre-Release Checks

```bash
# Run full test suite
pytest --cov=src

# Verify model performance
python src/evaluate.py --model models/final_model.joblib

# Type checking
mypy src/

# Security scan
bandit -r src/
```

### Step 3: Package Model

```bash
# Export model with metadata
python src/export_model.py \
    --model models/final_model.joblib \
    --output dist/model_v1.0.0/

# Create model card
python src/create_model_card.py \
    --model dist/model_v1.0.0/ \
    --metrics results/metrics.json
```

### Step 4: Release Checklist

#### Before Deployment
- [ ] All automated checks pass
- [ ] Model performance validated on holdout set
- [ ] Model card created
- [ ] Rollback plan documented
- [ ] Monitoring configured

#### Deployment Steps

**Batch Inference:**
1. Upload model to artifact store
2. Update batch job configuration
3. Run on sample data
4. Verify outputs

**Real-time Inference:**
1. Build Docker image
2. Deploy to staging
3. Run load tests
4. Deploy to production
5. Monitor latency and errors

#### After Deployment
- [ ] Model serving correctly
- [ ] Latency within bounds
- [ ] No prediction errors
- [ ] Monitoring dashboards active

### Step 5: Update State
```json
{
  "status": "released",
  "release_completed_at": "<ISO timestamp>",
  "version": "<version>",
  "model_version": "1.0.0",
  "deployed_to": ["staging", "production"],
  "model_metrics": {
    "accuracy": 0.87,
    "f1_score": 0.82
  }
}
```

### Step 6: Output Summary
- Release status
- Version number
- Model performance summary
- Deployment locations
- Monitoring links

## Model Versioning

```
models/
├── v1.0.0/
│   ├── model.joblib
│   ├── metadata.json
│   ├── model_card.md
│   └── requirements.txt
├── v1.1.0/
│   └── ...
```

## Rollback Procedure

If issues are discovered:
1. Switch to previous model version
2. Update state to "rollback"
3. Investigate root cause
4. Run additional experiments
5. Re-release when ready

## Guidelines

- Never skip quality gates
- Always have a rollback plan
- Monitor models after deployment
- Document model limitations
- Keep model versions reproducible
- Track data dependencies
