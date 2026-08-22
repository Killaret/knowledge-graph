# Celestial Body Semantics

This document defines what each celestial body type means in the knowledge graph and when users (and the system) should use it.

## Core idea

A note is not just a node — it plays a role inside a topic cluster. The celestial body type encodes **granularity**, **lifecycle stage**, and **centrality**.

## Hierarchy

```
Galaxy          # broad domain
  └── Star      # central topic / pillar
        ├── Planet      # major sub-topic
        │     ├── Moon        # detail / aspect
        │     ├── Satellite   # utility / checklist / config
        │     ├── Comet       # temporary / event-bound note
        │     └── Asteroid    # raw fragment / quote / TODO
        ├── Star            # neighboring topic in the same domain
        └── Black Hole      # hard problem / open question
```

## Type reference

| Type              | Level / role  | When to use                                                                                                         | Example                                  |
| ----------------- | ------------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| **Galaxy** 🌀     | Domain        | Broad context that groups several central topics. Use sparingly.                                                    | Media, Software Architecture             |
| **Star** ⭐       | Central topic | Root idea / pillar that anchors a cluster of notes.                                                                 | Culture, Anime, Backend                  |
| **Planet** 🪐     | Sub-topic     | Major section orbiting a star.                                                                                      | Manga, Manhwa, Authentication            |
| **Moon** 🌙       | Detail        | Specific aspect tied to a planet. Usually assigned automatically.                                                   | Manga genres, Publishers                 |
| **Satellite** 🛰️  | Utility       | Checklists, templates, configs, snippets attached to a topic.                                                       | Deploy checklist, .env.example           |
| **Comet** ☄️      | Event         | Temporary / time-bound notes: sprints, meetings, releases.                                                          | Sprint 42 retrospective, Anime Expo 2026 |
| **Asteroid** 🌑   | Fragment      | Raw material: quotes, bookmarks, quick thoughts, TODOs.                                                             | Interesting quote, TODO: check this      |
| **Nebula** 💫     | Draft         | Fuzzy, forming idea without clear borders.                                                                          | Future product ideas, Unsorted concepts  |
| **Black Hole** ⚫ | Problem       | Large unclear task, open question, or fundamental bug. **Not a deletion target.**                                   | How to scale ingestion?, Unsolved bug    |
| **Debris** 🌌     | Archive       | Outdated or rejected material kept for history. Assigned automatically when a note is archived via the Singularity. | API v1 (deprecated), Rejected proposal   |
| **Dust** 🌫️       | Inbox         | Quick-captured, unprocessed note waiting for triage.                                                                | Late-night idea, Quick bookmark          |
| **Technical** ⚙️  | Meta          | System notes used by the app itself.                                                                                | Tagging rules, CI/CD notes               |
| **Unknown** ❓    | Fallback      | Notes without a known classification.                                                                               | —                                        |
| **Anomalies**     | State markers | Special visual markers for unusual graph states. Not user notes.                                                    | Reality Rift, Void Whisper, …            |

## UI visibility

Types exposed in the note creation / type selector:

- `star`, `planet`, `comet`, `asteroid`, `nebula`, `galaxy`, `satellite`, `blackhole`, `dust`

Hidden / automatic types:

- `moon` — assigned automatically when a small note is created from a planet.
- `debris` — assigned automatically when a note is archived via the Singularity.
- `technical` — reserved for system-generated notes.
- `unknown` — fallback.
- anomaly types — reserved for graph-state visual effects.

## Singularity — archive / delete zone

The **Singularity** is a UI-only drop zone, not a celestial body type. It appears while dragging a note and provides a clear, separate gesture for archiving or deleting:

- Drag a note into the Singularity → a confirmation dialog opens.
- If confirmed, the note is removed from the graph and (depending on backend behavior) archived as `debris` or deleted.
- The central **Black Hole** visual element remains purely decorative and metaphorical; dropping a note on it does nothing destructive.

This separation prevents the `Black Hole` type from being confused with a deletion action.

## Creating child notes

Users can create a child note from any visible note in two ways:

- **Graph node context menu:** right-click a node and choose _Create child note_.
- **Note details panel:** click the _Create child note_ button in the right cockpit panel.

When the child note modal opens, the initial type is suggested by `CelestialBody.getChildSuggestion(parentType)`:

| Parent type | Suggested child type |
| ----------- | -------------------- |
| Galaxy      | Star                 |
| Black Hole  | Star                 |
| Star        | Planet               |
| Planet      | Satellite            |
| Nebula      | Star                 |
| Satellite   | Asteroid             |
| Comet       | Asteroid             |
| Asteroid    | Asteroid             |
| Dust        | Planet               |
| default     | Planet               |

`Moon` is intentionally excluded from the suggestions because it is an auto-assigned detail type.

After the new note is saved, a `parent` link is created from the selected parent to the child with `LinkType.PARENT.defaultWeight` (0.9). The user can still change the type before saving.

## Link type conventions

Use the existing link types to express relationships between celestial bodies:

| Source → Target           | Recommended link type    |
| ------------------------- | ------------------------ |
| Galaxy → Star             | `parent` / `child`       |
| Star → Planet             | `parent` / `child`       |
| Planet → Moon             | `parent` / `child`       |
| Planet → Satellite        | `dependency`             |
| Planet → Comet / Asteroid | `reference` or `related` |
| Planet → Black Hole       | `dependency`             |
| Star ↔ Star (same domain) | `related`                |

## Color variation and customization

Each celestial body type has a deterministic palette inspired by real-world cosmic objects (stellar spectral classes, planet atmospheres, comet ices, etc.):

- Colors are generated per node by `getVariation` in `frontend/src/shared/utils/variation.ts`.
- `frontend/src/shared/lib/graph/color-schemes.ts` defines the palettes.
- If a `GraphNode`/`SimulationNode` provides a `color` (and optional `glowColor`), it is used as-is and the palette is ignored.
- Generated colors are slightly hue- and lightness-shifted for visual diversity while staying inside the type's palette.
- Renderers use `variation.color`, `variation.glowColor`, and `variation.strokeColor` instead of hardcoded values.

## Design notes

- A user should be able to read the type description in the type selector and pick the right one without guessing.
- `blackhole` is intentionally exposed: users often need a place to park hard, ill-defined problems.
- `moon` stays hidden because manual classification at that granularity creates noise; the system can promote a note to planet later.
- Descriptions and examples live in `frontend/src/shared/utils/i18n.ts` and are rendered by `TypeSelector` and `CockpitTypeFilter`.
