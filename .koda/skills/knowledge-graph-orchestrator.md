# knowledge-graph-orchestrator

**Version:** 1.0  
**Purpose:** Meta-agent orchestrator that delegates tasks to specialized agents  
**Status:** Active

---

## Overview

`knowledge-graph-orchestrator` is a meta-agent that:
1. Receives complex prompts requiring multiple areas of expertise
2. Analyzes the prompt and splits it into subtasks by area of responsibility
3. Delegates subtasks to appropriate specialized agents
4. Monitors execution and verifies completion
5. Aggregates results and returns unified response

---

## Specialized Agents (Delegation Targets)

### Available Agents

| Agent | Area of Responsibility | Use Cases |
|-------|----------------------|-----------|
| `knowledge-graph-frontend-svelte` | Frontend, Svelte 5, UI/UX, component tests | UI components, Svelte patterns, Playwright/Vitest, visual tests |
| `knowledge-graph-backend-go` | Backend Go, DDD, Docker, PostgreSQL, Redis, API | Go code, database, infrastructure, API endpoints |
| `knowledge-graph-docs-maintenance` | Documentation: README, docs/, ADR, changelog | Update docs, create ADR, maintain documentation |
| `knowledge-graph-testing` | All testing levels: Go unit/integration, Frontend unit/E2E/BDD, Python pytest | Write tests, debug failing tests, test coverage |
| `knowledge-graph-integration` | Backend ↔ Frontend integration: API contracts, DTOs, middleware | API mapping, type synchronization, integration tests |

---

## Delegation Logic

### Task Classification Rules

#### Frontend Tasks → `knowledge-graph-frontend-svelte`
- Svelte component creation/modification
- UI/UX design and implementation
- Frontend routing and state management
- Visual styling and theming
- Frontend unit tests (Vitest)
- E2E tests (Playwright)
- Component patterns and architecture

#### Backend Tasks → `knowledge-graph-backend-go`
- Go code implementation
- Database schemas and migrations
- API endpoint creation
- Docker infrastructure
- Redis/cache implementation
- Background workers
- Domain logic and business rules

#### Documentation Tasks → `knowledge-graph-docs-maintenance`
- README updates
- ADR creation/modification
- docs/ folder maintenance
- Changelog generation
- Configuration documentation
- API documentation
- Architecture diagrams

#### Testing Tasks → `knowledge-graph-testing`
- Unit test creation (any language)
- Integration test setup
- E2E/BDD test scenarios
- Test coverage analysis
- Debug failing tests
- Test infrastructure configuration

#### Integration Tasks → `knowledge-graph-integration`
- API endpoint mapping (backend ↔ frontend)
- DTO type synchronization
- CORS/auth middleware configuration
- Integration test creation
- API contract validation
- Request/response format alignment

---

## Prompt Processing Workflow

### Step 1: Prompt Analysis

```
Input: Complex user prompt
↓
Analyze for keywords and task types
↓
Identify required agents based on:
  - File paths mentioned
  - Technologies referenced
  - Task categories
  - Domain language
```

### Step 2: Task Splitting

```
Example Prompt: "Add new API endpoint for user settings and update frontend form"

Split into:
1. Backend: Create Go endpoint, request/response DTOs
   → Delegate to: knowledge-graph-backend-go
   
2. Frontend: Create settings form component
   → Delegate to: knowledge-graph-frontend-svelte
   
3. Integration: Map endpoint to frontend API client, sync types
   → Delegate to: knowledge-graph-integration
   
4. Tests: Add unit/integration tests
   → Delegate to: knowledge-graph-testing
   
5. Docs: Update API documentation
   → Delegate to: knowledge-graph-docs-maintenance
```

### Step 3: Agent Delegation

For each subtask:
1. Select appropriate agent
2. Format prompt for agent's context
3. Execute agent task
4. Capture results

### Step 4: Result Aggregation

```
Collect outputs from all agents
↓
Check for consistency and conflicts
↓
Resolve any integration issues
↓
Format unified response
```

### Step 5: Quality Verification

```
Verify:
- All subtasks completed
- No conflicting changes
- Tests pass
- Documentation updated
- Code follows patterns
```

---

## Usage Examples

### Example 1: Simple Frontend Task

**User Prompt:** "Add dark mode toggle to the header"

**Orchestrator Logic:**
1. Analyze: UI component change
2. Identify: Only frontend needed
3. Delegate: `knowledge-graph-frontend-svelte`
4. Execute: Single agent task
5. Return: Frontend implementation

**Result:** Direct delegation, no splitting needed

---

### Example 2: Full-Stack Feature

**User Prompt:** "Implement note sharing with email invitations"

**Orchestrator Logic:**
1. Analyze: Multiple domains involved
2. Split:
   - Backend: Share endpoint, email service, database schema
   - Frontend: Share modal, email input, invitation list
   - Integration: API mapping, type sync
   - Tests: Unit + integration + E2E
   - Docs: API docs, user guide
3. Delegate: All 5 agents
4. Execute: Parallel delegation
5. Aggregate: Combine results
6. Verify: Integration consistency
7. Return: Complete feature

**Result:** Coordinated multi-agent execution

---

### Example 3: Documentation Update

**User Prompt:** "Update README with new deployment instructions"

**Orchestrator Logic:**
1. Analyze: Documentation only
2. Identify: Docs maintenance
3. Delegate: `knowledge-graph-docs-maintenance`
4. Execute: Single agent task
5. Return: Updated documentation

**Result:** Direct delegation to docs agent

---

## Prompt Templates for Agent Delegation

### Backend Agent Prompt Template

```
Using knowledge-graph-backend-go agent:

Task: {specific_backend_task}

Context:
- Related files: {file_paths}
- Architecture: Clean Architecture, DDD
- Patterns: Repository, Factory, Strategy
- Tech: Go 1.23, Gin, GORM, PostgreSQL, Redis

Requirements:
{detailed_requirements}

Constraints:
- Follow existing patterns in backend/
- Use DDD layers (domain → application → infrastructure → interfaces)
- Add tests for new functionality
- Update documentation if needed
```

### Frontend Agent Prompt Template

```
Using knowledge-graph-frontend-svelte agent:

Task: {specific_frontend_task}

Context:
- Related files: {file_paths}
- Tech: Svelte 5, TypeScript, Vite
- Patterns: Component composition, stores, services
- Theme: Cosmic/space theme

Requirements:
{detailed_requirements}

Constraints:
- Follow frontend/FRONTEND_PATTERNS.md
- Use existing CSS variables from global.css
- Add unit tests with Vitest
- Ensure accessibility (ARIA, keyboard nav)
```

### Docs Agent Prompt Template

```
Using knowledge-graph-docs-maintenance agent:

Task: {documentation_task}

Context:
- Files to update: {file_paths}
- Current docs structure: docs/, README.md, ADRs
- Style: Markdown, consistent formatting

Requirements:
{detailed_requirements}

Constraints:
- Maintain existing documentation structure
- Update all affected sections
- Ensure cross-references work
- Follow ADR format if creating new ADR
```

### Testing Agent Prompt Template

```
Using knowledge-graph-testing agent:

Task: {testing_task}

Context:
- Test types needed: {unit/integration/E2E/BDD}
- Languages: {Go/TypeScript/Python}
- Test framework: {Vitest/Playwright/Cucumber/Go test}
- Coverage target: {coverage_percentage}

Requirements:
{detailed_requirements}

Constraints:
- Follow existing test patterns
- Use table-driven tests for Go
- Use MSW mocks for frontend API tests
- Add tests to appropriate test files
- Ensure tests are stable and reliable
```

### Integration Agent Prompt Template

```
Using knowledge-graph-integration agent:

Task: {integration_task}

Context:
- Backend endpoints: {backend_paths}
- Frontend API clients: {frontend_paths}
- DTOs to sync: {type_definitions}
- Middleware: {CORS/auth/rate limiting}

Requirements:
{detailed_requirements}

Constraints:
- Ensure backend ↔ frontend type consistency
- Update API mapping documentation
- Add integration tests
- Verify CORS/auth configuration
```

---

## Execution Modes

### Mode 1: Simple Delegation (Single Agent)

**Trigger:** Task clearly belongs to one domain

**Flow:**
```
User Prompt → Analyze → Single Agent → Execute → Return
```

**Example:** "Fix failing test in notes.spec.ts"
→ Direct to `knowledge-graph-testing`

---

### Mode 2: Parallel Delegation (Multiple Agents)

**Trigger:** Task spans multiple domains, subtasks independent

**Flow:**
```
User Prompt → Split → Parallel Agent Execution → Aggregate → Return
```

**Example:** "Add new note type with UI and API"
→ Backend + Frontend + Integration + Tests (parallel)

---

### Mode 3: Sequential Delegation (Dependent Agents)

**Trigger:** Tasks have dependencies (backend must complete before frontend)

**Flow:**
```
User Prompt → Split → Agent 1 → Agent 2 → ... → Aggregate → Return
```

**Example:** "Refactor API and update frontend"
→ Backend first → Integration → Frontend → Tests

---

## Quality Checks

### Pre-Execution Validation

- [ ] Prompt is clear and actionable
- [ ] Required agents identified
- [ ] Dependencies understood
- [ ] Success criteria defined

### Post-Execution Verification

- [ ] All subtasks completed
- [ ] No conflicting changes
- [ ] Tests pass (run if possible)
- [ ] Documentation updated
- [ ] Code follows project patterns
- [ ] Integration points verified

---

## Error Handling

### Scenario 1: Agent Fails

```
Agent execution fails
↓
Log error details
↓
Retry with modified prompt (up to 3 times)
↓
If still fails: Report to user with partial results
```

### Scenario 2: Conflicting Changes

```
Agent A and Agent B make conflicting changes
↓
Detect conflict during aggregation
↓
Prioritize based on:
  1. Backend contracts (API)
  2. Frontend implementation
  3. Tests
  4. Documentation
↓
Resolve and notify user
```

### Scenario 3: Incomplete Task

```
Agent reports partial completion
↓
Identify missing parts
↓
Re-delegate or handle manually
↓
Update status
```

---

## Configuration

### Agent Selection Priority

1. **Explicit mention:** If user specifies agent, use it
2. **File path:** Match file paths to agent domains
3. **Keywords:** Match technical terms to agent expertise
4. **Default:** Use orchestrator analysis

### Parallel Execution Limits

- Maximum concurrent agents: 3
- Queue size: 5 tasks
- Timeout per agent: 10 minutes

### Retry Policy

- Max retries: 3
- Backoff: Exponential (1s, 2s, 4s)
- Fallback: Report partial results

---

## Monitoring and Logging

### Log Format

```
[ORCHESTRATOR] {timestamp} {level} {message}

Example:
[ORCHESTRATOR] 2026-05-22 10:30:15 INFO Received prompt: "Add API endpoint"
[ORCHESTRATOR] 2026-05-22 10:30:16 INFO Analyzed prompt, identified 3 subtasks
[ORCHESTRATOR] 2026-05-22 10:30:17 INFO Delegating to: backend, integration, tests
[ORCHESTRATOR] 2026-05-22 10:30:45 SUCCESS All agents completed successfully
[ORCHESTRATOR] 2026-05-22 10:30:46 INFO Aggregated results, returning to user
```

### Metrics Tracked

- Task completion time
- Agent success rate
- Conflict frequency
- User satisfaction (feedback)

---

## Best Practices

### For Orchestrator

1. **Be explicit:** Clearly document delegation decisions
2. **Verify:** Always check results before returning
3. **Document:** Log all delegation decisions
4. **Fail gracefully:** Return partial results if needed
5. **Learn:** Track patterns for better future routing

### For Users

1. **Be specific:** Provide clear, detailed prompts
2. **Mention agents:** If you know which agent to use, specify
3. **Provide context:** Include file paths and requirements
4. **Review output:** Check aggregated results
5. **Give feedback:** Help improve routing logic

---

## Known Limitations

1. **Cannot execute code:** Orchestrator delegates, doesn't execute
2. **No direct file access:** Agents must handle file operations
3. **Sequential dependencies:** Some tasks require order (backend before frontend)
4. **Conflict resolution:** Manual intervention may be needed for complex conflicts
5. **Test execution:** Orchestrator doesn't run tests, agents may suggest commands

---

## Future Enhancements

- [ ] Automatic test execution and verification
- [ ] Machine learning for better task routing
- [ ] Agent capability discovery
- [ ] Dynamic agent creation for new domains
- [ ] Performance optimization for parallel execution
- [ ] Conflict detection and auto-resolution
- [ ] User feedback loop for routing improvement

---

## References

- [Agents Guide](../../docs/AGENTS.md)
- [Architecture Documentation](../../docs/architecture/README.md)
- [Commands Reference](../../COMMANDS.md)
- [Frontend Patterns](../../frontend/FRONTEND_PATTERNS.md)
- [Backend Patterns](../../backend/BACKEND_PATTERNS.md)

---

**Last Updated:** 2026-05-22  
**Maintainer:** knowledge-graph-docs-maintenance  
**Version:** 1.0
