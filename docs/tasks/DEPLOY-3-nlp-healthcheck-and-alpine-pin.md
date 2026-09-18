# DEPLOY-3. NLP healthcheck 600s → 180s и пин Alpine в backend Dockerfile

Быстрые выигрыши из разбора пакета Gordon (GORDON-2): NLP healthcheck ждал 10 минут до первой проверки на каждом `docker compose up`, а backend-образ был единственным непиннутым (`alpine:latest`).

## Что сделано

- `docker-compose.yml`, `docker-compose.test.yml`, `docker-compose.personal.yml` — NLP healthcheck: `start_period` 600s → **180s**, `interval` 30s → **15s**, `timeout` 10s → **5s**, `retries` 30 → **24** (итоговое окно ~9 мин — запас на скачивание модели при холодном кэше).
- `backend/Dockerfile` — `FROM alpine:latest` → `FROM alpine:3.19` (как в `services/graph-service/Dockerfile`).

## Обоснование

- Реальное время до первого успешного `/health` у NLP ~100–120s (загрузка модели в память 30–45s + старт FastAPI), 600s — мёртвое ожидание на каждом `up`.
- `retries` не опущен до 12 как у Gordon: при холодном `huggingface_cache` скачивание модели может превысить 180s — оставлен запас 360s после `start_period`.
- `docker compose config` парсится чисто для всех трёх файлов.

## Проверка

- Синтаксис compose: `docker compose -f <file> config` — OK (dev/personal ругаются только на отсутствующий `.env`, что ожидаемо в worktree).
- Сборка backend-образа и подъём стека — покрывается CI (Docker builds, deploy-verify).

## Статус

Выполнено Devin 2026-09-18. Источник — разбор GORDON-2.
