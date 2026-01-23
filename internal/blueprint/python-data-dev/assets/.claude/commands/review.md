# /review - Review Implementation

Review data science code and validate model readiness.

## Usage

```
/review [feature-id]
```

## CRITICAL: Question Protocol

**YOU MUST FOLLOW THIS PROTOCOL - VIOLATIONS BREAK THE WORKFLOW**

### Before Starting Any Work

1. Read the state file for this feature
2. Check if `pending_questions` array exists with any `answered: false` items
3. If yes: Use AskUserQuestion tool for EACH unanswered question, then STOP
4. If no: Proceed with the command

### When You Need User Input

1. **STOP** all other work immediately
2. **DO NOT** write code, create files, or make decisions without user input
3. **USE** the AskUserQuestion tool (this blocks until user responds)
4. **WAIT** for the response before ANY further action
5. **UPDATE** the state file with the answer
6. **THEN** continue with the workflow

---

## What This Command Does

1. **Reviews Code**: Checks implementation against design and standards
2. **Validates Model**: Verifies model performance and reproducibility
3. **Security Check**: Scans for data and security issues
4. **Prepares PR**: Drafts pull request description
5. **Updates State**: Records review results

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Review experiment results if available

### Step 2: Run Automated Checks

```bash
# Code formatting
black --check .
isort --check .

# Linting
ruff check .

# Type checking
mypy .

# Tests
pytest --cov=src

# Security scan
bandit -r src/
```

### Step 3: Code Review Checklist

#### Code Quality
- [ ] Follows PEP 8 style guidelines
- [ ] Type hints present
- [ ] No unnecessary complexity
- [ ] Proper error handling
- [ ] No hardcoded paths or magic numbers

#### Data Pipeline
- [ ] Data validation implemented
- [ ] Missing value handling documented
- [ ] No data leakage between train/test
- [ ] Transformers follow sklearn API
- [ ] Pipeline is reproducible

#### Model Quality
- [ ] Model meets performance targets
- [ ] Cross-validation results are stable
- [ ] No significant overfitting
- [ ] Feature importance analyzed
- [ ] Edge cases handled

#### Reproducibility
- [ ] Random seeds set
- [ ] Dependencies pinned
- [ ] Experiment parameters logged
- [ ] Data versioning in place
- [ ] Results can be reproduced

#### Security & Privacy
- [ ] No secrets in code
- [ ] No PII in logs
- [ ] Input validation present
- [ ] Dependencies up to date

#### Documentation
- [ ] README updated
- [ ] Docstrings complete
- [ ] Experiment results documented
- [ ] Model card created (if applicable)

### Step 4: Validate Model Performance

```python
# Verify model meets acceptance criteria
assert metrics['f1_score'] >= 0.80, "F1 score below threshold"
assert metrics['accuracy'] >= 0.85, "Accuracy below threshold"

# Verify reproducibility
model1 = train_model(X, y, seed=42)
model2 = train_model(X, y, seed=42)
assert np.allclose(model1.predict(X), model2.predict(X))
```

### Step 5: Generate PR Description

```markdown
## Summary
<Brief description of the data science feature>

## Model Performance
| Metric | Target | Achieved |
|--------|--------|----------|
| Accuracy | > 85% | 87% |
| F1 Score | > 0.80 | 0.82 |

## Changes
- <Change 1>
- <Change 2>

## Experiments
- Best experiment: exp-model-v3
- Key finding: XGBoost outperforms RF by 4% F1

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Model reproducibility verified

## Related
- Design: {{.DocsDesignDir}}/<feature-id>-design.md
- Experiments: {{.DocsStateDir}}/experiments/
```

### Step 6: Update State
```json
{
  "status": "approved|changes_requested",
  "review_completed_at": "<ISO timestamp>",
  "review_checklist": {
    "code_quality": true,
    "model_quality": true,
    "reproducibility": true,
    "documentation": true
  },
  "model_metrics": {
    "accuracy": 0.87,
    "f1_score": 0.82
  }
}
```

### Step 7: Output Summary
- Review status
- Model performance summary
- Issues found (if any)
- PR description draft

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully (review approved):

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.review_to_release.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- IMMEDIATELY invoke: `Skill(skill="release", args="<feature-id>")`

### If mode is "prompt":
- Ask the user: "Ready to proceed to /release <feature-id>?"
- If "Yes": invoke `Skill(skill="release", args="<feature-id>")`
- If "No": print "Run /release <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /release <feature-id>"

**Note:** If the review requested changes, do NOT proceed to release.
