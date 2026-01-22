# /design - Create Technical Design

Create a technical design document for a data science feature.

## Usage

```
/design [feature-id]
```

## What This Command Does

1. **Loads Feature State**: Reads from `{{.DocsStateDir}}/<feature-id>.json`
2. **Creates Design Doc**: Generates `{{.DocsDesignDir}}/<feature-id>-design.md`
3. **Updates State**: Changes status to "design"
4. **Plans Pipeline**: Identifies data sources, features, and model approach

## Process

### Step 1: Load Feature Context
- Read the feature state file
- Understand requirements and success metrics
- Identify available data sources

### Step 2: Analyze Codebase
- Find similar pipelines in existing code
- Identify reusable transformers and utilities
- Note any dependencies or risks

### Step 3: Create Design Document

Write to `{{.DocsDesignDir}}/<feature-id>-design.md`:

```markdown
# Design: <Feature Title>

## Status
Design In Progress

## Problem Statement
<From the feature spec>

## Success Metrics
| Metric | Target | Baseline |
|--------|--------|----------|
| Accuracy | > 85% | 60% |
| F1 Score | > 0.80 | N/A |

## Data Pipeline

### Data Sources
| Source | Format | Volume | Update Frequency |
|--------|--------|--------|------------------|
| source1 | CSV | 1M rows | Daily |

### Data Processing
```
raw -> validate -> clean -> transform -> features -> train/test
```

### Feature Engineering
| Feature | Type | Description | Source |
|---------|------|-------------|--------|
| feature1 | numeric | ... | col_a |

## Model Architecture

### Approach
[Model type: Random Forest / Neural Network / etc.]

### Framework
[scikit-learn / PyTorch / etc.]

### Hyperparameters
| Parameter | Initial Value | Search Range |
|-----------|---------------|--------------|

## Evaluation Plan

### Offline Evaluation
- Cross-validation strategy
- Hold-out test set
- Metrics to track

### Online Evaluation (if applicable)
- A/B test design
- Monitoring plan

## Alternatives Considered
1. [Alternative 1] - [why rejected]

## File Changes
| File | Change Type | Description |
|------|-------------|-------------|

## Risks
- Data quality issues
- Model complexity
- Inference latency
```

### Step 4: Update State
```json
{
  "status": "design",
  "design_started_at": "<ISO timestamp>",
  "design_doc": "{{.DocsDesignDir}}/<feature-id>-design.md"
}
```

### Step 5: Output Summary
- Link to design document
- Key decisions made
- Open questions to resolve
- Suggested next step: `/implement <feature-id>`

## Guidelines

- Start with simple baselines
- Design for reproducibility
- Consider data versioning
- Plan for model monitoring
- Document assumptions about data
