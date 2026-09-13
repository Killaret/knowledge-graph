# NOTE-TYPE-TAXONOMY: таксономия типов заметок

## Статус

**2026-09-14.** Реализация завершена Devin в ветке `devin/batch-api-37`; передана на ревью Claude Code.
- `blackhole` выше `star`.
- `moon` включается в `UI_TYPES`.
- `debris` остаётся ниже `dust`.
- `dust` остаётся для быстрых заметок, которые потом обогащаются.
- Дефолтный тип — `star` (явно, не первый элемент списка).
- `NoteType` value object реализован; `NewNote`/`ReconstructNote`/`SetType` работают с `NoteType`.
- Backend, frontend и OpenAPI синхронизированы одним каноническим порядком.

## Текущее состояние

### Backend

- `backend/internal/domain/note/type.go` — `NoteType` value object с `scaleRank`, `isUserSelectable`, валидацией, фолбеками.
- `note.NewNote`/`NewNoteWithCreator`/`ReconstructNote`/`SetType` принимают `NoteType`; пустой/невалидный тип отклоняются.
- `backend/internal/interfaces/api/common/validation/validators.go` — `IsValidCelestialBodyType` делегирует `note.NewType`.
- `POST /notes`, `/notes/batch/create`, `/import/batch`, bookmarklet, batch обновления — все преобразуют строку в `NoteType` и отдают 400 при невалидном значении.
- `backend/openAPI.yaml` — все enum-списки типов заметок приведены к единому каноническому порядку.

### Frontend

- `CelestialBody.ALL` в `frontend/src/entities/shared/model/celestial-body.ts` содержит 16 типов в каноническом порядке:
  `galaxy, nebula, blackhole, star, planet, moon, comet, satellite, asteroid, dust, debris, technical, unknown, reality_rift, chromatic_maw, void_whisper, cosmic_abomination`.
- Каждый тип имеет `scaleRank`.
- `CelestialBody.UI_TYPES` = `isUi === true`, отсортирован по `scaleRank` убыванию, 11 типов:
  `galaxy, nebula, blackhole, star, planet, moon, comet, satellite, asteroid, dust, debris`.
- `moon` теперь в `UI_TYPES`.
- `TypeSelector` явно ищет `star` как дефолт; `CreateNoteModal`, `NoteForm`, graph-формы по умолчанию используют `star`.
- Селекторы импорта, фильтры графа и home page используют `CelestialBody.UI_TYPES`.

### OpenAPI

- Все `enum` типов заметок в `backend/openAPI.yaml` синхронизированы: 16 типов в каноническом порядке.

## Проблемы

1. **Нет единого источника правды.** Backend, frontend и OpenAPI держат списки вручную.
2. **Порядок `UI_TYPES` не логичен.** Сверху `star`, `planet` — пользователь думает, что они «самые большие», хотя по масштабу галактика/туманность больше.
3. **Отсутствует `moon` в UI**, хотя это логичная ступень между `planet` и `satellite`.
4. **Нет чёткой шкалы** для сортировки; сейчас порядок зависит от порядка объявлений в `ALL`.
5. **Аномалии (`reality_rift` и др.)** в `ALL`, но не для пользовательского выбора — это ок, но они должны быть явно отмечены как системные.

## Решённая шкала

**Решения владельца (2026-09-13):**
- Порядок «от большего к меньшему» согласован.
- `blackhole` находится **выше `star`** (массивнее и иное смысловое наполнение).
- `moon` **включается в `UI_TYPES`**.
- `debris` остаётся, но **ниже `dust`**.
- `dust` остаётся для быстрых заметок, которые потом обогащаются.
- Дефолтный тип при создании заметки — **`star`**.

Идея — единый `scaleRank`: чем больше число, тем «крупнее» объект в космической иерархии (и шире смысловой охват заметки).

| Тип | scaleRank | isUi | Группа | Смысл |
|-----|-----------|------|--------|-------|
| `galaxy` | 100 | да | Domain | Широкий домен, объединяет несколько звёзд-тем |
| `nebula` | 90 | да | Domain | Неоформившаяся масса идей/черновиков |
| `blackhole` | 85 | да | Central | Сложная неразрешённая задача, большой баг |
| `star` | 80 | да | Central | Центральная тема/столп |
| `planet` | 60 | да | Sub-topic | Крупный подраздел темы |
| `moon` | 50 | да | Sub-topic | Деталь/аспект, привязанный к планете |
| `comet` | 40 | да | Temporary | Временная/событийная заметка |
| `satellite` | 35 | да | Utility | Чек-лист, шаблон, конфиг, сниппет |
| `asteroid` | 30 | да | Fragment | Цитата, TODO, быстрая мысль |
| `dust` | 10 | да | Inbox | Быстро захваченное, необработанное |
| `debris` | 5 | да | Archive | Архив, устаревший материал |
| `technical` | 0 | нет | System | Системная/мета-заметка |
| `unknown` | 0 | нет | System | Fallback |
| `reality_rift` | 0 | нет | Anomaly | Специальный маркер графа |
| `chromatic_maw` | 0 | нет | Anomaly | Специальный маркер графа |
| `void_whisper` | 0 | нет | Anomaly | Специальный маркер графа |
| `cosmic_abomination` | 0 | нет | Anomaly | Специальный маркер графа |

### Пользовательский список (`UI_TYPES`)

Отсортированный по `scaleRank` сверху-вниз (11 типов):

1. `galaxy`
2. `nebula`
3. `blackhole`
4. `star`
5. `planet`
6. `moon`
7. `comet`
8. `satellite`
9. `asteroid`
10. `dust`
11. `debris`

### Дефолтный выбор

- Селектор по умолчанию выбирает **`star`**, не первый элемент списка.

### Пример для "аниме"

- `galaxy` = Media
- `star` = Anime
- `planet` = Concrete anime series
- `moon` = Characters / arcs of that series
- `comet` = Season release / event
- `satellite` = Episode checklist
- `asteroid` = Quote / scene note
- `dust` = Raw link from browser

## Что сделано

1. **Домен (`NoteType` value object)**
   - Создан `backend/internal/domain/note/type.go`.
   - Все типы с `scaleRank`, `IsUserSelectable`, `String`.
   - `NewNote`/`SetType` валидируют `NoteType`; пустые/невалидные значения отклоняются.

2. **Backend**
   - `IsValidCelestialBodyType` делегирует `note.NewType`.
   - `oneof` в DTO и OpenAPI обновлены каноническим списком.
   - Все обработчики и сервисы импорта преобразуют строку в `NoteType`.

3. **Frontend**
   - Добавлен `scaleRank` в `CelestialBodyProps`.
   - `moon` включён в `UI_TYPES`.
   - `UI_TYPES` сортируется по `scaleRank` убыванию.
   - `TypeSelector` выбирает `star` явно.
   - Все селекторы/фильтры (`CreateNoteModal`, `EditNoteModal`, `NoteForm`, graph-формы, импорт, фильтры графа/home page) используют `CelestialBody.UI_TYPES`.

4. **OpenAPI**
   - Все enum-списки типов заметок приведены к единому порядку и составу.

5. **Тесты**
   - `celestial-body.test.ts` — 11 UI-типов, порядок, `moon` присутствует, аномалии отсутствуют.
   - `note/type_test.go` и `note/entity_test.go` — валидация `NoteType`.
   - `notehandler` — негативные сценарии с невалидными типами.

## Результаты верификации

- `cd backend && go test ./...` — зелёное.
- `cd backend && go vet ./...` — чисто.
- `cd frontend && npm run test:unit` — 1381/1381 passed.
- `cd frontend && npm run build` — успешно.
- `cd frontend && npm run check` — 0 errors, 0 warnings.
- `cd backend && go test ./cmd/server/...` (OpenAPI router contract) — проходит.

## Следующий шаг

- Ревью Claude Code: `backend/internal/domain/note/type.go`, `entity.go`, `note_handler.go`, `validators.go`, `openAPI.yaml`, `frontend/src/entities/shared/model/celestial-body.ts`, `TypeSelector.svelte`, селекторы/фильтры.

## Открытые вопросы для Claude Code / владельца

- Проверить `scaleRank` и иерархию на предмет интуитивности.
- Подтвердить, что `moon` должен быть в пользовательском селекторе.

## Связанное

- `docs/tasks/BATCH-DDD-VALIDATION.md`
- `docs/tasks/IMP-1-import-type-restrictions.md`
- `docs/tasks/IMP-3-import-ghost-type-ux.md`
- `frontend/src/entities/shared/model/celestial-body.ts`
- `backend/internal/interfaces/api/common/validation/validators.go`
- `backend/openAPI.yaml`
