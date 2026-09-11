# Развёртывание Knowledge Graph на новой машине

> Практическое руководство по запуску трёх Docker-стеков: **dev** (разработка), **personal** (личные данные) и **test** (E2E/BDD).  
> Для production, Kubernetes и CI/CD детали смотри [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md).  
> Тестирование — [`docs/TESTING.md`](docs/TESTING.md), бэкапы — [`docs/BACKUP.md`](docs/BACKUP.md), конфигурация — [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md).

---

## Содержание

- [Развёртывание Knowledge Graph на новой машине](#развёртывание-knowledge-graph-на-новой-машине)
  - [Содержание](#содержание)
  - [Что понадобится](#что-понадобится)
  - [Быстрый старт (TL;DR)](#быстрый-старт-tldr)
  - [Клонирование и ветка](#клонирование-и-ветка)
  - [Настройка `.env`](#настройка-env)
  - [Базы данных: пользователи, пароли и подключение](#базы-данных-пользователи-пароли-и-подключение)
    - [PostgreSQL](#postgresql)
      - [Если PostgreSQL уже есть (не Docker)](#если-postgresql-уже-есть-не-docker)
    - [MongoDB](#mongodb)
      - [Если MongoDB внешняя или с авторизацией](#если-mongodb-внешняя-или-с-авторизацией)
    - [Краткая сводка](#краткая-сводка)
    - [Redis](#redis)
      - [Redis с паролем (внешний инстанс)](#redis-с-паролем-внешний-инстанс)
  - [Сервисы: что и как настраивать](#сервисы-что-и-как-настраивать)
    - [Backend и Worker](#backend-и-worker)
    - [Graph service](#graph-service)
      - [Внутренний токен graph-service](#внутренний-токен-graph-service)
    - [Frontend](#frontend)
    - [nginx](#nginx)
    - [Backup scheduler](#backup-scheduler)
  - [NLP-модель и `huggingface_cache`](#nlp-модель-и-huggingface_cache)
    - [Вариант А — с интернетом](#вариант-а--с-интернетом)
    - [Вариант Б — без интернета](#вариант-б--без-интернета)
    - [Вариант В — скачать локально (не в Docker)](#вариант-в--скачать-локально-не-в-docker)
  - [Три стека: как запустить](#три-стека-как-запустить)
    - [Personal](#personal)
    - [Dev](#dev)
    - [Test](#test)
  - [Порты и URL после старта](#порты-и-url-после-старта)
  - [Проверка, что всё поднялось](#проверка-что-всё-поднялось)
  - [Первая регистрация и вход](#первая-регистрация-и-вход)
  - [Дополнительная конфигурация: OAuth, SMTP, ресурсы](#дополнительная-конфигурация-oauth-smtp-ресурсы)
    - [OAuth через Яндекс](#oauth-через-яндекс)
    - [SMTP для сброса пароля](#smtp-для-сброса-пароля)
    - [Ресурсы Docker Desktop](#ресурсы-docker-desktop)
  - [Полный чек-лист развёртывания](#полный-чек-лист-развёртывания)
    - [1. Окружение](#1-окружение)
    - [2. NLP-модель](#2-nlp-модель)
    - [3. Запуск](#3-запуск)
    - [4. Первый пользователь](#4-первый-пользователь)
    - [5. Бэкап (опционально)](#5-бэкап-опционально)
  - [Частые проблемы по сервисам](#частые-проблемы-по-сервисам)
    - [`vitest is not recognized` / фронтенд не собирается локально](#vitest-is-not-recognized--фронтенд-не-собирается-локально)
    - [NLP не стартует: `Model not found`](#nlp-не-стартует-model-not-found)
    - [Фронт пишет `Could not load knowledge-graph.config.json`](#фронт-пишет-could-not-load-knowledge-graphconfigjson)
    - [Связи в карточке заметки не отображаются](#связи-в-карточке-заметки-не-отображаются)
    - [E2E падает с `ERR_CONNECTION_REFUSED http://localhost:5173`](#e2e-падает-с-err_connection_refused-httplocalhost5173)
    - [E2E берёт старый `auth` из другого пути](#e2e-берёт-старый-auth-из-другого-пути)
    - [Бэкенд не стартует: `migrations failed`](#бэкенд-не-стартует-migrations-failed)
    - [Graph service падает при старте](#graph-service-падает-при-старте)
    - [Worker висит без обработки](#worker-висит-без-обработки)
    - [PostgreSQL: `password authentication failed`](#postgresql-password-authentication-failed)
    - [MongoDB: `connection refused`](#mongodb-connection-refused)
    - [Backup: `BACKUP_YANDEX_TOKEN not set`](#backup-backup_yandex_token-not-set)
    - [Всё стартует, но фронт пустой / белый экран](#всё-стартует-но-фронт-пустой--белый-экран)
  - [Обновление кода](#обновление-кода)
  - [Перенос Personal-данных на другую машину](#перенос-personal-данных-на-другую-машину)
    - [Через SQL-бэкап (рекомендуется)](#через-sql-бэкап-рекомендуется)
    - [Через дамп Docker-томов](#через-дамп-docker-томов)
  - [Что не надо трогать](#что-не-надо-трогать)
  - [Связанные документы](#связанные-документы)

---

## Что понадобится

- **Windows 10/11** с **WSL2** и **Docker Desktop** (интеграция WSL2 включена).
- **Git**.
- **RAM**: минимум 8 ГБ, комфортно 16 ГБ (NLP-модель + PostgreSQL + Redis + Mongo).
- **Диск**: 30–50 ГБ свободно.
- **Интернет** на первый запуск (скачка Docker-образов и NLP-модели) или заранее скопированный `huggingface_cache`.

> Для Linux/Mac команды те же, кроме путей с `C:\`; PowerShell-специфичные фрагменты заменяй на `bash`.

---

## Быстрый старт (TL;DR)

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
cp .env.example .env
# открой .env и замени JWT_SECRET, POSTGRES_PASSWORD, PERSONAL_POSTGRES_PASSWORD

# Скачать NLP-модель (первый раз)
docker compose -f docker-compose.personal.yml run --rm -e HF_HUB_OFFLINE=0 nlp-personal python - <<'PY'
from sentence_transformers import SentenceTransformer
SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')
PY

# Запустить Personal (твои данные)
docker compose -f docker-compose.personal.yml up -d --build
```

---

## Клонирование и ветка

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
```

Актуальная ветка с последними фиксами — `ai-agents`. Если нужна она:

```powershell
git checkout ai-agents
git pull origin ai-agents
```

---

## Настройка `.env`

```powershell
cp .env.example .env
```

`.env` **не коммитится** (прописан в `.gitignore`). Открой файл и измени обязательные поля:

| Переменная | Значение | Почему важно |
|------------|----------|--------------|
| `JWT_SECRET` | Длинная случайная строка, минимум 32 символа | Подписывает access/refresh токены. Если пусто — авторизация не работает. |
| `POSTGRES_PASSWORD` | Сложный пароль | Пароль dev-базы. На локалке можно оставить `change_me_in_production`. |
| `PERSONAL_POSTGRES_PASSWORD` | Сложный пароль | Пароль Personal-базы. |
| `TEST_POSTGRES_PASSWORD` | Пароль | Для изолированного тест-стека. |
| `MONGO_URL` | `mongodb://kg-mongo-personal:27017` | Подключение к MongoDB. Бэкенд читает именно `MONGO_URL`, а не `MONGODB_URL`. |
| `MONGO_DATABASE` | `knowledge_graph` | Имя MongoDB базы. |

Остальные переменные можно оставить по умолчанию.  
Если планируешь облачный бэкап, добавь `BACKUP_YANDEX_TOKEN` **в окружение**, а не в `.env` — см. [`docs/BACKUP.md`](docs/BACKUP.md) и `SEC-2` в [`docs/AI_HANDOFF.md`](docs/AI_HANDOFF.md).

---

## Базы данных: пользователи, пароли и подключение

> Этот раздел — подробно о том, как приложение подключается к PostgreSQL и MongoDB, и когда нужно создавать пользователей вручную.

### PostgreSQL

Если PostgreSQL запускается внутри Docker Compose, **нового пользователя создавать не надо** — образ PostgreSQL сам создаёт суперпользователя и базу из переменных окружения:

```env
PERSONAL_POSTGRES_USER=personal
PERSONAL_POSTGRES_PASSWORD=change_me_personal
PERSONAL_POSTGRES_DB=knowledge_personal
```

Происходящее под капотом (Docker init):

```sql
CREATE USER personal WITH PASSWORD 'change_me_personal' SUPERUSER;
CREATE DATABASE knowledge_personal OWNER personal;
```

Бэкенд подключается через переменную `DATABASE_URL` (для dev) или `PERSONAL_DATABASE_URL` (для personal), либо через compose-шаблон:

```env
PERSONAL_DATABASE_URL=postgresql://personal:change_me_personal@postgres_personal:5432/knowledge_personal?sslmode=disable
```

#### Если PostgreSQL уже есть (не Docker)

1. Создай БД и пользователя:

```sql
-- подключись к postgres как суперпользователь (например, psql -U postgres)
CREATE USER knowledge WITH PASSWORD 'твой_пароль';
CREATE DATABASE knowledge_personal OWNER knowledge;
-- dev-база
CREATE DATABASE knowledge_base OWNER knowledge;
```

2. Проверь доступ:

```powershell
psql -h 127.0.0.1 -U knowledge -d knowledge_personal -c "SELECT 1;"
```

3. Запиши в `.env`:

```env
PERSONAL_DATABASE_URL=postgresql://knowledge:твой_пароль@127.0.0.1:5432/knowledge_personal?sslmode=disable
DATABASE_URL=postgresql://knowledge:твой_пароль@127.0.0.1:5432/knowledge_base?sslmode=disable
```

4. Убедись, что в `docker-compose.personal.yml` не поднимается контейнер `postgres_personal` (закомментируй или выключи его), иначе порт `5433` займёт другой инстанс.

### MongoDB

В штатном Docker Compose MongoDB запускается **без авторизации** (по умолчанию `auth` отключён). Создавать пользователя не нужно, если:

- Mongo поднята той же `docker compose` командой;
- она не торчит наружу (порты проброшены только на `127.0.0.1` или вообще не проброшены).

Переменные окружения, которые читает бэкенд:

```env
MONGO_URL=mongodb://kg-mongo-personal:27017
MONGO_DATABASE=knowledge_graph
```

> **Важно:** в `.env.example` переменные называются `MONGODB_URL` и `MONGODB_DATABASE`, но бэкенд ожидает именно `MONGO_URL` и `MONGO_DATABASE`. Для внешней/защищённой Mongo используй `MONGO_URL` и `MONGO_DATABASE`.

#### Если MongoDB внешняя или с авторизацией

1. Создай пользователя в базе:

```javascript
// подключись через mongosh
use knowledge_graph;
db.createUser({
  user: "kg_user",
  pwd: "твой_пароль",
  roles: [
    { role: "readWrite", db: "knowledge_graph" }
  ]
});
```

2. Проверь доступ:

```powershell
mongosh "mongodb://kg_user:твой_пароль@127.0.0.1:27017/knowledge_graph?authSource=knowledge_graph" --eval "db.getName()"
```

3. Запиши в `.env`:

```env
MONGO_URL=mongodb://kg_user:твой_пароль@127.0.0.1:27017/knowledge_graph?authSource=knowledge_graph
MONGO_DATABASE=knowledge_graph
```

4. Чтобы Docker Compose не поднимал свой Mongo, закомментируй сервис `mongo` в `docker-compose.personal.yml`.

### Краткая сводка

| База | Внутри Docker Compose | Внешний инстанс |
|---|---|---|
| **PostgreSQL** | Создаётся автоматически из `.env` | Создай вручную пользователя + БД, пропиши `DATABASE_URL` / `PERSONAL_DATABASE_URL` |
| **MongoDB** | Без авторизации, пользователь не нужен | Создай вручную, пропиши `MONGO_URL` и `MONGO_DATABASE` (не `MONGODB_URL`!) |

### Redis

Redis в Docker Compose поднимается **без пароля**. Бэкенд и graph-service подключаются по `REDIS_URL`:

```env
REDIS_URL=redis:6379
PERSONAL_REDIS_URL=redis_personal:6379
```

> Для Personal используется `PERSONAL_REDIS_URL`. Если переменная пустая, graph-service падает на дефолт `redis:6379`, поэтому лучше прописать явно.

#### Redis с паролем (внешний инстанс)

1. В `redis.conf` включи:

```
requirepass твой_пароль
```

2. В `.env`:

```env
PERSONAL_REDIS_URL=redis://:твой_пароль@127.0.0.1:6379/0
REDIS_URL=redis://:твой_пароль@127.0.0.1:6379/0
```

3. Проверь:

```powershell
redis-cli -h 127.0.0.1 -a "твой_пароль" ping
```

4. Закомментируй сервис `redis` в `docker-compose.personal.yml`.

---

## Сервисы: что и как настраивать

### Backend и Worker

Бэкенд и воркер — один и тот же Go-бинарник, но с разными `CMD`:

- `kg-backend-personal` — HTTP API (`./server`).
- `kg-worker-personal` — фоновая обработка (`./worker`).

Что должен получить каждый из `.env`:

```env
# Базы
DATABASE_URL=postgresql://...        # для dev
PERSONAL_DATABASE_URL=postgresql://... # для personal
REDIS_URL=redis://...
PERSONAL_REDIS_URL=redis://...
MONGO_URL=mongodb://...
MONGO_DATABASE=knowledge_graph

# Безопасность
JWT_SECRET=твой_секрет

# Graph service
GRAPH_SERVICE_URL=http://graph-service-personal:9091
GRAPH_SERVICE_INTERNAL_TOKEN=        # если настроен, см. ниже

# NLP
NLP_SERVICE_URL=http://nlp-personal:5000
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2

# Опционально
SKIP_AUTH=false
BACKUP_ENABLED=true
```

**Проверка бэкенда:**

```powershell
curl http://127.0.0.1:18085/health
```

**Проверка воркера:**

Воркер не открывает порт. Смотри логи:

```powershell
docker logs -f kg-worker-personal
```

Ожидаемое: `worker started`, `asynq: ready`, обработка задач без `connection refused`.

### Graph service

Graph service — отдельный Go-микросервис. Он читает связи из PostgreSQL и кэширует раскладки в Redis. Чтобы запросы с JWT проходили, у него должен быть тот же `JWT_SECRET`, что и у бэкенда.

Обязательные переменные:

```env
JWT_SECRET=твой_секрет
POSTGRES_URL=postgresql://personal:change_me_personal@postgres_personal:5432/knowledge_personal?sslmode=disable
REDIS_URL=redis://redis_personal:6379/0
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
```

> На практике compose задаёт `POSTGRES_URL` из `PERSONAL_DATABASE_URL`, а `REDIS_URL` из `PERSONAL_REDIS_URL`.

**Проверка:**

```powershell
curl http://127.0.0.1:9092/health
```

**Если карточка заметки не показывает связи** — почисти кэш:

```powershell
docker exec -i kg-redis-personal redis-cli --scan --pattern "graph-service:*" | ForEach-Object { docker exec -i kg-redis-personal redis-cli del $_ }
```

#### Внутренний токен graph-service

`GRAPH_SERVICE_INTERNAL_TOKEN` — опциональный shared secret между бэкендом и graph-service. Если задан:

- Бэкенд зовёт graph-service с заголовком `X-Internal-Auth`.
- Graph-service проверяет его и доверяет `X-User-Id`.
- В `.env` фронтенда он не нужен, потому что браузер не ходит напрямую в graph-service.

Для старта можно оставить пустым. В production — задать 32+ случайных символа.

### Frontend

Фронтенд собирается в Docker и отдаётся через nginx. На этапе сборки bake-ятся переменные:

```env
VITE_API_URL=/api
VITE_GRAPH_SERVICE_URL=/graph-service
VITE_API_TARGET=http://backend_personal:8080
GRAPH_SERVICE_URL=http://graph-service-personal:9091
GRAPH_SERVICE_INTERNAL_TOKEN=        # только для SSR
```

> Обычно не требует правки в `.env`. Меняй, только если поднимаешь фронтенд локально (`npm run dev`) или меняешь URL бэкенда.

**Проверка:**

```powershell
curl -s http://127.0.0.1:18084 | head
```

Ожидаем: HTML с `__data`.

### nginx

nginx служит единым шлюзом. Конфиг монтируется из `nginx.personal.conf`. Меняй его, только если:

- переносишь стек на другие порты;
- добавляешь новые `location` / заголовки безопасности.

Порт `18082` — API/бэкенд.  
Порт `18084` — фронтенд.

**Проверка:**

```powershell
curl http://127.0.0.1:18082/health
curl http://127.0.0.1:18084
```

### Backup scheduler

По умолчанию выключен в `docker-compose.personal.yml`? Проверь, включён ли сервис. Если включён, он бэкапит Postgres по cron.

Базовые переменные:

```env
BACKUP_ENABLED=true
BACKUP_LOCAL_PATH=./backups
BACKUP_SCHEDULE=0 23 * * 0
BACKUP_RETENTION_DAYS=14
```

Для облачного бэкапа на Yandex:

```powershell
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token"
docker compose -f docker-compose.personal.yml up -d backup_scheduler
```

Локальный бэкап вручную:

```powershell
.\scripts\devops\backup-personal.ps1 -Mode daily
```

---

## NLP-модель и `huggingface_cache`

NLP-сервис работает **офлайн** (`HF_HUB_OFFLINE=1`) и читает модель из bind-mount `./huggingface_cache:/root/.cache/huggingface`. При первом старте папка пустая.

### Вариант А — с интернетом

Скачать модель через тот же контейнер:

```powershell
docker compose -f docker-compose.personal.yml run --rm -e HF_HUB_OFFLINE=0 nlp-personal python - <<'PY'
from sentence_transformers import SentenceTransformer
SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')
PY
```

После этого в `./huggingface_cache` появится папка `models--sentence-transformers--paraphrase-multilingual-MiniLM-L12-v2`.

### Вариант Б — без интернета

Скопируй папку `huggingface_cache` со старой машины в корень репозитория.

### Вариант В — скачать локально (не в Docker)

```powershell
pip install sentence-transformers
python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')"
```

Затем найди кэш HuggingFace на своей машине (`%USERPROFILE%\.cache\huggingface` на Windows) и скопируй его в `./huggingface_cache`.

---

## Три стека: как запустить

### Personal

Личный стенд. Данные живут в именованных томах `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal`. **Не удаляй их** без бэкапа.

```powershell
docker compose -f docker-compose.personal.yml up -d --build
```

Флаг `--build` нужен, когда код меняется. В первый раз он соберёт все образы.  
Следующий раз можно без `--build`:

```powershell
docker compose -f docker-compose.personal.yml up -d
```

Проверить:

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}"
```

### Dev

Общий dev-стенд. Его данные — в томах `postgres_data`, `redis_data`, `mongodb_data`.

```powershell
docker compose up -d --build
```

Dev и Personal можно поднимать одновременно — порты не пересекаются. Но на слабом железе лучше по очереди.

### Test

Изолированный стек для E2E/BDD/Playwright. Перед ним останови Dev и Personal, чтобы не было конфликтов портов.

```powershell
# Остановить другие стеки, если подняты
docker compose down
docker compose -f docker-compose.personal.yml down

# Поднять и засеять
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42 -PublicPercent 50
```

Playwright:

```powershell
cd frontend
$env:FRONTEND_URL = "http://127.0.0.1:3002"
$env:BACKEND_URL = "http://127.0.0.1:18083"
npx playwright test --project=chromium-skip-auth
```

Потом:

```powershell
.\scripts\testing\stop-test.ps1
```

---

## Порты и URL после старта

| Стек | UI / Фронтенд | API-шлюз (nginx) | Backend напрямую | Graph service | PostgreSQL | Redis | MongoDB | NLP |
|------|---------------|------------------|------------------|---------------|------------|-------|---------|-----|
| **Personal** | http://127.0.0.1:18084 | http://127.0.0.1:18082 | http://127.0.0.1:18085 | http://127.0.0.1:9092 | 5433 | 16380 | 27018 | 5001 |
| **Dev** | http://127.0.0.1:18081 | http://127.0.0.1:18080 | http://127.0.0.1:9000 | http://127.0.0.1:9091 | 15432 | 16379 | 27017 | 5001 |
| **Test** | http://127.0.0.1:3002 | http://127.0.0.1:18086 | http://127.0.0.1:18083 | http://127.0.0.1:19091 | 15434 | 16381 | 27019 | 15002 |

Браузер всегда ходит через nginx (для Personal — `18084`). Бэкенд напрямую нужен только для отладки.

---

## Проверка, что всё поднялось

```powershell
# Список контейнеров и их статус
docker ps --format "table {{.Names}}\t{{.Status}}"

# Health backend / graph / nlp
curl http://127.0.0.1:18085/health
curl http://127.0.0.1:9092/health
curl http://127.0.0.1:5001/health

# Логи конкретного сервиса
docker logs -f kg-backend-personal
docker logs -f kg-graph-service-personal
docker logs -f kg-nlp-personal

# Статус через compose
docker compose -f docker-compose.personal.yml ps
```

Все сервисы должны быть `Up` и `(healthy)`.

---

## Первая регистрация и вход

1. Открой `http://127.0.0.1:18084`.
2. Зарегистрируйся через `/register` или войди, если учётка уже есть.
3. Создай первую заметку. Через несколько секунд NLP посчитает эмбеддинг и появятся рекомендации.

> Если `SKIP_AUTH=true` в `.env`, вход не требуется, но это режим только для локального теста.

---

## Дополнительная конфигурация: OAuth, SMTP, ресурсы

### OAuth через Яндекс

Если хочешь вход через Яндекс:

1. Создай приложение в [Yandex OAuth](https://oauth.yandex.ru/).
2. В `.env`:

```env
YANDEX_CLIENT_ID=твой_client_id
YANDEX_CLIENT_SECRET=твой_client_secret
PKCE_ENABLED=true
```

3. Пересобери backend: `docker compose -f docker-compose.personal.yml up -d --build backend_personal`.
4. URL редиректа в Яндексе: `http://127.0.0.1:18084/auth/yandex/callback`.

### SMTP для сброса пароля

Для production и для Personal-стека с внешним доступом:

```env
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=587
SMTP_USER=your@yandex.ru
SMTP_PASSWORD=app_password
SMTP_FROM=your@yandex.ru
```

> Gmail требует "App Password"; для теста можно оставить пустым — сброс пароля не будет работать.

### Ресурсы Docker Desktop

Открой **Settings → Resources → WSL integration**. Рекомендуется:

- **Memory**: минимум 6 ГБ, лучше 8–12 ГБ.
- **Swap**: 1–2 ГБ.
- **Disk image location**: на SSD.
- **WSL integration**: включена для дистрибутива, из которого запускаешь Docker.

Если стек падает по OOM, увеличь лимиты или снизь `RECOMMENDATION_TOP_N` и `GRAPH_MAX_LIMIT` в `knowledge-graph.config.json`.

---

## Полный чек-лист развёртывания

Распечатай/скопируй и отмечай галочками:

### 1. Окружение

- [ ] Docker Desktop + WSL2 установлены и запущены.
- [ ] Репозиторий склонирован: `git clone ... && cd knowledge-graph && git checkout ai-agents`.
- [ ] `.env` создан из `.env.example`.
- [ ] `JWT_SECRET` заполнен (32+ символов).
- [ ] Пароли PostgreSQL (`POSTGRES_PASSWORD`, `PERSONAL_POSTGRES_PASSWORD`) заполнены.
- [ ] `MONGO_URL` и `MONGO_DATABASE` заполнены (если внешняя Mongo).
- [ ] `REDIS_URL` / `PERSONAL_REDIS_URL` заполнены (если внешний Redis).
- [ ] `CORS_ALLOWED_ORIGINS` включает `http://127.0.0.1:18084` и `http://localhost:18084`.

### 2. NLP-модель

- [ ] Папка `huggingface_cache` на месте.
- [ ] Модель `paraphrase-multilingual-MiniLM-L12-v2` скачана (папка `models--sentence-transformers--...` внутри).

### 3. Запуск

- [ ] `docker compose -f docker-compose.personal.yml up -d --build`.
- [ ] Все контейнеры `Up (healthy)`:
  - `kg-postgres-personal`
  - `kg-redis-personal`
  - `kg-mongo-personal`
  - `kg-nlp-personal`
  - `kg-graph-service-personal`
  - `kg-backend-personal`
  - `kg-worker-personal`
  - `kg-frontend-personal`
  - `kg-nginx-personal`
- [ ] `curl http://127.0.0.1:18085/health` → OK.
- [ ] `curl http://127.0.0.1:9092/health` → OK.
- [ ] `curl http://127.0.0.1:5001/health` → OK.
- [ ] `curl http://127.0.0.1:18084` → HTML.

### 4. Первый пользователь

- [ ] Открыт `http://127.0.0.1:18084`.
- [ ] Регистрация прошла, в базе появилась запись в `users`.
- [ ] Создана первая заметка.
- [ ] Через 5–30 секунд в заметке появились рекомендации.

### 5. Бэкап (опционально)

- [ ] `BACKUP_ENABLED=true`.
- [ ] Папка `./backups` существует.
- [ ] Тестовый ручной бэкап: `.\scripts\devops\backup-personal.ps1`.

---

## Частые проблемы по сервисам

### `vitest is not recognized` / фронтенд не собирается локально

Фронтенд собирается внутри Docker. Локальный `npm install` не обязателен. Для тестов внутри контейнера:

```powershell
docker exec -it kg-frontend-personal sh
npm run build
```

### NLP не стартует: `Model not found`

- Проверь `huggingface_cache/models--sentence-transformers--...`.
- Посмотри логи: `docker logs -f kg-nlp-personal`.
- Если папка пустая — скачай модель заново командой из раздела [NLP-модель](#nlp-модель-и-huggingface_cache).

### Фронт пишет `Could not load knowledge-graph.config.json`

- Убедись, что `knowledge-graph.config.json` лежит в корне репозитория.
- Пересобери образ: `docker compose -f docker-compose.personal.yml up -d --build frontend-personal`.

### Связи в карточке заметки не отображаются

Graph-service кэширует раскладки в Redis. Если после обновления видишь пустоту:

```powershell
docker exec -i kg-redis-personal redis-cli --scan --pattern "graph-service:*" | ForEach-Object { docker exec -i kg-redis-personal redis-cli del $_ }
```

### E2E падает с `ERR_CONNECTION_REFUSED http://localhost:5173`

Playwright берёт `baseURL` из `FRONTEND_URL`. На Windows `localhost` резолвится в `::1`, поэтому используй `127.0.0.1`:

```powershell
$env:FRONTEND_URL = "http://127.0.0.1:3002"
```

### E2E берёт старый `auth` из другого пути

Запускай из `frontend/`. Относительные пути `tests/setup/.auth/` разрешаются от каталога процесса.

### Бэкенд не стартует: `migrations failed`

- Проверь `DATABASE_URL` / `PERSONAL_DATABASE_URL`.
- Убедись, что Postgres отвечает:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT 1;"
```

- Если схема сильно устарела, можно удалить только dev-базу, но **НЕ** Personal-данные.

### Graph service падает при старте

- Проверь `POSTGRES_URL` и `REDIS_URL` внутри контейнера:

```powershell
docker logs -f kg-graph-service-personal
```

- Частая ошибка: `JWT_SECRET` пустой → `unauthorized` на все приватные endpoint.

### Worker висит без обработки

- Проверь логи: `docker logs -f kg-worker-personal`.
- Убедись, что Redis видит очередь:

```powershell
docker exec -i kg-redis-personal redis-cli llen asynq:{default}
```

Если очередь растёт, а воркер не забирает — проверь `ASYNQ_CONCURRENCY` и `REDIS_URL`.

### PostgreSQL: `password authentication failed`

- В `.env` и в `docker-compose.personal.yml` разные пароли? Они должны совпадать.
- Если менял пароль в `.env`, но том `pgdata_personal` уже был создан со старым, нужно либо поменять пароль через psql, либо удалить том (с бэкапом) и пересоздать.

### MongoDB: `connection refused`

- Убедись, что `MONGO_URL` указывает на правильный хост. Внутри Docker — `mongodb://kg-mongo-personal:27017`, снаружи — `mongodb://127.0.0.1:27018`.
- Проверь, что Mongo поднялся: `docker ps | grep mongo`.

### Backup: `BACKUP_YANDEX_TOKEN not set`

- Либо задай токен в окружении, либо отключи облако:

```env
BACKUP_CLOUD_ENABLED=false
```

- Проверь, что папка `backups` доступна контейнеру:

```powershell
docker exec -i kg-backup-scheduler ls -la /backups
```

### Всё стартует, но фронт пустой / белый экран

- Логи фронтенда: `docker logs -f kg-frontend-personal`.
- Проверь, что `knowledge-graph.config.json` смонтирован:

```powershell
docker exec -i kg-frontend-personal cat /app/knowledge-graph.config.json | head
```

- Пересобери фронтенд: `docker compose -f docker-compose.personal.yml up -d --build frontend-personal`.

---

## Обновление кода

```powershell
git pull origin ai-agents

# Пересобрать и перезапустить Personal
docker compose -f docker-compose.personal.yml down
docker compose -f docker-compose.personal.yml up -d --build
```

Данные в томах сохранятся.

---

## Перенос Personal-данных на другую машину

### Через SQL-бэкап (рекомендуется)

```powershell
.\scripts\devops\backup-personal.ps1 -Mode daily
```

Появится `backups/backup-personal-daily-<timestamp>.sql.gz`. Скопируй на новую машину и восстанови:

```powershell
docker cp backup-personal-daily-....sql.gz kg-postgres-personal:/tmp/
docker exec -i kg-postgres-personal gunzip -c /tmp/backup-personal-daily-....sql.gz | psql -U personal -d knowledge_personal
```

### Через дамп Docker-томов

```powershell
# Остановить стек
docker compose -f docker-compose.personal.yml down

# Забэкапить тома
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/pgdata_personal.tar /data
docker run --rm -v redisdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/redisdata_personal.tar /data
docker run --rm -v mongodbdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/mongodbdata_personal.tar /data
```

На новой машине — развернуть:

```powershell
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine sh -c "cd / && tar xvf /backup/pgdata_personal.tar"
# аналогично для redis и mongo
```

---

## Что не надо трогать

- **Не удаляй** Personal-тома `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal` без бэкапа.
- **Не коммить** `.env` и `huggingface_cache`.
- **Не запускай** E2E/BDD против Personal-стека: для этого есть изолированный test-стек.
- **Не правь** `docker-compose.personal.yml`, если не уверен в портах и томах — сначала прочитай [`docs/DOCKER.md`](docs/DOCKER.md).

---

## Связанные документы

- [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md) — production и Kubernetes.
- [`docs/TESTING.md`](docs/TESTING.md) — тест-стек, регрессия, Playwright.
- [`docs/BACKUP.md`](docs/BACKUP.md) — бэкапы Personal.
- [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md) — переменные среды и `knowledge-graph.config.json`.
- [`docs/DOCKER.md`](docs/DOCKER.md) — карта портов и томов.
- [`docs/GRAPH_SERVICE_AUTH.md`](docs/GRAPH_SERVICE_AUTH.md) — авторизация graph-service.
