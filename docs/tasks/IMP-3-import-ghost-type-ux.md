# IMP-3. Подсказки типов в импорте и неполный список в ghost-форме

Постановка для Devin. Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит пользователь/Claude Code, реализует Devin, проверяет Claude Code на живом Personal-стеке.

## Зачем

В двух местах создания/импорта заметки сейчас плохо с типами:

1. **Массовый импорт** (`/import/bookmarks`) показывает тип в `<select>` только как одно слово (Stars, Planets, Asteroids и т.п.) — без пояснения, что этот тип означает. Пользователь не понимает, какой тип выбрать.
2. **Создание заметки через ghost-форму** (туманная кнопка-призрак в левом верхнем углу графа, горячая клавиша `N`) показывает **неполный список типов**.

Оба места должны использовать единый источник правды — `CelestialBody.UI_TYPES` — и показывать emoji, название и краткое описание (или tooltip).

## Что выяснено в коде

### Импорт

- `frontend/src/routes/import/bookmarks/+page.svelte` использует жёстко закодированный массив `noteTypes` (строки 22-39) и plain `<select>` (строки 346-355).
- Опции рендерятся через `t(`filter.type.${nt}`)`.
- Переводы `filter.type.*` (`frontend/src/shared/utils/i18n/messages/graph.ts`) — это множественные названия без описания.
- В то же время `CelestialBody` уже содержит `description` и `example`, а i18n-ключи `celestialBody.type.<type>.description` / `example` заполнены (`frontend/src/shared/utils/i18n/messages/ui.ts`).
- Компонент `TypeSelector` (`frontend/src/components/molecules/TypeSelector.svelte`) уже умеет показывать emoji, label, description, example и `title`-tooltip.

### Ghost-форма

- `frontend/src/features/graph-ui/modals.svelte` (строки 70-86) открывает ghost-форму и передаёт `types={CelestialBody.UI_TYPES}` в `NoteForm`.
- `NoteForm` использует `TypeSelector`.
- `CelestialBody.UI_TYPES` (`frontend/src/entities/shared/model/celestial-body.ts:517`) включает 10 пользовательских типов: star, planet, comet, galaxy, nebula, asteroid, satellite, blackhole, debris, dust.
- Тем не менее владелец отмечает, что в ghost-форме виден неполный список. Возможные причины:
  - `.ghost-note-form` имеет `max-width: min(420px, calc(100vw - 140px))` — `TypeSelector` может обрезаться;
  - `TypeSelector` оборачивается, но часть кнопок уезжает за границу или уходит за пределы экрана;
  - отличие от `CreateNoteModal` / `EditNoteModal` — там модальное окно даёт больше пространства.

## Что сделать

### 1. Подсказки в импорте
- Заменить hardcoded `noteTypes` и plain `<select>` на `CelestialBody.UI_TYPES`.
- Показывать emoji, название и краткое описание для каждого типа.
- Возможные варианты: использовать `TypeSelector` (если влезает в таблицу) или `<select>` с `title`/`aria-label` на `<option>` (tooltip по ховеру).
- Это пересекается с [`IMP-1`](IMP-1-import-type-restrictions.md) — селектор в импорте должен быть ограничен `CelestialBody.UI_TYPES`.

### 2. Неполный список в ghost-форме
- Проверить, что `TypeSelector` рендерит все 10 `UI_TYPES`.
- Если кнопки обрезаются — добавить `max-height` + `overflow-y: auto`, расширить форму или переключить `TypeSelector` в компактный dropdown.
- Убедиться, что `NoteForm` в режиме `ghost` и в `CreateNoteModal` / `EditNoteModal` используют одинаковый, полный набор.

### 3. Единообразие подсказок
- В ghost-форме должны быть видны описания/примеры так же, как в `CreateNoteModal`.
- Возможно, стоит добавить мини-hint рядом с селектором.

### 4. Тесты
- Юнит-тест: `CelestialBody.UI_TYPES` содержит ровно 10 типов.
- Playwright / Vitest: в ghost-форме присутствуют все 10 `data-type` (star, planet, comet, galaxy, nebula, asteroid, satellite, blackhole, debris, dust) и ни одного `technical`/`unknown`/`reality_rift`.
- Playwright / Vitest: в превью импорта у каждого варианта типа видна подсказка (tooltip или description).

## Критерии приёмки

1. В `/import/bookmarks` пользователь видит, что означает каждый тип (описание/пример) рядом с названием или по ховеру.

## Связанное

- `docs/tasks/NOTE-TYPE-TAXONOMY.md` — актуальный состав `UI_TYPES` и порядок типов; перед реализацией IMP-3 убедиться, что таксономия зафиксирована.
2. В `/import/bookmarks` селектор типа не содержит `technical`, `unknown`, аномалии и `moon` — только `CelestialBody.UI_TYPES`.
3. В ghost-форме (кнопка-призрак / `N`) доступны все 10 типов из `CelestialBody.UI_TYPES`; ни один не обрезается UI.
4. В ghost-форме у каждого типа отображается emoji и, по возможности, краткое описание.
5. Линтер/типчек и `npm run test:unit` проходят.
6. Ручная проверка на Personal-стеке: открыть ghost-форму и убедиться, что `dust`, `debris` и `blackhole` присутствуют и кликабельны.

## Что приложить к результату

Скриншот `/import/bookmarks` с открытым селектором типа и видимыми описаниями. Скриншот ghost-формы с 10 типами.