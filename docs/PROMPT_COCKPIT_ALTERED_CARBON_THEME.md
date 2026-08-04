# Prompt: Altered Carbon theme for Cosmic Cockpit (revised)

> Revised version. See review notes in `docs/MANUAL_TEST_FEEDBACK.md` (Cockpit
> §6) for why the original prompt was reworked. Do not push without
> confirmation — implement on a branch/worktree and stop for review.

---

Ты — AI-агент в проекте Knowledge Graph (ветка `feature/cosmic-cockpit` или
аналогичная рабочая ветка от `ai-agents`).

## Задача

Усилить визуальный стиль **Cosmic Cockpit** в духе "Altered Carbon"
(неоновый бирюзовый + фиолетовый/магента, голографический текст, тёмный
металлик) **только** внутри компонентов кабины, переиспользуя существующую
тему **Allotropic Carbon**, а не заводя вторую параллельную палитру.
Публичный граф и остальной интерфейс не трогать.

## Важно: реальные файлы (не придумывай новые имена)

- Рамка: `frontend/src/features/cosmic-ui/CockpitFrame.svelte` — **уже**
  реализует grid + "звёзды" + угловые болты + градиентную рамку
  cyan→dark→magenta через хардкод rgba. Не переписывать с нуля — перевести
  на переменные (шаг 2).
- Панели: `frontend/src/widgets/cosmic-cockpit/CockpitPanel.svelte`,
  `CockpitTopPanel.svelte`, `CockpitLeftPanel.svelte`,
  `CockpitRightPanel.svelte`, `CockpitBottomPanel.svelte`,
  `CockpitNoteDetails.svelte`.
- HUD: `frontend/src/widgets/cosmic-cockpit/CockpitHUD.svelte`.
- First-person exit: `CockpitFirstPersonButton.svelte`.
- Layout root: `frontend/src/widgets/cosmic-cockpit/CosmicCockpitLayout.svelte`
  (already renders `.cosmic-cockpit` root class — reuse it, do not add a
  second wrapper).
- **НЕ трогать** `frontend/src/widgets/notification/ToastNotification.svelte`
  — это общий компонент, используется вне кабины. Если тосты внутри кабины
  должны выглядеть иначе, заводить это отдельной задачей (условный рендер
  варианта), а не в этом промте.

## Шаг 1 — Алиасы переменных, а не новая палитра

В `frontend/src/shared/styles/global.css` внутри `:root` уже есть:
`--color-info` (#2dd4bf, бирюзовый), `--color-glow-purple`,
`--carbon-gradient-primary` (`linear-gradient(135deg, #22d3ee 0%, #8b5cf6 100%)`),
`--carbon-glow-primary`, `--carbon-glow-cyan`, `--carbon-border-glow` и т.д.

Добавь **только алиасы**, ограниченные скоупом кабины (не новые цвета):

```css
/* Cockpit-scoped aliases into the existing Allotropic Carbon palette.
   Do NOT introduce new hues — the cockpit uses the same theme as the rest
   of the app, just applied to different components. */
.cosmic-cockpit {
  --cockpit-accent: var(--color-info);
  --cockpit-accent-2: var(--color-glow-purple);
  --cockpit-panel-bg: var(--carbon-gradient-surface);
  --cockpit-panel-border: 1px solid var(--carbon-border-glow);
  --cockpit-glow: var(--carbon-glow-primary);
  --cockpit-gradient-text: var(--carbon-gradient-primary);
}
```

## Шаг 2 — Анимированный градиентный текст (не статичные 2 цвета)

Вместо статичного `background: linear-gradient(...); background-clip: text;`
сделай **один общий CSS-класс с двигающимся градиентом**, чтобы не плодить
копии в каждом компоненте:

```css
/* shared/styles/global.css or a new shared/styles/cockpit-text.css,
   imported once */
.cockpit-gradient-text {
  background: var(--cockpit-gradient-text, var(--carbon-gradient-primary));
  background-size: 200% auto;
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: cockpit-gradient-shift 6s ease-in-out infinite;
  animation-delay: var(--cockpit-text-delay, 0s);
}

@keyframes cockpit-gradient-shift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

@media (prefers-reduced-motion: reduce) {
  .cockpit-gradient-text {
    animation: none;
    background-position: 0% 50%;
  }
}
```

**Обязательно:**
- Каждый экземпляр, где применяешь класс, задаёт свой
  `style="--cockpit-text-delay: {N}s"` (разное `N` на разных заголовках —
  0, -1.3, -2.7 и т.п.), чтобы заголовки не мигали в такт друг с другом
  (та же ошибка синхронности, которую мы недавно чинили на графе).
- Проверь контраст в **каждом** кадре анимации (не только на 0%/100%), не
  только на конечных стопах градиента — минимум 4.5:1 против фона панели.
- Если в проекте есть проверка возможностей устройства
  (`frontend/src/shared/utils/deviceCapabilities.ts` или аналог) — на слабых
  устройствах анимация отключается тем же способом, что и для других
  эффектов кабины.

### Где применять `.cockpit-gradient-text`

- **Да, анимированный градиент:** заголовки панелей (`panel-title` в
  `CockpitPanel.svelte`), текст кнопки First Person, заголовки секций в
  `CockpitLeftPanel`/`CockpitNoteDetails` — это декоративные элементы.
- **Нет / только статичный градиент без анимации:** числовые метрики в
  `CockpitHUD.svelte` (FPS, notes, links, health%, sync-статус) — это
  read-at-a-glance данные, анимация мешает считыванию. Если хочется акцента
  — используй `var(--cockpit-accent)` сплошным цветом или статичный
  градиент (тот же класс, но с `animation: none` через модификатор
  `.cockpit-gradient-text--static`).

## Шаг 3 — Рамка и панели на переменные

- `CockpitFrame.svelte`: замени хардкод rgba на `var(--cockpit-accent)` /
  `var(--cockpit-accent-2)` / `var(--cockpit-glow)` там, где это буквально
  тот же бирюзовый/магента, что уже используется. Не меняй геометрию
  (углы/сетку/звёзды) — только источник цвета.
- `CockpitPanel.svelte`, `CockpitHUD.svelte`, `CockpitFirstPersonButton.svelte`:
  замени хардкод `#2dd4bf` / `rgba(45, 212, 191, ...)` на
  `var(--cockpit-accent)` / `var(--cockpit-glow)`.

## Шаг 4 — Фон кабины

В `CosmicCockpitLayout.svelte` (не заводить новый компонент) фон уже
частично покрыт через `CockpitFrame`'s `.frame-grid`. Если нужен фон и вне
рамки (за пределами `cockpit-frame-wrapper`), переиспользуй тот же паттерн
grid+radial-gradient "звёзд", а не JS/canvas — только CSS, статично (без
анимации, чтобы не грузить GPU).

## Ограничения

- Только CSS и Svelte scoped styles. Никакой новой JS-анимации фона.
- Не менять структуру/логику компонентов, кроме добавления класса
  `cockpit-gradient-text` и точечных `style` для `--cockpit-text-delay`.
- Не трогать `ToastNotification.svelte` и любые компоненты вне
  `frontend/src/widgets/cosmic-cockpit/` и `features/cosmic-ui/`.
- Не создавать новых цветовых переменных с нуля — только алиасы в
  существующую Allotropic Carbon палитру.

## Проверка

- `npm run check` — 0 ошибок.
- `npm run test:unit` — все тесты проходят.
- Ручная проверка по `docs/MANUAL_TEST_CHECKLIST_COCKPIT.md` § 6
  (визуальная тема) — дополни секцию § 6 конкретными пунктами под
  анимированный градиент (разная фаза у заголовков, контраст на всех
  кадрах, `prefers-reduced-motion` отключает анимацию, HUD-метрики остаются
  легко читаемыми).
- Публичный (неавторизованный) вид — без изменений, без класса
  `cosmic-cockpit` и без градиентного текста.

## Порядок действий

1. Добавь алиасы переменных (шаг 1) — не создавай новых цветов.
2. Добавь общий `.cockpit-gradient-text` с раздельной фазой (шаг 2).
3. Точечно замени хардкод в `CockpitFrame`, `CockpitPanel`, `CockpitHUD`,
   `CockpitFirstPersonButton` (шаг 3).
4. Проверь фон (шаг 4) — без новых JS-анимаций.
5. Прогони `npm run check` и `npm run test:unit`, исправь ошибки.
6. Обнови `docs/MANUAL_TEST_CHECKLIST_COCKPIT.md` § 6 и оставь находки в
   `docs/MANUAL_TEST_FEEDBACK.md`.
7. **Не пушить** — остановиться для подтверждения.
