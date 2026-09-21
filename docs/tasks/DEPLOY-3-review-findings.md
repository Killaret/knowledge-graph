# DEPLOY-3 — разбор ревью

Ревьюер: Claude Code. Дата: 2026-09-21. Реализация: Devin (09-18).

**Вердикт: отклонено** — из четырёх compose-файлов исправлены три; пропущен тот, где время
старта важнее всего.

## Проверено

| Файл | `start_period` NLP |
|---|---|
| `docker-compose.yml` | 180s |
| `docker-compose.test.yml` | 180s |
| `docker-compose.personal.yml` | 180s |
| **`docker-compose.deploy.yml`** | **600s** |

`backend/Dockerfile`: `FROM alpine:3.19` — запинен. Оба compose парсятся.

Достаточно ли 180 с: сегодня `kg-test-nlp` на пересобранном образе стал healthy, пока шли
соседние сборки — модель лежит в образе (`snapshot_download` на этапе сборки), в рантайме
сеть не нужна. Запас есть.

## Блокер

Задача: «600s → 180s во всех compose». В `docker-compose.deploy.yml` — 600s. Статус сдачи честно
говорит «во всех трёх», то есть deploy-компоуз исключён сознательно, но задача его не исключала,
а это единственный файл, который поднимает стек на чужой машине. Одна строка.

## Доработка `5adf09f` — принято (2026-09-21, вечер)

`docker-compose.deploy.yml`: NLP healthcheck `interval 15s / timeout 5s / retries 24 / start_period 180s` —
как в трёх остальных.