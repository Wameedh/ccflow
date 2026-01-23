# /design - Create Technical Design

Create a technical design document for a data science feature.

## Usage

```
/design [feature-id]
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

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully:

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.design_to_implement.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- IMMEDIATELY invoke: `Skill(skill="implement", args="<feature-id>")`

### If mode is "prompt":
- Ask the user: "Ready to proceed to /implement <feature-id>?"
- If "Yes": invoke `Skill(skill="implement", args="<feature-id>")`
- If "No": print "Run /implement <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /implement <feature-id>"
