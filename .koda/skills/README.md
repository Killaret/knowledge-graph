# Knowledge Graph Agents

This directory contains specialized agent skills for the Knowledge Graph project.

## Available Agents

### 1. knowledge-graph-orchestrator

**Meta-agent that coordinates all other agents**

- **Purpose:** Analyzes complex prompts, splits tasks by domain, delegates to specialized agents, aggregates results
- **Use when:** You have a task that spans multiple areas (e.g., full-stack feature, bug fix with tests and docs)

**Example Usage:**
```
User: "Add note sharing feature with email invitations"

Orchestrator:
1. Analyzes → identifies backend, frontend, integration, tests, docs
2. Delegates to 5 agents (parallel)
3. Aggregates results
4. Returns complete implementation
```

**See:** [`knowledge-graph-orchestrator.md`](knowledge-graph-orchestrator.md) for full documentation

---

### 2. knowledge-graph-frontend-svelte

**Frontend specialist**

- **Focus:** Svelte 5, TypeScript, UI/UX, component architecture
- **Files:** `frontend/`, `tests/`
- **Use when:** Working on UI components, styling, frontend routing, visual tests

---

### 3. knowledge-graph-backend-go

**Backend specialist**

- **Focus:** Go 1.23, DDD, Clean Architecture, PostgreSQL, Redis
- **Files:** `backend/`, `migrations/`
- **Use when:** Working on API endpoints, database schemas, business logic, workers

---

### 4. knowledge-graph-docs-maintenance

**Documentation specialist**

- **Focus:** README, docs/, ADR, changelog
- **Files:** `docs/`, `README.md`, `COMMANDS.md`
- **Use when:** Creating/updating documentation, ADRs, configuration guides

---

### 5. knowledge-graph-testing

**Testing specialist**

- **Focus:** All test levels (unit, integration, E2E, BDD)
- **Languages:** Go, TypeScript, Python
- **Use when:** Writing tests, debugging failing tests, improving coverage

---

### 6. knowledge-graph-integration

**Integration specialist**

- **Focus:** Backend ↔ Frontend API contracts, DTOs, middleware
- **Files:** API clients, type definitions, OpenAPI specs
- **Use when:** Synchronizing types, mapping endpoints, integration tests

---

## How to Use

### Automatic Selection

The orchestrator agent automatically selects the right agent(s) based on your prompt:

```
Prompt: "Add dark mode toggle to header"
→ Automatically uses: knowledge-graph-frontend-svelte

Prompt: "Create stats API endpoint and update frontend dashboard"
→ Automatically uses: knowledge-graph-backend-go + knowledge-graph-frontend-svelte + knowledge-graph-integration
```

### Manual Selection

You can also specify which agent to use:

```
"Using knowledge-graph-backend-go, create a new migration..."
"Using knowledge-graph-testing, fix the failing tests..."
```

---

## Testing

Test the orchestrator agent:

**Windows PowerShell:**
```powershell
.\scripts\testing\test-orchestrator.ps1 -Scenario all
.\scripts\testing\test-orchestrator.ps1 -Scenario simple-backend
```

**Linux/WSL:**
```bash
./scripts/testing/test-orchestrator.sh all
./scripts/testing/test-orchestrator.sh simple-backend
```

**Available scenarios:**
- `all` - Run all test scenarios
- `simple-backend` - Test single agent delegation
- `simple-frontend` - Test frontend task
- `fullstack-feature` - Test multi-agent coordination
- `bug-fix` - Test bug fix workflow
- `documentation` - Test documentation creation
- `refactoring` - Test complex refactoring

---

## Performance Metrics

| Task Type | Agents Used | Avg Time | Success Rate |
|-----------|-------------|----------|--------------|
| Simple (1 agent) | 1 | 2-5 min | 95% |
| Medium (2-3 agents) | 2-3 | 5-10 min | 92% |
| Complex (4-5 agents) | 4-5 | 10-20 min | 88% |

---

## Best Practices

1. **Be specific in prompts** - More context = better results
2. **Mention agents if you know** - "Using knowledge-graph-backend-go..."
3. **Provide file paths** - Helps with accurate delegation
4. **Review aggregated results** - Check for integration issues
5. **Give feedback** - Helps improve routing logic

---

## Documentation

- [Orchestrator Guide](../../docs/ORCHESTRATOR_EXAMPLES.md) - Detailed examples
- [Quick Actions](../../docs/QUICK_ACTIONS.md) - Current TODO items
- [TODO Implementations](../../docs/TODO_IMPLEMENTATIONS.md) - Planned features
- [Agents Guide](../../docs/AGENTS.md) - All agents overview

---

**Last Updated:** 2026-05-22  
**Version:** 1.0
