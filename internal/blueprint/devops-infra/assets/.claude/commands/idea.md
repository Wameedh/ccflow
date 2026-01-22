# /idea - Capture Infrastructure Request

Create a new infrastructure request specification with blast radius assessment.

## Usage

```
/idea [infrastructure request description]
```

## What This Command Does

1. **Gathers Requirements**: Asks clarifying questions about the infrastructure need
2. **Assesses Blast Radius**: Evaluates potential impact
3. **Creates State File**: Initializes tracking in `{{.DocsStateDir}}/`
4. **Suggests Next Steps**: Recommends moving to `/design`

## Process

### Step 1: Understand the Request
Ask the user to describe:
- What infrastructure change is needed?
- What is the business justification?
- Which environments will be affected (dev/staging/production)?
- Are there compliance requirements (SOC2, HIPAA, PCI)?
- What is the urgency level?
- What systems depend on this infrastructure?

### Step 2: Assess Blast Radius

| Level | Criteria | Examples |
|-------|----------|----------|
| **Low** | Non-production, isolated | Dev environment changes |
| **Medium** | Staging or limited production | Staging infra, isolated services |
| **High** | Production, many users | Production databases, APIs |
| **Critical** | Core infrastructure, all users | Networking, IAM, security |

### Step 3: Create State File
Create `{{.DocsStateDir}}/<feature-id>.json`:

```json
{
  "id": "<feature-id>",
  "title": "<Request Title>",
  "description": "<User's description>",
  "status": "ideation",
  "blast_radius": "low|medium|high|critical",
  "environments": ["dev", "staging", "production"],
  "acceptance_criteria": [],
  "dependencies": [],
  "created_at": "<ISO timestamp>",
  "updated_at": "<ISO timestamp>"
}
```

### Step 4: Define Acceptance Criteria
Work with the user to define clear, verifiable criteria:
- Infrastructure created/modified correctly
- Security scans pass
- Cost within budget
- No unplanned downtime
- Rollback procedure verified

### Step 5: Output Summary
Print a summary including:
- Request ID and title
- Blast radius assessment
- Target environments
- Key requirements
- Suggested next step: `/design <feature-id>`

## Example

```
User: /idea Add a new RDS instance for the analytics team