# UX-2: нормальная страница 500 ошибки

**Статус:** реализовано — на ревью у Claude Code (повторный раунд)  
**Создано:** 2026-09-11  
**Обновлено:** 2026-09-11  
**Связано:** `frontend/src/routes/+error.svelte`, `frontend/src/components/atoms/StateIllustration.svelte`, SvelteKit error handling, i18n.

## Проблема

- При 500 ошибке в приложении отображается лишь небольшой кусок экрана с текстом `internal error 500` (или `Internal Server Error 500`).
- Страница не растянута на весь браузер, не использует i18n, не соответствует дизайну Cosmic Cockpit.
- SvelteKit по умолчанию рендерит `+error.svelte` внутри layout; если такого файла нет, fallback выглядит как встроенная минимальная заглушка.

## Что сделано

1. Создан `frontend/src/routes/+error.svelte`:
   - full-viewport (`position: fixed; inset: 0; z-index: 900`),
   - Cosmic Cockpit фон,
   - i18n-ключи `error.500.*` и `error.unknown.*` (en/ru),
   - кнопки «Обновить страницу» и «На главную»,
   - dev/personal-режим — `import.meta.env.DEV` показывает stack trace; в production не показывает.
2. Добавлена иллюстрация для 5xx — `server-error` в `StateIllustration.svelte`: разъединённый удлинитель / вилка и розетка с искрой (по запросу владельца).
3. `+error.svelte` выбирает иллюстрацию по статусу: `404` → `404`, `5xx` → `server-error`, остальное → `error`.
4. Добавлен `frontend/src/routes/error-page.spec.ts` (Vitest): рендеринг, i18n, кнопки, иллюстрация.
5. В ходе работы найден и исправлен баг авторизации: `+layout.svelte` использовал `currentPath.startsWith("/")`, из-за чего **все** маршруты считались публичными. Вынесена функция `isPublicRoute` в `shared/utils/route-match.ts` и добавлены unit-тесты.
6. Добавлен Playwright-сценарий `frontend/tests/error-500-page.spec.ts` и test-only route `src/routes/test/500/+page.server.ts` (триггер `?trigger=500`) для проверки full-viewport рендера.
7. `ApiErrorDisplay.svelte` теперь использует иллюстрацию `server-error` для API-ошибок с кодом `INTERNAL_ERROR`.
8. Исправлены `golangci-lint` замечания в `note_handler.go` и `import_fetcher_test.go`.
9. Восстановлен гейт форматирования: `npm ci` + `npm run format` привели 6 Svelte-файлов в соответствие с lock-версией `prettier-plugin-svelte`.
10. Обновлена языковая политика: UI по умолчанию — English (`en`), документация — Russian.
11. Проверки:
   - `npm run check` — 0 ошибок, 0 warnings.
   - `npm run test:unit` — 111 файлов, 1009 тестов passed.
   - `npx vitest run src/routes/error-page.spec.ts src/shared/utils/route-match.test.ts` — passed.

## Аудит обработки ошибок по приложению

- **SvelteKit route errors (`+error.svelte`)**: обрабатывает ошибки в `load` / SSR (5xx, 404). Теперь полноэкранный и с i18n.
- **API errors в компонентах**: вместо `+error.svelte` страницы графа, импорта, поиска, заметок и главной страницы показывают **локальные** ошибки через `StateIllustration type="error"` / `ApiErrorDisplay`.
- **Graph / 3D graph**: `GraphPageShell.svelte` и `GraphCanvas` ловят ошибки загрузки и рисуют их внутри canvas-области, а не full-screen. Это сознательный UX: граф падает, но остальной UI остаётся.
- **Auth / network errors**: `+layout.svelte` редиректит на `/auth/login` для непубличных маршрутов. Исправлен баг `startsWith("/")`.
- **Backend 500**: `frontend/src/hooks.server.ts` проксирует `/api/v1/*` на backend и возвращает status/body. Если сам SvelteKit `load` поймает reject/throw, сработает `+error.svelte`. Чистые API-ошибки (JSON `{code, message}`) обрабатываются клиентским `ky`/API-слоем и показываются в виджетах.

## Решения по открытым вопросам

- **Язык по умолчанию**: приложение — `en`, документация — Russian. Зафиксировано в `.windsurfrules`, `MASTER_PROMPT.md`, `MASTER_PROMPT_RU.md`.
- **Graph / 3D при 500**: оставить локальную ошибку внутри canvas — full-screen `+error.svelte` остаётся для route/SSR/ `load`, а не для изолированных widget-ошибок.
- **Playwright-регрессия**: реализована `frontend/tests/error-500-page.spec.ts` + `src/routes/test/500/+page.server.ts`.
- **ApiErrorDisplay + `server-error`**: да, `INTERNAL_ERROR` теперь рисует `server-error`.

## Замечания для ревью Claude Code

- Проверить все коммиты в окне 2026-09-11, включая `aff53f2`, `460e913`, `5c69aa3`, `bcf7b59`, `0d2655e` и все последующие до слияния.

## Критерий приёмки

- [x] `+error.svelte` существует.
- [x] 500 ошибка занимает весь браузер, а не кусок экрана.
- [x] Текст и кнопки через i18n.
- [x] Есть иллюстрация для 5xx.
- [x] `npm run check` и `npm run test:unit` зелёные.
- [x] Unit-регрессионный тест на `+error.svelte`.
- [x] Playwright-регрессия на full-viewport.
- [x] Решение по языку (приложение `en`, документация `ru`).
- [x] `ApiErrorDisplay` для `INTERNAL_ERROR` использует `server-error`.
- [ ] CI полностью зелёный (следующий прогон после `0d2655e`).
