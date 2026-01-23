# /idea - Capture Infrastructure Request

Create a new infrastructure request specification with blast radius assessment.

## CRITICAL: User Input Required

**YOU MUST follow these rules:**

1. **DO NOT** create any files until Step 5 is complete and user confirms
2. **DO NOT** generate, invent, or make up infrastructure requests
3. **DO NOT** proceed past any step until you have the user's response
4. **ALWAYS** capture what the USER describes, not what you imagine

---

## Step 0: Check for User Input

**MODE A - User provided description:**
If the user wrote something like `/idea add a new RDS instance`:
- Acknowledge their description: "I understand you want to capture an infrastructure request about: [their description]"
- Continue to Step 1

**MODE B - No description provided:**
If the user only typed `/idea` with nothing after:
- **STOP IMMEDIATELY**
- Say: "What infrastructure change do you need? Please describe the request."
- **WAIT** - Do not continue until the user responds

---

## Step 1: Clarify the Request

Ask the user:

"What infrastructure change is needed? What is the business justification?"

**WAIT for the user's response before continuing.**

---

## Step 2: Define Scope

Ask the user:

"Which environments will be affected (dev/staging/production)? What systems depend on this?"

**WAIT for the user's response before continuing.**

---

## Step 3: Identify Requirements

Ask the user:

"Are there compliance requirements (SOC2, HIPAA, PCI)? What is the urgency level?"

**WAIT for the user's response before continuing.**

---

## Step 4: Assess Blast Radius

Based on the user's answers, determine blast radius:

| Level | Criteria | Examples |
|-------|----------|----------|
| **Low** | Non-production, isolated | Dev environment changes |
| **Medium** | Staging or limited production | Staging infra, isolated services |
| **High** | Production, many users | Production databases, APIs |
| **Critical** | Core infrastructure, all users | Networking, IAM, security |

Present your assessment and ask: "I've assessed this as [level] blast radius because [reason]. Does this seem right?"

**WAIT for the user's response before continuing.**

---

## Step 5: Confirm Understanding (Gate)

**BEFORE creating any files**, present a summary:

```
## Infrastructure Request Summary

**Title:** [derived from user's description]
**Request:** [from Step 1]
**Environments:** [from Step 2]
**Requirements:** [from Step 3]
**Blast Radius:** [from Step 4]

**Proposed Acceptance Criteria:**
1. Infrastructure created/modified correctly
2. Security scans pass
3. Cost within budget
4. Rollback procedure verified
```

Then ask: "Does this capture your request correctly? (yes/no/changes needed)"

**DO NOT proceed to Step 6 without explicit user confirmation.**

---

## Step 6: Create State File

Only after user confirms in Step 5, create `{{.DocsStateDir}}/<feature-id>.json`:

```json
{
  "id": "<feature-id>",
  "title": "<Request Title>",
  "description": "<User's description from Steps 1-4>",
  "status": "ideation",
  "blast_radius": "low|medium|high|critical",
  "environments": ["dev", "staging", "production"],
  "acceptance_criteria": ["<from Step 5>"],
  "dependencies": [],
  "created_at": "<ISO timestamp>",
  "updated_at": "<ISO timestamp>"
}
```

---

## Step 7: Output Summary and Next Steps

Print:
- Confirmation that the request has been captured
- The file path where it was saved
- Blast radius level with explanation
- Suggested next step: `/design <feature-id>`
