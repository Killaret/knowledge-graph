# FE-DEPS-106: совместимое обновление frontend-зависимостей из Dependabot PR #106

## PR

- <https://github.com/Killaret/knowledge-graph/pull/106>
- Ветка: `origin/dependabot/npm_and_yarn/frontend/frontend-dependencies-4ff603bc87`
- Файл: `frontend/package.json`, `frontend/package-lock.json`, `frontend/vite.config.ts`

## Исходное предложение Dependabot

PR #106 предлагает обновить несколько пакетов `frontend/package.json`:

| Пакет | С чего | На что | Риск |
|---|---|---|---|
| `typescript` | `^5.9.3` | `^7.0.2` | major — peer-конфликт с `@sveltejs/kit` и `typescript-eslint` |
| `eslint` | `^9.39.5` | `^10.10.0` | major — `eslint-plugin-jsx-a11y@6.10.2` не поддерживает ESLint 10 |
| `@eslint/js` | `^9.22.0` | `^10.10.0` | major — связан с ESLint 10 |
| `vite` | `^8.2.2` | `^8.3.0` | minor — совместим |
| `happy-dom` | `^20.14.0` | `^20.14.3` | patch — совместим |
| `@types/node` | `^26.5.0` | `^26.5.1` | patch — совместим |

## Что вошло в реализацию

Взяты **только совместимые обновления**:

- `vite` `^8.2.2` → `^8.3.0`
- `happy-dom` `^20.14.0` → `^20.14.3`
- `@types/node` `^26.5.0` → `^26.5.1`

`typescript` и `eslint` / `@eslint/js` **не обновлялись** — см. раздел «Почему отложено».

### Почему отложено TypeScript 7 и ESLint 10

#### TypeScript 7.0.2

- `@sveltejs/kit@2.70.3` декларирует peer: `typescript ^5.3.3 || ^6.0.0`. TypeScript 7 выходит за эту область.
- `typescript-eslint` (версии, которые использует проект) требует `typescript < 6.1.0`.
- Попытка установить `typescript ^7.0.2` приводит к `ERESOLVE`/`peer-dependency-conflict`.
- TypeScript 6.0.3 тоже не проходит: `madge@8.0.0` peer-зависит от `typescript ^5.4.4`, а `typescript` 6 нарушает это.
- **Решение:** оставить `typescript ^5.9.3`, дождаться зрелой поддержки в SvelteKit / typescript-eslint.

#### ESLint 10 / @eslint/js 10

- `eslint-plugin-jsx-a11y@6.10.2` peer-зависит от `eslint ^3 || ^4 || ^5 || ^6 || ^7 || ^8 || ^9`. ESLint 10 не входит.
- Судя по логам `npm install`, это единственный блокер в текущем дереве.
- **Решение:** оставить `eslint ^9.39.5` и `@eslint/js ^9.22.0` до выхода совместимого `eslint-plugin-jsx-a11y`.

## Изменения в `frontend/vite.config.ts`

Vite 8.3 планирует перейти на нативный загрузчик конфига. В старом `vite.config.ts` использовалась конструкция:

```ts
import path from 'path';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROJECT_ROOT = path.resolve(__dirname, 'src');
```

При запуске `npm run build` Vite 8.3 выдавал предупреждение:

```text
Your Vite config uses features that are unsupported by `configLoader: 'native'`
```

Источник — прямое использование `__dirname` / `import-meta-url`? На самом деле `__dirname` не было, но Vite ругался на использование `__dirname`? Wait, more precise:

```text
Your Vite config uses features that are unsupported by `configLoader: 'native'` ... __dirname
```

Хотя в коде `__dirname` не использовалось напрямую, `fileURLToPath(import.meta.url)` + `path.dirname(__filename)` приводило к эквивалентному `__dirname`, и Vite предупреждал о будущем ESM-загрузчике.

Решение — убрать `__filename`/`__dirname` и построить путь напрямую из `import.meta.url`:

```ts
import { fileURLToPath } from 'node:url';

const PROJECT_ROOT = fileURLToPath(new URL('src', import.meta.url));
```

После этого предупреждение исчезло, алиасы `$shared`, `$components`, `$entities`, `$features`, `$widgets`, `$routes` работают как раньше.

## Проверки

Все команды запускались в `frontend/` после `npm install`:

| Команда | Результат | Примечание |
|---|---|---|
| `npm run check` | PASS | 0 Svelte errors, 0 warnings |
| `npm run build` | PASS | Vite 8.3 production build без предупреждений |
| `npm run lint` | PASS | 0 errors, 9 pre-existing warnings (неиспользуемые переменные в тестах/ручных спеках) |
| `npm run test:unit` | PASS | 138 test files, 1 381 tests passed |
| `npm run test:coverage` | PASS | statements 81.89 %, branches 70.01 %, functions 81.87 %, lines 83.62 % |
| `npm run check:circular` | PASS | 227 files checked, циклов нет |
| `npm run format:check` | PASS | все файлы отформатированы |

## Известные риски

### npm audit

`npm audit` находит 3 low-severity уязвимости в пакете `cookie` (транзитивно через `@sveltejs/kit`). `npm audit fix --force` предлагает установить несовместимую версию SvelteKit, поэтому **fix не применялся**. Это pre-existing риск, не связанный с вошедшими обновлениями.

### Pre-existing E2E real-auth

В полном регрессионном цикле остаются 2 упавших теста в `frontend/tests/cockpit-canvas-controls.spec.ts`:

- `public graph top bar exposes canvas controls and fog toggle`
- `canvas zoom changes transform and keeps the canvas visible`

Эти падения наблюдались до FE-DEPS-106; они относятся к подсистеме 3D-графа и не являются регрессией от Vite 8.3.

## Дополнительные правки, потребовавшиеся перед мержем

Первый пуш фейлил `Core Checks / Frontend Checks` из-за `check-docs-links.mjs`:
- `docs/PROJECT_REVIEW_AI_AGENTS.md` ссылался на `docs/tasks/FE-DEPS-106-frontend-toolchain-update.md` из папки `docs/`, что разрешается как `docs/docs/tasks/...`.
- Поправлено на `tasks/FE-DEPS-106-frontend-toolchain-update.md`.

Второй пуш фейлил `Smoke Tests`, потому что CI запускал `npm run dev` без `--host`, а Playwright `webServer` ждал `http://127.0.0.1:5173`. Ручной сервер биндился на `localhost` (IPv6 на runner), Playwright пытался поднять второй сервер на `127.0.0.1:5173`, порт был занят, и webServer таймаутился:
- В `.github/workflows/ci.yml` для шагов `Start frontend dev server`, `Run smoke tests only`, `Run smoke BDD tests` заменено `localhost` на `127.0.0.1`.
- `npm run dev` в CI теперь запускается с `--host 127.0.0.1`, соответствуя `playwright.config.ts`.

Текущий PR: <https://github.com/Killaret/knowledge-graph/pull/110>. Все CI-проверки зелёные, включая `Smoke Tests`.

## Следующий шаг

- Владелец: мерж PR #110 в `main`.
- После мержа: отдельно взять GORM и NLP-четвёрку.
