# /idea - Capture a Data Science Idea

Create a new feature specification for data science and ML projects.

## CRITICAL: User Input Required

**YOU MUST follow these rules:**

1. **DO NOT** create any files until Step 4 is complete and user confirms
2. **DO NOT** generate, invent, or make up feature ideas
3. **DO NOT** proceed past any step until you have the user's response
4. **ALWAYS** capture what the USER describes, not what you imagine

---

## Step 0: Check for User Input

**MODE A - User provided description:**
If the user wrote something like `/idea build a churn prediction model`:
- Acknowledge their description: "I understand you want to capture an idea about: [their description]"
- Continue to Step 1

**MODE B - No description provided:**
If the user only typed `/idea` with nothing after:
- **STOP IMMEDIATELY**
- Say: "What data science problem would you like to solve? Please describe the problem."
- **WAIT** - Do not continue until the user responds

---

## Step 1: Clarify the Problem

Ask the user:

"What data problem are we solving? What type of ML task is this (classification, regression, clustering, etc.)?"

**WAIT for the user's response before continuing.**

---

## Step 2: Define Data and Metrics

Ask the user:

"What data is available? What are the success metrics (accuracy, F1, RMSE, business KPI)?"

**WAIT for the user's response before continuing.**

---

## Step 3: Identify Constraints

Ask the user:

"Are there any constraints? (latency, model size, interpretability, baseline to beat)"

**WAIT for the user's response before continuing.**

---

## Step 4: Confirm Understanding (Gate)

**BEFORE creating any files**, present a summary:

```
## Feature Summary

**Title:** [derived from user's description]
**Problem Type:** [from Step 1]
**Data & Metrics:** [from Step 2]
**Constraints:** [from Step 3]

**Proposed Success Criteria:**
1. [criterion based on user input]
2. [criterion based on user input]
3. [criterion based on user input]
```

Then ask: "Does this capture your idea correctly? (yes/no/changes needed)"

**DO NOT proceed to Step 5 without explicit user confirmation.**

---

## Step 5: Create State File

Only after user confirms in Step 4, create `{{.DocsStateDir}}/<feature-id>.json`:

```json
{
  "id": "<feature-id>",
  "title": "<Feature Title>",
  "description": "<User's description from Steps 1-3>",
  "status": "ideation",
  "acceptance_criteria": ["<from Step 4>"],
  "dependencies": [],
  "datasets": [],
  "created_at": "<ISO timestamp>",
  "updated_at": "<ISO timestamp>"
}
```

---

## Step 6: Output Summary and Next Steps

Print:
- Confirmation that the idea has been captured
- The file path where it was saved
- Suggested next step: `/design <feature-id>`
