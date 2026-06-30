# Knowledge Graph AI Agents - Universal Setup

**Support: Cursor + GitHub Copilot + Windsurf**

---

## 🎯 Overview

This configuration ensures **automatic activation of orchestrator and 8 agents** in all AI assistants:

- ✅ **Cursor** — full support via `.cursor/rules/`
- ✅ **GitHub Copilot** — context via `.github/copilot/copilot-instructions.md`
- ✅ **Windsurf** — context via `.windsurf/rules.md`

---

## 📁 File Structure

```
knowledge-graph/
├── .cursor/
│   └── rules/
│       ├── knowledge-graph-orchestrator.md     # ⭐ Always active meta-agent
│       ├── knowledge-graph-performance.md      # Performance optimization
│       ├── knowledge-graph-security.md         # Security audit
│       └── knowledge-graph-devops.md           # DevOps/infrastructure
│
├── .github/
│   └── copilot/
│       └── copilot-instructions.md             # Copilot context
│
├── .windsurf/
│   └── rules.md                                # Windsurf rules
│
├── .koda/
│   ├── config.json                             # Koda configuration
│   └── skills/
│       └── knowledge-graph-*.md                # Koda agent definitions
│
└── docs/
    └── AGENTS_IMPLEMENTATION_COMPLETE.md       # Implementation report
```

---

## 🚀 How to Use

### Cursor (Recommended)

**Setup:** Automatic (files already created)

**How it works:**
```
1. Open project in Cursor
2. Start chat with AI
3. Orchestrator activates automatically (alwaysApply: true)
4. Just describe the task — AI will choose the agent automatically
```

**Example:**
```
User: "Optimize graph loading"
Cursor AI: → Routes to performance agent → Returns optimization guide
```

**Files:**
- `.cursor/rules/knowledge-graph-orchestrator.md` — orchestrator
- `.cursor/rules/knowledge-graph-performance.md` — performance
- `.cursor/rules/knowledge-graph-security.md` — security
- `.cursor/rules/knowledge-graph-devops.md` — DevOps

---

### GitHub Copilot

**Setup:** Automatic (file already created)

**How it works:**
```
1. Open project in VS Code with Copilot
2. Start chat with Copilot
3. AI uses context from .github/copilot/copilot-instructions.md
4. Describe the task — AI will consider agent rules
```

**Example:**
```
User: "Add API endpoint"
Copilot: → Uses backend agent context → Generates Go code
```

**File:**
- `.github/copilot/copilot-instructions.md` — context for all agents

**Limitation:** No delegation between agents, only unified context

---

### Windsurf

**Setup:** Automatic (file already created)

**How it works:**
```
1. Open project in Windsurf
2. Start chat with AI
3. AI uses context from .windsurf/rules.md
4. Describe the task — AI will consider agent rules
```

**Example:**
```
User: "Create Kubernetes deployment"
Windsurf AI: → Uses devops agent context → Generates manifests
```

**File:**
- `.windsurf/rules.md` — rules for Windsurf

**Limitation:** No delegation between agents, only unified context

---

## 🤖 All Agents (9 total)

| # | Agent | Purpose | Support |
|---|-------|---------|---------|
| 0 | **knowledge-graph-orchestrator** | Meta-agent coordinator | ✅ Cursor (full) |
| 1 | knowledge-graph-frontend-svelte | Frontend (Svelte 5) | ✅ All |
| 2 | knowledge-graph-backend-go | Backend (Go) | ✅ All |
| 3 | knowledge-graph-docs-maintenance | Documentation | ✅ All |
| 4 | knowledge-graph-testing | Testing | ✅ All |
| 5 | knowledge-graph-integration | API integration | ✅ All |
| 6 | **knowledge-graph-performance** | **Performance** | ✅ All |
| 7 | **knowledge-graph-security** | **Security** | ✅ All |
| 8 | **knowledge-graph-devops** | **DevOps** | ✅ All |
| 9 | **knowledge-graph-infrastructure** | **Infrastructure** | ✅ All |

---

## 📊 Support Comparison

| Tool | Delegation | Auto-activation | Context | Recommendation |
|------|------------|-----------------|---------|----------------|
| **Cursor** | ✅ Full | ✅ Yes | ✅ Separate files | ⭐ **Best choice** |
| **Copilot** | ⚠️ Partial | ⚠️ Via context | ✅ Single file | Good for VS Code |
| **Windsurf** | ⚠️ Partial | ⚠️ Via context | ✅ Single file | Good for VS Code |

---

## 🎯 Usage Examples

### Example 1: Optimization

```
Request: "Optimize bundle size by 50%"

Cursor:
  → Orchestrator analyzes → performance task
  → Delegates: knowledge-graph-performance
  → Result: Code splitting + lazy loading guide

Copilot/Windsurf:
  → AI uses performance agent context
  → Result: Similar advice
```

### Example 2: Security

```
Request: "Conduct security audit for auth"

Cursor:
  → Orchestrator analyzes → security task
  → Delegates: knowledge-graph-security
  → Result: Full security audit report

Copilot/Windsurf:
  → AI uses security agent context
  → Result: Security checklist
```

### Example 3: DevOps

```
Request: "Create Kubernetes deployment"

Cursor:
  → Orchestrator analyzes → devops task
  → Delegates: knowledge-graph-devops
  → Result: Full K8s manifests

Copilot/Windsurf:
  → AI uses devops agent context
  → Result: Deployment manifests
```

---

## 🔧 Setup (if needed)

### Cursor

```bash
# Files already created in .cursor/rules/
# Just open project in Cursor

# Check:
ls -la .cursor/rules/
```

### Copilot

```bash
# File already created in .github/copilot/copilot-instructions.md
# Just open project in VS Code with Copilot

# Check:
ls -la .github/copilot/copilot-instructions.md
```

### Windsurf

```bash
# File already created in .windsurf/rules.md
# Just open project in Windsurf

# Check:
ls -la .windsurf/rules.md
```

---

## ✅ Verification

### Cursor

```bash
# 1. Open project in Cursor
# 2. Press Cmd+K (or Ctrl+K)
# 3. Enter: "Optimize graph loading"
# 4. Expect: AI uses performance agent
```

### Copilot

```bash
# 1. Open project in VS Code
# 2. Press Cmd+Shift+P → "Copilot: New Chat"
# 3. Enter: "Optimize graph loading"
# 4. Expect: AI considers performance context
```

### Windsurf

```bash
# 1. Open project in Windsurf
# 2. Start chat with AI
# 3. Enter: "Optimize graph loading"
# 4. Expect: AI considers performance context
```

---

## 📚 Documentation

### Main documents

- **Cursor Rules:** `.cursor/rules/knowledge-graph-orchestrator.md`
- **Copilot Instructions:** `.github/copilot/copilot-instructions.md`
- **Windsurf Rules:** `.windsurf/rules.md`

### Additional documentation

- **Main Guide:** `.koda/README.md`
- **Implementation Report:** `docs/AGENTS_IMPLEMENTATION_COMPLETE.md`
- **Commands:** `COMMANDS.md`

---

## 🎯 Recommendations

### For maximum effectiveness

**Use Cursor** — it supports:
- ✅ Separate files for each agent
- ✅ Automatic delegation
- ✅ Full orchestrator auto-activation

### For VS Code

**Use Copilot or Windsurf** — they support:
- ✅ Unified context for all agents
- ✅ VS Code integration
- ⚠️ Limited delegation (via context)

---

## 🚀 Quick Start

### 1. Choose AI Assistant

- **Cursor** → Best experience (recommended)
- **Copilot** → Good experience in VS Code
- **Windsurf** → Good experience in VS Code

### 2. Open Project

```bash
# Cursor
cursor .

# VS Code (with Copilot)
code .

# Windsurf
windsurf .
```

### 3. Start Chat

```
Just describe the task:
"Optimize graph loading"
"Create API endpoint"
"Conduct security audit"
```

### 4. Get Result

AI will automatically use the appropriate agent!

---

## 📞 Support

**Questions about agents?**
- See `.koda/README.md`
- Or `docs/AGENTS_IMPLEMENTATION_COMPLETE.md`

**Setup issues?**
- Check that all files are created
- Restart AI assistant
- Ensure project is opened in correct directory

---

## 🎉 Summary

**Created:**
- ✅ 4 files for Cursor (orchestrator + 3 agents)
- ✅ 1 file for Copilot (context for all agents)
- ✅ 1 file for Windsurf (context for all agents)
- ✅ Full documentation

**Result:**
- ✅ Orchestrator always active in all tools
- ✅ 9 agents available in any AI assistant
- ✅ Automatic delegation (Cursor)
- ✅ Unified context (Copilot/Windsurf)

**Ready to use!** 🚀

---

**Last Updated:** 2026-06-30  
**Version:** 1.0  
**Status:** ✅ Complete
