# Knowledge Graph Frontend

SvelteKit frontend for the Knowledge Graph application: 2D/3D graph visualization, note editing, achievements, drafts.

**Stack:** Svelte 5 (runes only), TypeScript strict, SvelteKit, ky, D3-force, Three.js. See `../.windsurfrules` for the normative architecture rules (FSD + Atomic Design layers, `$shared/*` aliases).

## Developing

```sh
npm install
npm run dev        # Vite dev server on :5173; proxies /api/v1 → :9000, /graph-service/api → :9091
```

In Docker the frontend is served by the `kg-frontend` container and reached through nginx (`:18081` on the dev stack); there is no published host port.

## Testing

```sh
npm run test:unit      # Vitest
npm run test           # Playwright E2E against the isolated test stack
npm run test:bdd       # Cucumber BDD
npm run check          # svelte-check / types
```

E2E/BDD require the isolated test stack (`scripts/testing/start-test.ps1`); see `../docs/operations/TESTING.md`.

## Layout

```
src/routes/      pages (SvelteKit / FSD pages layer)
src/widgets/     self-contained sections (cosmic-cockpit, graph-canvas, …)
src/features/    user scenarios (graph-ui, graph-3d, …)
src/entities/    domain entities (note, link, user, achievement, …)
src/components/  atoms / molecules / organisms
src/shared/      api clients, rune stores, services, utils, config
```
