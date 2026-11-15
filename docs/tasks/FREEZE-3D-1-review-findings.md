# FREEZE-3D-1 — разбор ревью

## Ревью — Claude Code, 2026-11-15 (ревью №8)

Работа: Devin, `12bd8a1`. **Вердикт: отклонено, один блокер.** Заморозка сделана как поставлено, все три заявленные
мутации красные. Но нормализация вида в геттере `graphStore.currentView` уронила существующий тест домашней
страницы: фронтовые юнит-тесты красные с 01.10.

### Блокер: `home-page.svelte.test.ts` — «switches views and toggles layout provider»

`frontend/src/features/home-page/home-page.svelte.test.ts:202–203`: тест пишет `handleToggleView("3d")` и
ждёт `graphStore.currentView === "3d"`. После `12bd8a1` геттер при выключенном 3D отвечает `"graph"`:

```
npx vitest run src/features/home-page/home-page.svelte.test.ts
× switches views and toggles layout provider
AssertionError: expected 'graph' to be '3d'
```

Повторяется на `origin/ai-agents` 2026-11-15. Его же видят прогоны `check-all` других ревью: LINKS-2 и
PANEL-LINKS-1 записали его как «FAIL вне задачи». Отчёт исполнителя этот файл не упоминает: Devin гонял
только свои спеки, а не весь набор фронта.

Что нужно: тест приводится к новому контракту. При выключенном 3D запись `"3d"` читается как `"graph"`. При
`isGraph3DEnabled() → true` (заглушка, как в `graph.svelte.test.ts`) читается `"3d"`. Перед сдачей —
`npm run test:unit` целиком, а не отдельные файлы.

### Критерии

1. Кнопка «3D» по настройке — проверено. `GraphTopBar.spec.ts`; мутация «кнопка всегда» (`true ? [...]`)
   красная.
2. `/graph/3d` при `false` → 2D, запомненный `3d` → `graph` — проверено. `routes/graph/3d/page.spec.ts`;
   мутация «без редиректа» красная. `graph.svelte.test.ts`; мутация «геттер без нормализации» красная.
3. Тесты 3D выполняются и зелёные — проверено: `features/graph-3d` и `widgets/graph-3d-viewer` 32/32.
4. Снимок верхней панели без «3D» — проверено: `docs/agents/screenshots/freeze-3d-1/top-bar.png`, в
   переключателе «◯» и «☰».

### Принято без замечаний

- Настройка — в `config/frontend.json`, собирается `build-config`, описана в `CONFIGURATION_EN/RU` и
  `CONFIG_REGISTRY.md`.
- Редирект с `replaceState` до `initAuth`: в истории не остаётся шага «назад в 3D».
- Код и тесты 3D не тронуты, `GRAPH3D.md` помечен.
