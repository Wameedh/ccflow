# /idea - Capture a Data Science Idea

Create a new feature specification for data science and ML projects.

## Usage

```
/idea [feature name or description]
```

## What This Command Does

1. **Gathers Requirements**: Asks clarifying questions about the data problem
2. **Creates State File**: Initializes tracking in `{{.DocsStateDir}}/`
3. **Defines Success Metrics**: Establishes measurable criteria
4. **Suggests Next Steps**: Recommends moving to `/design`

## Process

### Step 1: Understand the Problem
Ask the user to describe:
- What data problem are we solving?
- What type of ML task is this (classification, regression, etc.)?
- What data is available?
- What are the success metrics (accuracy, F1, RMSE, business KPI)?
- What is the baseline performance to beat?
- Are there constraints (latency, model size, interpretability)?

### Step 2: Create State File
Create `{{.DocsStateDir}}/<feature-id>.json`:

```json
{
  "id": "<feature-id>",
  "title": "<Feature Title>",
  "description": "<User's description>",
  "status": "ideation",
  "acceptance_criteria": [],
  "dependencies": [],
  "datasets": [],
  "created_at": "<ISO timestamp>",
  "updated_at": "<ISO timestamp>"
}
```

### Step 3: Define Success Criteria
Work with the user to define clear, measurable criteria:
- Performance metrics with target thresholds
- Data quality requirements
- Inference requirements (latency, throughput)
- Business impact metrics

### Step 4: Output Summary
Print a summary including:
- Feature ID and title
- Problem type
- Success metrics and targets
- Data requirements
- Suggested next step: `/design <feature-id>`

## Example

```
User: /idea Build a customer churn prediction model