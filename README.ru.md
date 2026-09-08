# 🌌 Knowledge Graph

<div align="center">

[![CI](https://github.com/Killaret/knowledge-graph/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/Killaret/knowledge-graph/actions/workflows/main.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?logo=svelte&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.11-3776AB?logo=python&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)

**База знаний, в которой заметки образуют граф, и по нему можно летать.**

[English version](README.md) • [Быстрый старт](#-быстрый-старт) • [Архитектура](#️-архитектура) • [Инженерная практика](#-инженерная-практика) • [Документация](#-документация)

![3D-вид графа](docs/assets/a1-3d-visual-regression/3d-baseline.png)

</div>

> Нормативная версия — английская, [`README.md`](README.md). Этот файл переводится
> в том же изменении, что и оригинал; при расхождении верен английский.

---

## Что это

Заметки — узлы, связи между ними — рёбра. NLP-сервис укладывает каждую заметку
в мультиязычное векторное пространство, поэтому близкие по смыслу заметки
притягиваются независимо от того, написаны они по-русски или по-английски.
Результат рисуется навигируемой трёхмерной сценой — звёзды, планеты, кометы —
либо обычным двумерным графом.

Опубликованные заметки образуют публичный граф сообщества: его можно смотреть
без учётной записи. После входа вид переключается на собственный граф.

## Происхождение

Проект вырос из собственной задачи автора: заметки копились быстрее, чем
находилась структура, способная их удержать, а инструмента, который считал бы
связи между ними главным объектом, а не побочным, не нашлось.

Продуктовая модель и архитектура — авторские, как и каждое решение в
[`docs/architecture/decisions/`](docs/architecture/decisions/): сначала гипотеза,
затем проработка, обсуждение с коллегами там, где это помогало, и принятие или
отклонение по существу. Отклонённые варианты записаны рядом с принятыми, потому
что именно они показывают ход рассуждения. Систему построил автор.

AI-агенты появились позже и работают по записанному порядку: один реализует,
другой проверяет, никогда не в одной сессии, и ни один не правит собственные
ограничения. Что строится и по каким критериям принимается — решается вне их.

---

## ✨ Возможности

- **Трёхмерная визуализация** — заметки как небесные тела, навигация камерой, адаптивный туман
- **Графовая структура** — типизированные связи, вес считается по семантической близости
- **Мультиязычная семантика** — одно пространство для 50+ языков: «машинное обучение» и "machine learning" оказываются рядом
- **Семантический поиск** — по векторной близости через pgvector
- **Рекомендации** — предрассчитываются из расстояния в графе и семантики
- **Публичный граф сообщества** — опубликованные заметки доступны анонимно
- **Черновики** — автосохранение в MongoDB по ходу набора
- **Аутентификация** — JWT, API-ключи, OAuth2 (Яндекс)
- **Достижения** — лёгкая геймификация
- **Облачный бэкап** — по расписанию на Яндекс.Диск

---

## 🚀 Быстрый старт

### Требования

Docker и Docker Compose. Остальное работает в контейнерах; Go 1.25+, Node.js 20+
и Python 3.11+ нужны только для прямого запуска сервисов.

### Запуск

```bash
# Стек разработки
docker compose up -d

# Личный экземпляр (свои порты и тома)
docker compose -f docker-compose.personal.yml up -d

# Изолированный тестовый стек, уничтожается после использования
docker compose -f docker-compose.test.yml up -d --build
```

Вспомогательные скрипты тестового стека:

```powershell
.\scripts\testing\start-test.ps1       # поднять
.\scripts\testing\seed-test-data.ps1   # засеять детерминированными данными
.\scripts\testing\stop-test.ps1        # остановить и уничтожить
```

### Где что слушает

| Стек | Фронтенд | Шлюз API | Примечание |
|---|---|---|---|
| Разработка | http://localhost:18081 | http://localhost:18080 | dev-сервер Vite на 5173 |
| Личный | http://localhost:18084 | http://localhost:18082 | здесь лежат настоящие данные |
| Тестовый | http://localhost:3002 | http://localhost:18083 | изолирован, одноразовый |

Полная карта портов — [`docs/DOCKER.md`](docs/DOCKER.md).

### Прямой запуск сервисов

```bash
cd backend      && go run ./cmd/server
cd frontend     && npm run dev
cd nlp-service  && uvicorn app.main:app --reload
```

---

## 🏗️ Архитектура

Четыре сервиса за шлюзом nginx.

```
                    ┌──────────────┐
                    │    nginx     │  шлюз, CORS, заголовки безопасности
                    └──────┬───────┘
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
   ┌────────────┐   ┌────────────┐   ┌───────────────┐
   │  Frontend  │   │  Backend   │   │ Graph Service │
   │ SvelteKit  │──►│    Go      │──►│  Go, gRPC +   │
   │  Svelte 5  │   │  Gin/GORM  │   │     HTTP      │
   └────────────┘   └─────┬──────┘   └───────┬───────┘
                          │                  │
                ┌─────────┼──────────┐       │
                ▼         ▼          ▼       ▼
         ┌───────────┐ ┌──────┐ ┌───────┐ ┌─────────┐
         │PostgreSQL │ │Redis │ │MongoDB│ │   NLP   │
         │ +pgvector │ │ кэш  │ │чернов.│ │ FastAPI │
         └───────────┘ └──────┘ └───────┘ └─────────┘
```

**Бэкенд** построен по Clean Architecture — `domain`, `application`,
`infrastructure`, `interfaces`, — и границы слоёв держит `depguard`, а не
договорённость. **Фронтенд** построен по Feature-Sliced Design поверх Atomic
Design, правила импортов проверяет ESLint.

**Graph Service** — отдельный сервис на Go, считает раскладку и обходы графа,
инвалидирует кэш через pub/sub. **NLP Service** строит эмбеддинги и извлекает
ключевые слова.

Решения зафиксированы: 18 ADR в [`docs/architecture/decisions/`](docs/architecture/decisions/),
модель C4 и UML — в [`docs/architecture/`](docs/architecture/README.md).

---

## 🔬 Инженерная практика

Самое интересное здесь — не список возможностей.

**Тесту не верят, пока не увидели его красным.** Каждый сторож в этом проекте
проверен тем, что ломали охраняемое и смотрели, упадёт ли тест: границы слоёв,
гейты покрытия, модель доступа к заметкам, детектор расхождения с CI. Зелёный
набор, который ни разу не падал, неотличим от отсутствующего.

**Правила держат машины, а не память.** Архитектурные границы проверяются через
`depguard` и правила импортов ESLint. Сгенерированная конфигурация
перегенерируется в CI, и сборка падает при расхождении. Локальный прогонщик
сверяется с workflow и краснеет, когда они разошлись. Ссылки в документации и
упомянутые в ней команды проверяются при каждом прогоне.

**Одна локальная команда повторяет CI.** `scripts/testing/check-all.ps1`
прогоняет те же шестнадцать фаз, докладывает каждый недоступный инструмент
явным пропуском с причиной и завершается ненулевым кодом при любом отказе.

**Два AI-агента, разведённые по ролям.** Реализация и ревью никогда не
совмещаются в одной сессии, и ни один агент не правит собственные ограничения.
Порядок — в [`docs/AI_AGENT_PROTOCOL.md`](docs/AI_AGENT_PROTOCOL.md), рабочая
доска — в [`docs/AI_HANDOFF.md`](docs/AI_HANDOFF.md).

Внешний аудит репозитория, его 22 находки и их закрытие — в
[`docs/EXTERNAL_AUDIT_2026-09.md`](docs/EXTERNAL_AUDIT_2026-09.md).

---

## 💻 Технологии

| Слой | Что используется |
|---|---|
| **Бэкенд** | Go 1.25, Gin, GORM, pgx/v5, asynq |
| **Фронтенд** | SvelteKit, Svelte 5 runes, TypeScript, Three.js |
| **Graph Service** | Go 1.25, gRPC + HTTP |
| **NLP** | Python 3.11, FastAPI, sentence-transformers, YAKE |
| **Данные** | PostgreSQL 16 + pgvector, Redis 7, MongoDB |
| **Инфраструктура** | Docker Compose, nginx |
| **Тестирование** | testify + testcontainers, Vitest, Playwright, Cucumber, Argos |

---

## 📁 Структура проекта

```
knowledge-graph/
├── backend/                  # API и воркеры на Go
│   ├── cmd/                  # server, worker, seed, cli, embed-recompute
│   ├── internal/
│   │   ├── domain/           # сущности и value objects
│   │   ├── application/      # сценарии использования
│   │   ├── infrastructure/   # хранилища, кэш, очередь
│   │   └── interfaces/       # HTTP-хендлеры и middleware
│   └── migrations/           # 30 SQL-миграций
├── frontend/                 # SvelteKit, Feature-Sliced Design
│   └── src/
│       ├── shared/           # примитивы, клиенты API, сторы
│       ├── entities/         # UI, привязанный к домену
│       ├── features/         # пользовательские возможности
│       ├── widgets/          # составные блоки
│       ├── components/       # atoms, molecules, organisms
│       └── routes/           # страницы
├── services/graph-service/   # раскладка и обходы графа
├── nlp-service/              # эмбеддинги и ключевые слова
├── docs/                     # архитектура, ADR, эксплуатация, тестирование
├── scripts/                  # тестирование, очистка, devops
└── tests/                    # BDD-наборы
```

---

## 🧪 Тестирование

993 юнит-теста фронтенда в 109 файлах, 47 Go-пакетов под тестами в бэкенде плюс
graph-service, интеграционные тесты на настоящих контейнерах через
testcontainers, E2E и BDD через Playwright, визуальная регрессия через Argos.

```bash
# всё, что гоняет CI, локально
.\scripts\testing\check-all.ps1
.\scripts\testing\check-all.ps1 -Quick    # без интеграционных

# по отдельности
cd backend     && go test ./...
cd backend     && go test -tags=integration -p=1 ./...
cd frontend    && npm run test:unit
cd nlp-service && pytest
```

Полный регрессионный цикл, поднимающий изолированный стек и убирающий его за
собой, описан в [`docs/REGRESSION_TEST_PLAN.md`](docs/REGRESSION_TEST_PLAN.md).

---

## 🔒 Безопасность

- Объектная авторизация на каждом маршруте заметки: чужая приватная отвечает
  `404`, так что её существование не подтверждается
- Публичный и внутренний периметры разделены — шлюз вычищает внутренние
  заголовки, а graph-service доверяет им только по выключенному по умолчанию флагу
- Приватные ответы отдаются как `Cache-Control: private` с `Vary`
- Токен доступа не принимается из строки запроса; OAuth использует PKCE `S256`
- Тестовый обход авторизации отказывается стартовать вне тестового окружения
- В CI только `npm ci`, `minimumReleaseAge` в `.npmrc`, гейт `npm audit`,
  Dependabot и проверка зависимостей на pull request

---

## 🛠️ Разработка

```bash
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph

cd backend     && go mod download
cd ../frontend && npm install
cd ../nlp-service && pip install -r requirements.txt
```

Качество кода:

```bash
cd backend  && golangci-lint run
cd frontend && npm run lint     # только проверка; npm run lint:fix — с правками
cd frontend && npm run check    # svelte-check
```

Справочник команд — [`COMMANDS.md`](COMMANDS.md).

---

## 📚 Документация

Указатель по каталогу — [`docs/README.md`](docs/README.md).

| Тема | Документ |
|---|---|
| Куда идёт проект | [`ROADMAP.md`](ROADMAP.md), [`docs/BACKLOG.md`](docs/BACKLOG.md), [`docs/IDEAS.md`](docs/IDEAS.md) |
| Что уже выпущено | [`CHANGELOG.md`](CHANGELOG.md) |
| Архитектура | [`docs/architecture/README.md`](docs/architecture/README.md), [`docs/ARCHITECTURE_SUMMARY.md`](docs/ARCHITECTURE_SUMMARY.md) |
| Развёртывание и конфигурация | [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md), [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md), [`docs/DOCKER.md`](docs/DOCKER.md) |
| Тестирование | [`docs/TESTING.md`](docs/TESTING.md), [`docs/REGRESSION_TEST_PLAN.md`](docs/REGRESSION_TEST_PLAN.md), [`docs/ARGOS.md`](docs/ARGOS.md) |
| Резервное копирование | [`docs/BACKUP.md`](docs/BACKUP.md) |
| Авторизация graph-service | [`docs/GRAPH_SERVICE_AUTH.md`](docs/GRAPH_SERVICE_AUTH.md) |
| Рекомендации | [`docs/RECOMMENDATION_ARCHITECTURE.md`](docs/RECOMMENDATION_ARCHITECTURE.md) |

---

## 🤝 Участие

Проект личный, но договорённости записаны и проверяются машинами:
[`.windsurfrules`](.windsurfrules) — единственный нормативный источник по
архитектуре, тестированию, безопасности и языковой политике.

Два правила важнее прочих. Документация обновляется тем же изменением, которое
меняет поведение, а не потом. И поиск находит кандидата, но никогда не
подтверждает находку — она подтверждается чтением контекста или исполнением.

---

## 📄 Лицензия

MIT — см. [`LICENSE`](LICENSE).

---

## 🙏 Благодарности

Three.js, Svelte, Gin, PostgreSQL с pgvector и sentence-transformers.
