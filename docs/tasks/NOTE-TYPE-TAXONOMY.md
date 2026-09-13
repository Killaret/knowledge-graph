# NOTE-TYPE-TAXONOMY: таксономия типов заметок

## Статус

**2026-09-13.** Обсуждение между владельцем и Devin. Нужно привести списки типов в единый, логичный порядок «от большего к меньшему» и зафиксировать его в домене (`NoteType` value object), backend-валидации, OpenAPI и frontend. После принятия постановки — реализация в ветке `BATCH-DDD-VALIDATION`.

## Текущее состояние

### Backend

- `ValidCelestialBodyTypes` в `backend/internal/interfaces/api/common/validation/validators.go` перечисляет 16 типов.
- `POST /notes`, `/notes/batch/create`, `/import/batch` валидируют `type` через Gin `oneof`.
- `note.NewNote` не валидирует `noteType` (записано в `BATCH-DDD-VALIDATION.md`).

### Frontend

- `CelestialBody.ALL` в `frontend/src/entities/shared/model/celestial-body.ts` содержит 16 типов в порядке:
  `star, planet, moon, comet, galaxy, nebula, asteroid, satellite, blackhole, debris, dust, technical, unknown, reality_rift, chromatic_maw, void_whisper, cosmic_abomination`.
- `CelestialBody.UI_TYPES` = `isUi === true` даёт 10 типов:
  `star, planet, comet, galaxy, nebula, asteroid, satellite, blackhole, debris, dust`.
- `moon` есть в `ALL`, но **нет в `UI_TYPES`** — его нельзя выбрать в селекторе, хотя backend разрешает.
- `TypeSelector` берёт `defaultSelected = types[0]`, поэтому по умолчанию выбирается `star` (первый в `UI_TYPES`).

### OpenAPI

- Несколько схем содержат разные enum-списки: где-то 8 типов, где-то 16, порядок разнится. Это дрейф.

## Проблемы

1. **Нет единого источника правды.** Backend, frontend и OpenAPI держат списки вручную.
2. **Порядок `UI_TYPES` не логичен.** Сверху `star`, `planet` — пользователь думает, что они «самые большие», хотя по масштабу галактика/туманность больше.
3. **Отсутствует `moon` в UI**, хотя это логичная ступень между `planet` и `satellite`.
4. **Нет чёткой шкалы** для сортировки; сейчас порядок зависит от порядка объявлений в `ALL`.
5. **Аномалии (`reality_rift` и др.)** в `ALL`, но не для пользовательского выбора — это ок, но они должны быть явно отмечены как системные.

## Предлагаемая шкала

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

Отсортированный по `scaleRank` сверху-вниз:

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

(11 типов вместо 10; `moon` добавляется, `technical`/`unknown`/аномалии исключаются.)

### Пример для "аниме"

- `galaxy` = Media
- `star` = Anime
- `planet` = Concrete anime series
- `moon` = Characters / arcs of that series
- `comet` = Season release / event
- `satellite` = Episode checklist
- `asteroid` = Quote / scene note
- `dust` = Raw link from browser

## Что нужно сделать

1. **Домен (`NoteType` value object)**
   - Создать `backend/internal/domain/note/type.go`.
   - Типы с `scaleRank`, `isUserSelectable`, `label`.
   - Валидация в `NewNote`/`SetType`.

2. **Backend**
   - Убрать `ValidCelestialBodyTypes` из `interfaces` или сделать его производным от доменного списка.
   - Обновить `oneof` в DTO и OpenAPI одним единым списком.

3. **Frontend**
   - Добавить `scaleRank` в `CelestialBodyProps`.
   - Добавить `moon` в `UI_TYPES`.
   - Сортировать `UI_TYPES` по `scaleRank`.
   - В `TypeSelector` сделать `defaultSelected = "star"` явно, а не `types[0]`.

4. **OpenAPI**
   - Привести все enum-списки к одному порядку и составу.

5. **Тесты**
   - `celestial-body.test.ts` — проверить сортировку и 11 UI-типов.
   - `note/type_test.go` — валидация.
   - `note_handler_test.go` — не-UI типы отклоняются.

## Открытые вопросы для владельца / Claude Code

1. Согласны ли с `scaleRank` выше?
2. Добавлять ли `moon` в UI?
3. Должен ли `blackhole` быть выше `star` (он массивнее, но это «проблема», а не тема)?
4. Что делать с `comet`/`satellite`/`asteroid`/`dust` — текущий порядок логичен?
5. Стоит ли `debris` переместить в конец (ниже `dust`) как архивное состояние?

## Связанное

- `docs/tasks/BATCH-DDD-VALIDATION.md`
- `docs/tasks/IMP-1-import-type-restrictions.md`
- `docs/tasks/IMP-3-import-ghost-type-ux.md`
- `frontend/src/entities/shared/model/celestial-body.ts`
- `backend/internal/interfaces/api/common/validation/validators.go`
- `backend/openAPI.yaml`
