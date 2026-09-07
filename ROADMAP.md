# Roadmap

Where Knowledge Graph is going. What the project is and how to run it — [README](README.md).
What shipped — [CHANGELOG](CHANGELOG.md). The detailed plan behind each line below —
[docs/BACKLOG.md](docs/BACKLOG.md). Untested hypotheses — [docs/IDEAS.md](docs/IDEAS.md).

**Stage:** alpha. Single-user and local by design at the moment. A September 2026 audit found
issues that block any multi-user deployment — they are listed under *Now* and tracked in
[docs/EXTERNAL_AUDIT_2026-09.md](docs/EXTERNAL_AUDIT_2026-09.md).

## Now

Correctness of the foundation, before new surface area.

| Work | Why it comes first |
|---|---|
| Close the audit blockers: data isolation under `SKIP_AUTH`, seeded credentials in migrations, internal auth headers at the gateway, private responses marked cacheable | Each one blocks multi-user deployment |
| Make verification tell the truth: honest regression exit codes, a real 3D readiness signal, visual baselines that capture the scene | A green run that proves nothing is worse than a red one |
| Enforce coverage thresholds in CI and make the orphaned BDD scenarios executable | Declared thresholds that never run are decoration |

## Next

| Work | Notes |
|---|---|
| Multilingual embeddings | One vector space for Russian and English notes; the current model is English-only |
| Keyword normalization | Lemmatization at extraction; today word forms are compared as raw strings |
| Graph clustering | Communities over a hybrid metric — semantics as the base, existing links reinforcing it |
| Link types in the UI | Native picker and documentation for link semantics |
| Delta-update flicker | A race in the preload path makes the graph blink on refresh |

## Later

Honeycomb and orbital 3D layouts, zoomable navigation into a cluster, note archive and hygiene,
Obsidian import, PWA quick capture, periodic notes, sharing between users, external integrations.
Each is described in [docs/BACKLOG.md](docs/BACKLOG.md).

## Exploring

Guardians of a cluster, factory-line visualization, leaderboards and public universes, a social
layer, graph motion driven by graph metrics. These are hypotheses, not commitments —
[docs/IDEAS.md](docs/IDEAS.md).

---

Planning currently lives in these files rather than in an issue tracker, which is why status has
to be read rather than queried. Moving the working plan to issues is itself on the list.
