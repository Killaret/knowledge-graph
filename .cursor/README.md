# Cursor Agent Configuration

This directory contains rules for Cursor AI.

## Important
- Cursor must read all files in `.cursor/rules/`
- Orchestrator must be active first
- Each agent has its own role and command set
- Do not delete this directory and files

## Structure
- `.cursor/rules/knowledge-graph-orchestrator.md`
- `.cursor/rules/knowledge-graph-backend-go.md`
- `.cursor/rules/knowledge-graph-frontend-svelte.md`
- `.cursor/rules/knowledge-graph-integration.md`
- `.cursor/rules/knowledge-graph-infrastructure.md`
- `.cursor/rules/knowledge-graph-devops.md`
- `.cursor/rules/knowledge-graph-performance.md`
- `.cursor/rules/knowledge-graph-security.md`
- `.cursor/rules/knowledge-graph-testing.md`

## Usage
Cursor AI must read these files before executing tasks.

## Language Policy

All user-facing content MUST be in English:
- Note titles and content
- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips
- Code comments in public documentation
- Commit messages

**Exceptions:** Internal code comments (brief explanations in any language OK)
