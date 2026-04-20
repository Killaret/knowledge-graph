# Architecture Roadmap: From Single-User to SaaS

## North Star Goal
Transform the current educational DDD project into a **production-grade SaaS reference architecture** demonstrating Senior System Architect competencies:
- ✅ Strict Multi-tenancy (Data Isolation)
- ✅ Security by Design (RLS + App-layer checks)
- ✅ Resilience & Scalability (Queues, CQRS, Async processing)
- ✅ Observability & Auditability

---

## Phase 1: Domain & Security Foundation (Weeks 1-2)
*Focus: Embedding tenancy into the core domain model.*

| ID | Task | Architectural Value | Risk if Ignored |
|----|------|---------------------|-----------------|
| **DOM-01** | **Inject `TenantID` into Aggregate Roots** | Ensures every business operation is scope-bound. Prevents accidental cross-tenant logic. | Data leakage at the source; complex refactoring later. |
| **DOM-02** | **Refactor Factories & Constructors** | Enforces that entities *cannot* be created without a tenant context. | Inconsistent state; "orphan" records. |
| **SEC-01** | **Define Tenant Context Propagation** | Establishes how `TenantID` flows from API Gateway → App Service → Domain. | Lost context leading to unauthorized access. |
| **DOC-01** | **Write ADR-015: Multi-tenancy Strategy** | Documents the "Why" and "How" of isolation (DB-per-tenant vs RLS). | Lack of clarity on security boundaries. |

## Phase 2: Data Isolation & Persistence (Weeks 3-5)
*Focus: Hardening the database layer against leaks.*

| ID | Task | Architectural Value | Risk if Ignored |
|----|------|---------------------|-----------------|
| **DB-01** | **Implement PostgreSQL Row Level Security (RLS)** | Database-level enforcement as a safety net, even if app logic fails. | Catastrophic data breach via SQL injection or bug. |
| **DB-02** | **Refactor Repositories to be Tenant-Aware** | Application-layer enforcement; repositories accept only domain objects, not raw IDs. | Bypassing domain logic; tight coupling to DB schema. |
| **MIG-01** | **Zero-Downtime Migration Strategy** | Safely add `tenant_id` columns to existing tables without locking. | Service outage during deployment; data corruption. |
| **AUD-01** | **Implement Audit Logging for Data Access** | Tracks *who* accessed *what* and *when*. Crucial for SOC2 compliance. | Inability to investigate incidents or prove compliance. |

## Phase 3: Resilience & Asynchronous Processing (Weeks 6-8)
*Focus: Handling load and failures gracefully.*

| ID | Task | Architectural Value | Risk if Ignored |
|----|------|---------------------|-----------------|
| **Q-01** | **Extract Heavy Tasks to Background Queues** | Decouples user-facing latency from heavy processing (embeddings, reports). | Poor UX; thread pool exhaustion under load. |
| **Q-02** | **Implement Idempotency Keys** | Guarantees exactly-once processing semantics for critical operations. | Duplicate charges; data inconsistency on retries. |
| **RES-01** | **Circuit Breakers for External Services** | Prevents cascade failures when 3rd party APIs (AI providers) slow down. | Total system outage due to external dependency. |

## Phase 4: Verification & Showcase (Weeks 9-10)
*Focus: Proving the architecture works under pressure.*

| ID | Task | Architectural Value | Risk if Ignored |
|----|------|---------------------|-----------------|
| **TEST-01** | **Chaos Engineering: Simulate Tenant Leak** | Intentionally break RLS/App checks to verify monitoring alerts trigger. | False sense of security; undetected vulnerabilities. |
| **LOAD-01** | **Multi-tenant Load Testing** | Verify isolation doesn't cause "noisy neighbor" performance issues. | SLA violations for premium tenants. |
| **DOC-02** | **"Architecture Decision Record" Portfolio** | Finalize ADRs explaining trade-offs made in each phase. | Inability to articulate decisions in an interview. |

---

## Definition of Done (DoD) for Architecture
A task is not complete until:
1.  Code is implemented and tested.
2.  **ADR is updated/created** explaining the trade-off.
3.  **Security implication** is documented (Threat Model update).
4.  **Rollback plan** exists for database changes.

## Current Status
- [ ] **Phase 1**: Not Started
- [ ] **Phase 2**: Not Started
- [ ] **Phase 3**: Not Started
- [ ] **Phase 4**: Not Started
