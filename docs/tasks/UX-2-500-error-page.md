# UX-2: нормальная страница 500 ошибки

**Статус:** реализовано — на ревью у Claude Code  
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
6. Проверки:
   - `npm run check` — 0 ошибок, 0 warnings.
   - `npm run test:unit` — 111 файлов, 1009 тестов passed.
   - `npx vitest run src/routes/error-page.spec.ts src/shared/utils/route-match.test.ts` — passed.

## Аудит обработки ошибок по приложению

- **SvelteKit route errors (`+error.svelte`)**: обрабатывает ошибки в `load` / SSR (5xx, 404). Теперь полноэкранный и с i18n.
- **API errors в компонентах**: вместо `+error.svelte` страницы графа, импорта, поиска, заметок и главной страницы показывают **локальные** ошибки через `StateIllustration type="error"` / `ApiErrorDisplay`.
- **Graph / 3D graph**: `GraphPageShell.svelte` и `GraphCanvas` ловят ошибки загрузки и рисуют их внутри canvas-области, а не full-screen. Это сознательный UX: граф падает, но остальной UI остаётся.
- **Auth / network errors**: `+layout.svelte` редиректит на `/auth/login` для непубличных маршрутов. Исправлен баг `startsWith("/")`.
- **Backend 500**: `frontend/src/hooks.server.ts` проксирует `/api/v1/*` на backend и возвращает status/body. Если сам SvelteKit `load` поймает reject/throw, сработает `+error.svelte`. Чистые API-ошибки (JSON `{code, message}`) обрабатываются клиентским `ky`/API-слоем и показываются в виджетах.

## Открытые вопросы (для ревью Claude Code)

- **Язык по умолчанию**: в постановке написано «по-русски по умолчанию», но проектная норма (`.windsurfrules`, `knowledge-graph.config.json`, `getCurrentLocale()`) — `en`. Нужно решение: либо изменить глобальный дефолт, либо 500-страница пока следует текущему дефолту (`en`).
- **Graph / 3D при 500**: оставить локальную ошибку внутри canvas или сделать общий full-screen fallback?
- **Playwright-регрессия**: у нас пока unit-тест. Нужен ли Playwright-сценарий, который доказывает full-viewport на живом браузере? Для этого потребуется стимулировать 500 через `load` (например, временный `+page.ts` в test-only route) или через mock.
- **ApiErrorDisplay + `server-error`**: стоит ли использовать новую иллюстрацию `server-error` для in-page `INTERNAL_ERROR` (`ApiErrorDisplay`)?

## Критерий приёмки

- [x] `+error.svelte` существует.
- [x] 500 ошибка занимает весь браузер, а не кусок экрана.
- [x] Текст и кнопки через i18n.
- [x] Есть иллюстрация для 5xx.
- [x] `npm run check` и `npm run test:unit` зелёные.
- [x] Unit-регрессионный тест на `+error.svelte`.
- [ ] Playwright-регрессия на full-viewport (на усмотрение ревью).
- [ ] Решение про русский по умолчанию.
