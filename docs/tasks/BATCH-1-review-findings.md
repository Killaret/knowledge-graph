# Ревью BATCH-1, NOTE-TYPE-TAXONOMY и BATCH-DDD-1 (Claude Code, 2026-09-13)

Три строки доски — один дифф, одно ревью.

**Принято всё.** Блокеров нет. Ниже — что проверено мутациями, и мелочь.

## BATCH-1: batch-маршруты

Маршруты `POST /notes/batch/create`, `POST /notes/batch/delete`, `POST /import/batch` —
все под `writeLimiter` и глобальным `JWTAuth` (`router.go:172-177`). Лимиты: `notes`
`max=50` и в create, и в import; `links` `max=50`.

**Владение при пакетном удалении** — проверено мутацией. Убрал проверку
`!n.IsOwnedBy(userID)` в `DeleteBatch`:

```
--- FAIL: TestDeleteBatchNotes_ForeignNote
```

Семантика — «чужая заметка в списке → 404 и ничего не удалено» — такая же, как в SEC-1.

**Защита от перезаписи по клиентскому id** (`52f23bb`): `FindByID` возвращает
`(nil, nil)` на отсутствие, так что ветка «существует → отказ» работает, а не роняет
каждый новый id. Тест `TestImportBatch_ClientIDCollisionWithExisting` есть. У
`batchNoteItem` поля `id` нет вовсе — через create перезаписать нечего.

**Создатель берётся из контекста**, не из тела — `TestCreateBatchNotes_SetsCreator`.

### Мелочь

- **Оракул существования.** Проверка клиентского id в `ImportBatch` идёт через
  `FindByID` без учёта владельца: чужой id → «note with this id already exists». Это
  подтверждает существование чужой заметки, тогда как `GET` на неё отвечает 404. UUID v4
  не перебираются, так что практической утечки нет; но политика «существование не
  подтверждаемо» нарушена в одном месте. Лечится ответом «id недопустим» без различения
  причин.
- `deleteBatchRequest.IDs` — `required,dive,uuid` **без `max`**. Create и import
  ограничены полусотней, delete — нет. Один вызов на десять тысяч id даст десять тысяч
  `FindByID`. Поставить `max=50` для симметрии.

## NOTE-TYPE-TAXONOMY и BATCH-DDD-1: `NoteType` как value object

**Домен действительно сторожит.** Мутация — `NewType` принимает любую строку:

```
--- FAIL: TestNewType, TestMustType, TestNoteTypeJSON          (domain)
--- FAIL: TestCreateNoteInvalidTypeInMetadata,
          TestCreateBatchNotesInvalidTypeInMetadata,
          TestImportBatch_InvalidTypeInMetadata                 (handlers)
--- FAIL: TestIsValidCelestialBodyType                          (validation)
```

Падают все три слоя — значит обработчики реально зависят от доменной проверки, а не
от собственной копии.

**База не мешает.** `notes.type VARCHAR(50) NOT NULL DEFAULT 'star'`
(`migrations/014`), CHECK-ограничения нет — `moon` принимается без миграции.

**Фронтенд.** `moon` в `UI_TYPES` со `scaleRank: 50`, сортировка по рангу, 8/8 тестов
`celestial-body` зелёные.

### Мелочь — одна, но с последствием

Список из 17 типов живёт **в пятнадцати копиях**: семь тегов `oneof=` в
`note_handler.go` (строки 131, 142, 166, 747, 875, 879, 1058) и восемь `enum` в
`openAPI.yaml`. Сверил все пятнадцать с реестром `type.go` скриптом — совпадают
**сегодня**. Постановка BATCH-DDD-1 прямо разрешает `oneof` как синтаксическую
проверку, так что это не нарушение, но восемнадцатый тип потребует шестнадцати
правок, и одна забытая даст 400 в batch при 201 в single.

Дешёвая страховка — тест, который читает теги структур через `reflect` и enum'ы из
`openAPI.yaml` и сравнивает с `note.AllTypeStrings()`. Половину работы уже делает
`TestRouterMatchesOpenAPISpec`: спеку он парсит. Не блокер; кандидат в первую же
задачу, которая тронет типы.

Вариант A из постановки (`NewNote(...) (*Note, error)`) — согласен, что чище, и согласен,
что не сейчас: 80 тестов ради формы, когда защита уже есть по существу.
