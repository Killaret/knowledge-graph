# AUD-2. Замечания ревью

Ревьюер: Claude Code. Исполнитель: Devin. Предмет: коммит `889c33e` по задаче [`AUD-2-data-isolation.md`](AUD-2-data-isolation.md).

Статус: **не принято, есть блокер.**

## Блокер. Сидер не работает — тестовую учётку создать нечем

`backend/cmd/seed/main.go:56-58`:

```go
var roleID uuid.UUID
if err := database.WithContext(ctx).Raw("SELECT id FROM user_roles WHERE name = 'user' LIMIT 1").Scan(&roleID).Error; err != nil {
```

GORM `Scan` в `uuid.UUID` (это `[16]byte`) пытается разложить строку в массив байт поэлементно и падает. Проверено запуском против базы тест-стека:

```
APP_ENV=test SEED_TEST_USER_PASSWORD=verify-pass DATABASE_URL=...knowledge_test go run ./cmd/seed

2026/09/05 20:01:19 failed to look up 'user' role: sql: Scan error on column index 0,
name "id": converting driver.Value type string ("3916db86-7cd1-49b2-9aae-687e369fb1cf")
to a uint8: value out of range
exit status 1
```

Сидер не отработал ни разу. Проверка `if roleID == uuid.Nil` на строке 60 недостижима — до неё выполнение не доходит.

Чинится сканированием в строку с последующим `uuid.Parse`, либо `Pluck`, либо `Row().Scan`. Ветку с `uuid.Nil` после этого стоит сохранить: она осмысленна.

**Почему это блокер, а не мелочь.** Миграция `029` удаляет старую тестовую учётку, а создать новую нечем. В связке это оставляет тест-стек без пользователя с нулевым UUID, на которого завязан режим `SKIP_AUTH`.

Проверено на базе тест-стека, а не выведено рассуждением. Чтение в режиме `SKIP_AUTH` пользователя не требует — `applyNoteScope` (`note_repo.go:40-44`) обходит фильтр и к таблице `users` не обращается. Запись требует: `notes.creator_id` несёт внешний ключ на `users(id)` (`migrations/016_add_auth_and_sharing.up.sql:17`), а middleware пишет в него нулевой UUID (`skip_auth.go:31`).

```
INSERT INTO notes (..., creator_id, ...) VALUES (..., '00000000-0000-0000-0000-000000000000', ...);

ERROR:  insert or update on table "notes" violates foreign key constraint "notes_creator_id_fkey"
DETAIL:  Key (creator_id)=(00000000-0000-0000-0000-000000000000) is not present in table "users".
```

Та же зависимость у `user_settings.user_id`, `user_achievements.user_id`, `note_likes.user_id` и таблиц шаринга — все `NOT NULL REFERENCES users(id)`. Значит любой сценарий, который создаёт заметку, меняет настройку или начисляет достижение, падает; сценарии только на чтение проходят. Отказ выглядит плавающим, хотя он системный.

**Про способ проверки.** `go test ./...` зелёный и это правда, но ни один тест не выполняет путь сидера против базы — поэтому отказ был невидим. Постановка требовала критерий 4 «на тест-стеке после `seed-test-data.ps1` вход под тестовой учёткой работает»; он не выполнялся, и в отчёте честно сказано, что живая проверка оставлена ревьюеру. Нужен тест, который поднимает базу и прогоняет сидер, иначе следующая регрессия здесь снова пройдёт мимо.

## Состояние тест-стека после ревью

Проверка выполнялась запуском свежесобранного сервера против базы тест-стека, поэтому миграции применились: `029` отработала.

```
schema_migrations: 029, 028, 027
select count(*) from users  ->  0
```

**Тестовая база сейчас без пользователей.** Это ожидаемое поведение AUD-2, а не поломка, но до починки сидера тест-стек непригоден для прогонов с `SKIP_AUTH`. После правки — прогнать `seed-test-data.ps1` и убедиться, что пользователь создан.

## Что принято

Проверено исполнением, замечаний нет.

| Критерий | Проверка | Итог |
|---|---|---|
| 1. `go test ./...`, `go vet ./...` | Прогон: 51 пакет, ни одного `FAIL`, exit 0 | Зелёные |
| 2. Отказ старта при `SKIP_AUTH=true` вне теста | Запуск сервера в трёх окружениях | Подтверждено |
| 3. Ни один ответ не содержит `Cache-Control: public` | Живые запросы к пяти эндпоинтам | Подтверждено |
| 5. Нулевой UUID в миграциях | `grep` по `backend/migrations/` | Соответствует |

**Критерий 2, вывод запуска:**

```
APP_ENV=development -> FATAL: Failed to load configuration:
    SKIP_AUTH=true is only allowed when APP_ENV=test; current APP_ENV=development
APP_ENV=production  -> FATAL: ... current APP_ENV=production
APP_ENV=test        -> Config loaded: alpha=0.50, ... (старт продолжается)
```

**Критерий 3, живые заголовки** (сервер собран из `889c33e`, база тест-стека):

```
/api/v1/notes             Cache-Control: private, max-age=60   Vary: Authorization, Cookie
/api/v1/notes/search      Cache-Control: private, max-age=30   Vary: Authorization, Cookie
/api/v1/graph/all         Cache-Control: private, max-age=300  Vary: Authorization, Cookie
/api/v1/me/graph/fresh    Cache-Control: private, max-age=0    Vary: Authorization, Cookie
/api/v1/graph/analytics   Cache-Control: private, max-age=300  Vary: Authorization, Cookie
```

Ни одного `public`.

**Критерий 5.** `019_add_test_user.up.sql:10` — вставка осталась нетронутой, как и требовала постановка. `019_add_test_user.down.sql:2` — UUID исправлен на `...0000`. `029_remove_test_user.up.sql:7` — удаление. Больше нулевой UUID в миграциях не встречается.

**Про тест на заголовки — отдельно.** `router_cache_test.go` проверялся мутацией: в `cacheControlMiddleware` временно возвращён `public`, прогон упал с `"public, max-age=0" should not contain "public"`. Тест действительно ловит дефект, а не только фиксирует текущее поведение. Файл восстановлен `git checkout`.

## Не проверено

- **Критерий 4** — заблокирован блокером.
- **Критерий 6** — прогон `029` на чистой базе и на базе с уже применённой `019`. На базе с `019` проверено (тест-стек), на чистой — нет.
- Негативные ветки сидера при этом работают верно: `APP_ENV=development` даёт `seeder can only run with APP_ENV=test`, отсутствие `SEED_TEST_USER_PASSWORD` — отказ с ненулевым кодом.

## Вопрос по конструкции, отдельно от блокера

`SKIP_AUTH` — тестовый обход, но он зависит от конкретной строки в базе. Именно эта связанность превратила баг сидера в неработающий тест-стек.

`notes.creator_id` допускает `NULL`, и вставка с `creator_id = NULL` проходит — то есть обход в принципе может не опираться на персистентного пользователя, либо фикстуру должен создавать тестовый контур, а не продуктовый путь. Стоит решить осознанно, а не восстанавливать прежнюю схему по умолчанию.

Комментарий `skip_auth.go:55` — «Test user must exist in DB (migration 019)» — устарел в любом случае: миграция `019` этого пользователя больше не даёт.

## Мелочь

`backend/cmd/seed/main.go:78` вставляет логин `testuser`. Это согласуется с `scripts/testing/seed-test-data.ps1:26` и `.sh:27`, которые всегда создавали `testuser` через API — здесь всё в порядке.

Но миграция `019` создавала `test_user` (через подчёркивание), и на это имя есть живая ссылка: `frontend/tests/auth-functional.spec.ts:30` ищет `text=test_user`. После удаления учётки миграцией `029` этот спек может начать падать или, хуже, молча проходить по другому селектору из того же `or`-списка. Проверить при следующем прогоне E2E.
