# /idea - Capture a New Feature Idea

Create a new feature specification and initialize workflow state tracking.

## CRITICAL: User Input Required

**YOU MUST follow these rules:**

1. **DO NOT** create any files until Step 4 is complete and user confirms
2. **DO NOT** generate, invent, or make up feature ideas
3. **DO NOT** proceed past any step until you have the user's response
4. **ALWAYS** capture what the USER describes, not what you imagine

---

## Step 0: Check for User Input

**MODE A - User provided description:**
If the user wrote something like `/idea add biometric login`:
- Acknowledge their description: "I understand you want to capture an idea about: [their description]"
- Continue to Step 1

**MODE B - No description provided:**
If the user only typed `/idea` with nothing after:
- **STOP IMMEDIATELY**
- Say: "What feature or idea would you like to capture? Please describe the problem you're trying to solve."
- **WAIT** - Do not continue until the user responds

---

## Step 1: Clarify the Problem

Ask the user:

"What problem does this feature solve? Who is the target user?"

**WAIT for the user's response before continuing.**

---

## Step 2: Define the Expected Behavior

Ask the user:

"What is the expected behavior? How should users interact with this feature?"

**WAIT for the user's response before continuing.**

---

## Step 3: Identify Constraints

Ask the user:

"Are there any constraints or requirements? (iOS version, device support, accessibility, etc.)"

**WAIT for the user's response before continuing.**

---

## Step 4: Confirm Understanding (Gate)

**BEFORE creating any files**, present a summary:

```
## Feature Summary

**Title:** [derived from user's description]
**Problem:** [from Step 1]
**Behavior:** [from Step 2]
**Constraints:** [from Step 3]

**Proposed Acceptance Criteria:**
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
