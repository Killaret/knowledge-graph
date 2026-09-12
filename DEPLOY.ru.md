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
  - [Создание первого пользователя и администратора](#создание-первого-пользователя-и-администратора)
    - [Через фронтенд](#через-фронтенд)
    - [Через API (headless / автоматизация)](#через-api-headless--автоматизация)
    - [Назначить роль admin](#назначить-роль-admin)
    - [Первый пользователь для тест-стека](#первый-пользователь-для-тест-стека)
  - [Дополнительная конфигурация: OAuth, SMTP, ресурсы](#дополнительная-конфигурация-oauth-smtp-ресурсы)
    - [OAuth через Яндекс](#oauth-через-яндекс)
    - [SMTP для сброса пароля](#smtp-для-сброса-пароля)
    - [Ресурсы Docker Desktop](#ресурсы-docker-desktop)
    - [CI/CD и GitHub Actions](#cicd-и-github-actions)
    - [Production: HTTPS и SSL](#production-https-и-ssl)
  - [Полный аннотированный `.env` (шаблон)](#полный-аннотированный-env-шаблон)
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
  - [Работа внутри Docker-сети](#работа-внутри-docker-сети)
    - [Узнать имя сети](#узнать-имя-сети)
    - [Зайти внутрь контейнера](#зайти-внутрь-контейнера)
    - [Выполнить SQL-запрос в PostgreSQL](#выполнить-sql-запрос-в-postgresql)
    - [Redis](#redis-1)
    - [MongoDB](#mongodb-1)
    - [Выполнить запрос с хоста, но через опубликованный порт](#выполнить-запрос-с-хоста-но-через-опубликованный-порт)
    - [Запустить временный контейнер внутри сети](#запустить-временный-контейнер-внутри-сети)
    - [Поднять отладочный контейнер с клиентами](#поднять-отладочный-контейнер-с-клиентами)
    - [Скопировать файл в контейнер или из него](#скопировать-файл-в-контейнер-или-из-него)
    - [Обратиться к бэкенду/граф-сервису изнутри сети](#обратиться-к-бэкендуграф-сервису-изнутри-сети)
  - [Обновление кода](#обновление-кода)
  - [Перенос Personal-данных на другую машину](#перенос-personal-данных-на-другую-машину)
    - [Через SQL-бэкап (рекомендуется)](#через-sql-бэкап-рекомендуется)
    - [Через дамп Docker-томов](#через-дамп-docker-томов)
  - [Обновление между релизами](#обновление-между-релизами)
    - [Безопасный процесс обновления](#безопасный-процесс-обновления)
    - [Миграции базы данных](#миграции-базы-данных)
    - [Если миграция не применилась](#если-миграция-не-применилась)
    - [Обновление с релиза на релиз (main)](#обновление-с-релиза-на-релиз-main)
  - [Конфликты портов](#конфликты-портов)
    - [Windows](#windows)
    - [Linux / macOS](#linux--macos)
    - [Что может занимать порт](#что-может-занимать-порт)
    - [Быстрое решение: сменить порты в `.env`](#быстрое-решение-сменить-порты-в-env)
    - [Если запущены сразу несколько стеков](#если-запущены-сразу-несколько-стеков)
    - [Проверка](#проверка)
  - [Бэкап, восстановление и проверка](#бэкап-восстановление-и-проверка)
    - [Рекомендуемая стратегия](#рекомендуемая-стратегия)
    - [SQL-бэкап](#sql-бэкап)
    - [Восстановление PostgreSQL из SQL-бэкапа](#восстановление-postgresql-из-sql-бэкапа)
    - [Дамп Docker-томов](#дамп-docker-томов)
    - [Восстановление из дампа томов](#восстановление-из-дампа-томов)
    - [Smoke-тест после восстановления](#smoke-тест-после-восстановления)
    - [Хранение бэкапов](#хранение-бэкапов)
  - [WSL2 и Docker Desktop: ловушки Windows](#wsl2-и-docker-desktop-ловушки-windows)
    - [Почему `localhost` не работает, а `127.0.0.1` — да](#почему-localhost-не-работает-а-127001--да)
    - [Диск WSL2 растёт](#диск-wsl2-растёт)
    - [Docker Desktop не перезапускает WSL](#docker-desktop-не-перезапускает-wsl)
    - [Docker Desktop использует Hyper-V или WSL2](#docker-desktop-использует-hyper-v-или-wsl2)
    - [Антивирус и файервол](#антивирус-и-файервол)
    - [Пути Windows ↔ WSL](#пути-windows--wsl)
    - [`wsl --shutdown` безопасный](#wsl---shutdown-безопасный)
  - [Журнал логов: типичные сообщения и что они значат](#журнал-логов-типичные-сообщения-и-что-они-значат)
    - [Как смотреть логи](#как-смотреть-логи)
    - [Бэкенд](#бэкенд)
    - [Graph service](#graph-service-1)
    - [Worker](#worker)
    - [NLP](#nlp)
    - [PostgreSQL](#postgresql-1)
    - [Redis](#redis-2)
    - [MongoDB](#mongodb-2)
  - [Харднинг и безопасность](#харднинг-и-безопасность)
    - [Секреты](#секреты)
    - [SKIP\_AUTH](#skip_auth)
    - [CORS](#cors)
    - [HTTPS](#https)
    - [Сетевой доступ](#сетевой-доступ)
    - [Docker](#docker)
    - [Бэкапы и целостность](#бэкапы-и-целостность)
  - [Перформанс-тюнинг](#перформанс-тюнинг)
    - [Docker Desktop](#docker-desktop)
    - [Лимиты в compose](#лимиты-в-compose)
    - [PostgreSQL](#postgresql-2)
    - [Redis](#redis-3)
    - [NLP](#nlp-1)
    - [Frontend](#frontend-1)
    - [Мониторинг нагрузки](#мониторинг-нагрузки)
    - [Признаки, что не хватает ресурсов](#признаки-что-не-хватает-ресурсов)
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

## Создание первого пользователя и администратора

### Через фронтенд

1. Открой `http://127.0.0.1:18084`.
2. Перейди `/register`, введи email, имя и пароль.
3. Войди через `/login`.
4. Создай первую заметку. Через несколько секунд NLP посчитает эмбеддинг и появятся рекомендации.

### Через API (headless / автоматизация)

```powershell
$body = @{
    email = "admin@example.com"
    name = "Admin"
    password = "StrongPass123!"
} | ConvertTo-Json

Invoke-RestMethod -Method POST -Uri http://127.0.0.1:18082/api/v1/auth/register -ContentType "application/json" -Body $body
```

После регистрации залогинься:

```powershell
$login = @{
    email = "admin@example.com"
    password = "StrongPass123!"
} | ConvertTo-Json

$response = Invoke-RestMethod -Method POST -Uri http://127.0.0.1:18082/api/v1/auth/login -ContentType "application/json" -Body $login -SessionVariable s
```

### Назначить роль admin

Регистрация по умолчанию даёт роль `user`. Роли создаются миграциями: `admin`, `user`, `guest`.

1. Найди ID нужной роли:

```sql
SELECT id, name FROM user_roles;
```

2. Найди ID пользователя:

```sql
SELECT id, email FROM users;
```

3. Обнови роль:

```sql
UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'admin') WHERE email = 'admin@example.com';
```

> Выполняй через `psql` под контейнером:
> ```powershell
> docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'admin') WHERE email = 'admin@example.com';"
> ```

### Первый пользователь для тест-стека

Для тестов используй `cmd/seed`:

```powershell
$env:APP_ENV = "test"
$env:SEED_TEST_USER_PASSWORD = "TestPassword123!"
docker compose -f docker-compose.test.yml run --rm seed
```

Создаётся пользователь `testuser` с паролем `TestPassword123!`.

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

### CI/CD и GitHub Actions

Проект проверяется в `.github/workflows/main.yml` (на пуш в `main`) и `_core-checks.yml` (reusable). Чтобы CI работал:

1. **Репозиторий на GitHub** должен быть публичным или иметь GitHub Actions.
2. **Секреты** в `Settings → Secrets → Actions`:

| Секрет | Зачем |
|---|---|
| `JWT_SECRET` | Для тестового бэкенда и миграций. Можно любой 32+ символа. |
| `GRAPH_SERVICE_INTERNAL_TOKEN` | Для E2E-тестов (если настроен). |

3. **Ветки**:
   - `main` — production, CI/CD, релиз.
   - `ai-agents` — Devin/Windsurf.
   - `feature/...` — мелкие правки.

4. **Что CI проверяет:**
   - `go test ./...` в backend.
   - `npm run test:unit`, `npm run lint` во frontend.
   - `golangci-lint`.
   - Проверку миграций на `pgvector/pgvector:pg16`.
   - Playwright и Cucumber с изолированным PostgreSQL/Redis/Mongo.

5. **Локальная проверка перед пушем:**

```powershell
cd backend  && go test ./...
cd frontend && npm run lint
cd ..       && .\scripts\ci\check-stacks-identity.ps1
```

### Production: HTTPS и SSL

Локальный `nginx.personal.conf` работает по HTTP. Для production с HTTPS:

1. Получи сертификат. Пример с Let's Encrypt (на хосте, не внутри Docker):

```bash
sudo certbot certonly --standalone -d example.com
```

2. Создай `nginx.production.conf` на основе `nginx.personal.conf` и добавь:

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_protocols TLSv1.3;
    ssl_prefer_server_ciphers on;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header Referrer-Policy strict-origin-when-cross-origin;

    # location /api/... — скопируй из nginx.personal.conf
}

server {
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}
```

3. Пропиши `CORS_ALLOWED_ORIGINS=https://example.com` в `.env`.
4. Пересобери nginx-контейнер с production-конфигом.

> Для локалки HTTPS не обязателен. Если нужен только внутри домашней сети, можно самоподписанный сертификат, но браузер будет ругаться.

---

## Полный аннотированный `.env` (шаблон)

> Единый источник правды по переменным — файл `.env.example` в корне.
> Ниже полный шаблон с английскими комментариями из `.env.example`. Замени `change_me_*`, `your_*` и пустые значения на свои. Для Personal-стека ключевые блоки: **Database Configuration**, **Personal Development Environment**, **Redis**, **MongoDB**, **NLP Service**, **Authentication**, **OAuth**, **SMTP**, **Cloud Backup**.

<details>
<summary>Развернуть полный `.env`</summary>

```env
# Knowledge Graph Environment Configuration
# Copy this file to .env and fill in your actual values
# DO NOT commit .env files with real credentials!

# Database Configuration
POSTGRES_USER=kb_user
POSTGRES_PASSWORD=change_me_in_production
POSTGRES_DB=knowledge_base
DATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable

# Personal Development Environment
PERSONAL_POSTGRES_USER=personal
PERSONAL_POSTGRES_PASSWORD=change_me_personal
PERSONAL_POSTGRES_DB=knowledge_personal

# Test Stack (isolated E2E/BDD testing via docker-compose.test.yml)
TEST_POSTGRES_USER=kb_user
TEST_POSTGRES_PASSWORD=kb_password
TEST_POSTGRES_DB=knowledge_test
#APP_ENV=development      # development | test | production
#SKIP_AUTH=true           # ONLY allowed when APP_ENV=test, NEVER for production
#SEED_TEST_USER_PASSWORD=TestPassword123!  # Used by cmd/seed to create the test user when APP_ENV=test
#JWT_SECRET=              # REQUIRED — set a strong random secret for the test stack

# Redis Configuration
REDIS_URL=redis:6379
PERSONAL_REDIS_URL=redis_personal:6379

# NLP Service
NLP_SERVICE_URL=http://nlp:5000
PERSONAL_NLP_SERVICE_URL=http://nlp-personal:5000

# API Configuration
API_PORT=8080
PERSONAL_API_PORT=8081

# CORS Configuration
# Comma-separated list of allowed origins (e.g., http://localhost:3000,https://example.com)
# If not set, defaults to localhost origins for development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:5173,http://localhost:18080,http://localhost:18081,http://localhost:18082,http://localhost:18084,http://127.0.0.1:3000,http://127.0.0.1:3001,http://127.0.0.1:5173,http://127.0.0.1:18080,http://127.0.0.1:18081,http://127.0.0.1:18082,http://127.0.0.1:18084
# Allowed HTTP methods (default: GET,POST,PUT,DELETE,OPTIONS)
#CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
# Allowed headers (default: Content-Type,Authorization)
#CORS_ALLOWED_HEADERS=Content-Type,Authorization
# Preflight cache duration in seconds (default: 86400 = 24 hours)
#CORS_MAX_AGE=86400

# Frontend
VITE_API_URL=http://localhost:8080
PERSONAL_VITE_API_URL=http://localhost:8081

# =============================================================================
# Optional Configuration Overrides (env > json > default)
# These variables override values from knowledge-graph.config.json
# Uncomment and modify only if you need to change defaults
# =============================================================================

# Server Configuration
#SERVER_PORT=8080
#SERVER_RATE_LIMIT_ENABLED=false
#SERVER_RATE_LIMIT_REQUESTS=1000
#SERVER_RATE_LIMIT_WINDOW_SECONDS=60

# Database Retry Configuration
#DATABASE_RETRY_MAX_ATTEMPTS=3
#DATABASE_RETRY_DELAY_SECONDS=5
#MIGRATIONS_FAIL_ON_ERROR=false

# PostgreSQL Connection Pool (Backend)
#POSTGRES_MAX_OPEN_CONNS=25
#POSTGRES_MAX_IDLE_CONNS=5
#POSTGRES_CONN_MAX_LIFETIME_MINUTES=5
#POSTGRES_CONN_MAX_IDLE_TIME_MINUTES=1

# PostgreSQL Connection Pool (Graph Service)
#GRAPH_POSTGRES_MAX_CONNS=10
#GRAPH_POSTGRES_MIN_CONNS=2
#GRAPH_POSTGRES_MAX_CONN_LIFETIME_MINUTES=5
#GRAPH_POSTGRES_MAX_CONN_IDLE_TIME_MINUTES=1
#GRAPH_POSTGRES_HEALTH_CHECK_PERIOD_MINUTES=1

# Graph Service Integration
#GRAPH_SERVICE_URL=http://graph-service:9091
#GRAPH_SERVICE_INTERNAL_TOKEN=     # REQUIRED for backend→graph-service private calls
#GRAPH_SERVICE_TRUST_USER_HEADER=false  # Enable only on an internal network whose public proxy strips X-User-Id

# Redis Connection Pool (Both Services)
#REDIS_POOL_SIZE=10
#REDIS_MIN_IDLE_CONNS=3
#REDIS_MAX_CONN_AGE_MINUTES=5
#REDIS_POOL_TIMEOUT_SECONDS=4
#REDIS_IDLE_TIMEOUT_MINUTES=5

# Search Configuration
#SEARCH_FALLBACK_TO_ILIKE=true

# Graph API Limits
#GRAPH_DEFAULT_LIMIT=100
#GRAPH_MAX_LIMIT=1000
#GRAPH_LINK_DEFAULT_LIMIT=500
#GRAPH_LINK_MAX_LIMIT=5000
#GRAPH_LOAD_DEPTH=2

# Pagination
#PAGINATION_DEFAULT_LIMIT=100
#PAGINATION_MAX_LIMIT=300

# Recommendation Algorithm Parameters
#RECOMMENDATION_ALPHA=0.5
#RECOMMENDATION_BETA=0.5
#RECOMMENDATION_GAMMA=0.2
#RECOMMENDATION_DEPTH=3
#RECOMMENDATION_DECAY=0.5
#RECOMMENDATION_TOP_N=50
#RECOMMENDATION_CACHE_TTL_SECONDS=300
#RECOMMENDATION_TASK_DELAY_SECONDS=5
#RECOMMENDATION_BATCH_RATE_LIMIT=10
#RECOMMENDATION_FALLBACK_ENABLED=true
#RECOMMENDATION_FALLBACK_TTL_SECONDS=3600
#RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true
#RECOMMENDATION_KEYWORD_ENABLED=true
#RECOMMENDATION_KEYWORD_SIMILARITY_METHOD=jaccard
#RECOMMENDATION_KEYWORD_TVERSKY_ALPHA=0.5
#RECOMMENDATION_KEYWORD_TVERSKY_BETA=0.5

# BFS Algorithm Settings
#BFS_AGGREGATION=max
#BFS_NORMALIZE=true

# Embedding Configuration
#EMBEDDING_SIMILARITY_LIMIT=30

# Asynq Queue Configuration
#ASYNQ_CONCURRENCY=10
#ASYNQ_QUEUE_DEFAULT=1
#ASYNQ_QUEUE_MAX_LEN=10000

# Authentication Configuration (REQUIRED)
#JWT_SECRET=
#JWT_ACCESS_TTL_SECONDS=900
#JWT_REFRESH_TTL_SECONDS=604800
#ARGON2_TIME=3
#ARGON2_MEMORY=65536
#ARGON2_THREADS=4
#API_KEY_ENABLED=true
#STATIC_API_KEY=kg-admin-2024-secret-key
#SKIP_AUTH=true  # ONLY allowed when APP_ENV=test, NEVER for production

# OAuth Configuration (Yandex)
#YANDEX_CLIENT_ID=
#YANDEX_CLIENT_SECRET=
#PKCE_ENABLED=true
#PKCE_CODE_CHALLENGE_LENGTH=128
# When PKCE is enabled, the OAuth flow uses the S256 method (RFC 7636).

# SMTP Configuration (Password Reset)
#SMTP_HOST=
#SMTP_PORT=587
#SMTP_USER=
#SMTP_PASSWORD=
#SMTP_FROM=noreply@example.com
#PASSWORD_RESET_TTL_SECONDS=900

# Password Policy
#PASSWORD_POLICY_MIN_LENGTH=10
#PASSWORD_POLICY_REQUIRE_UPPER=true
#PASSWORD_POLICY_REQUIRE_LOWER=true
#PASSWORD_POLICY_REQUIRE_DIGIT=true
#PASSWORD_POLICY_REQUIRE_SPECIAL=true

# =============================================================================
# MongoDB Configuration
# =============================================================================
# Backend reads MONGO_URL and MONGO_DATABASE (not MONGODB_*).
#MONGO_URL=mongodb://localhost:27017
#MONGO_DATABASE=knowledge_graph

# =============================================================================
# Graph Service Configuration
# =============================================================================
#GRAPH_SERVICE_GRPC_PORT=9090
#GRAPH_SERVICE_HTTP_PORT=9091
#GRAPH_SERVICE_FULL_LIMIT=1000
#GRAPH_SERVICE_DEFAULT_DEPTH=2
#GRAPH_SERVICE_EVENT_CHANNEL=graph:events

# Graph Service Cache Configuration
#GRAPH_SERVICE_CACHE_NOTE_LAYOUT_TTL_SECONDS=300
#GRAPH_SERVICE_CACHE_FULL_LAYOUT_TTL_SECONDS=300
#GRAPH_SERVICE_CACHE_DELTA_TTL_SECONDS=60

# Graph Service Layout Configuration
#GRAPH_SERVICE_LAYOUT_2D_RADIUS=100.0
#GRAPH_SERVICE_LAYOUT_3D_RADIUS=120.0
#GRAPH_SERVICE_LAYOUT_3D_Z_STEP=5.0
#GRAPH_SERVICE_LAYOUT_DEFAULT_NODE_SIZE=1.0

# Graph Service Streaming Configuration
#GRAPH_SERVICE_STREAM_CHUNK_SIZE=100

# Graph Service Event Processing
#GRAPH_SERVICE_EVENT_TRACKING_TTL_HOURS=24
#GRAPH_SERVICE_UNPROCESSED_EVENT_CHECK_INTERVAL_MINUTES=5

# =============================================================================
# NLP Service Configuration
# =============================================================================
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
#NLP_MAX_TEXT_LENGTH=10000

# =============================================================================
# Frontend Configuration
# =============================================================================
# Frontend Test Configuration
#FRONTEND_TEST_DEBOUNCE_TIMEOUT_MS=300
#FRONTEND_TEST_MAX_RETRY_COUNT=3
#FRONTEND_TEST_MOCK_GOTO_DELAY_MS=0

# Frontend Graph Configuration
#FRONTEND_GRAPH_2D_MAX_NODES=500
#FRONTEND_GRAPH_2D_SHADOWS_THRESHOLD=100
#FRONTEND_GRAPH_3D_MAX_NODES=500

# Frontend API Configuration
#FRONTEND_API_DEFAULT_LIMIT=100
#FRONTEND_API_LINK_LIMIT=0

# Frontend Achievements Configuration
#FRONTEND_ACHIEVEMENTS_POLL_INTERVAL_MS=7000

# =============================================================================
# Cloud Backup Configuration (Optional)
# Configure cloud backups (Yandex.Disk)
# =============================================================================
#BACKUP_CLOUD_ENABLED=false
#BACKUP_CLOUD_PROVIDER=yandex  # Options: yandex
#BACKUP_LOCAL_PATH=./backups
#BACKUP_SCHEDULE=0 23 * * 0
#BACKUP_RETENTION_DAYS=14
#BACKUP_DRAFT_TTL_HOURS=168

# Yandex.Disk configuration (OAuth token)
# CRITICAL: Set this in production to enable cloud backups
# SECURITY: Do NOT store the actual token in this file!
# Set BACKUP_YANDEX_TOKEN as an environment variable instead:
#
# Windows PowerShell:
#   $env:BACKUP_YANDEX_TOKEN = "your_oauth_token"
#   docker-compose -f docker-compose.personal.yml up -d backup_scheduler
#
# Linux/Mac:
#   export BACKUP_YANDEX_TOKEN="your_oauth_token"
#   docker-compose -f docker-compose.personal.yml up -d backup_scheduler
#
#BACKUP_YANDEX_TOKEN=your_oauth_token  # <-- NOT RECOMMENDED
#BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups
#BACKUP_YANDEX_MAX_BACKUPS=10

# =============================================================================
# CI/CD Configuration
# =============================================================================
#CI_INTEGRATION_TEST_MIGRATE_ALL=true
#CI_INTEGRATION_TEST_TRUNCATE_LIST=notes,links,embeddings,recommendations


```

</details>

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

## Работа внутри Docker-сети

Docker Compose создаёт отдельную bridge-сеть. Сервисы внутри неё общаются по именам: `backend_personal`, `postgres_personal`, `redis_personal`, `mongo_personal`, `graph-service-personal`, `nlp-personal`.

### Узнать имя сети

```powershell
docker network ls
docker network inspect knowledge-graph_default
```

> Имя сети обычно `knowledge-graph_default` (для Personal — `knowledge-graph_personal_default`, если имя проекта задано через `docker compose --project-name`).

### Зайти внутрь контейнера

```powershell
# shell PostgreSQL
docker exec -it kg-postgres-personal bash

# shell Redis
docker exec -it kg-redis-personal sh

# shell бэкенда
docker exec -it kg-backend-personal sh

# shell фронтенда
docker exec -it kg-frontend-personal sh
```

### Выполнить SQL-запрос в PostgreSQL

```powershell
docker compose -f docker-compose.personal.yml exec -T postgres_personal psql -U personal -d knowledge_personal -c "SELECT id, email, role_id FROM users;"
```

Или через `docker exec`:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT * FROM notes LIMIT 5;"
```

### Redis

```powershell
docker compose -f docker-compose.personal.yml exec -T redis_personal redis-cli ping
docker compose -f docker-compose.personal.yml exec -T redis_personal redis-cli --scan --pattern "graph-service:*"
```

### MongoDB

```powershell
docker compose -f docker-compose.personal.yml exec -T mongo_personal mongosh knowledge_graph --eval "db.users.countDocuments()"
```

Если внутри контейнера нет `mongosh`, можно зайти в оболочку:

```powershell
docker exec -it kg-mongo-personal mongosh
use knowledge_graph
db.users.find().limit(5)
```

### Выполнить запрос с хоста, но через опубликованный порт

Если на хосте установлены клиенты, можно подключаться к проброшенным портам:

```powershell
psql -h 127.0.0.1 -p 5433 -U personal -d knowledge_personal -c "SELECT 1;"
redis-cli -h 127.0.0.1 -p 16380 ping
mongosh "mongodb://127.0.0.1:27018/knowledge_graph"
```

> Пароль PostgreSQL спросят. Для автоматизации: `set PGpassword=change_me_personal` (Windows) или `export PGPASSWORD=...`.

### Запустить временный контейнер внутри сети

```powershell
docker run --rm --network knowledge-graph_default -it nicolaka/netshoot
```

Внутри него:

```bash
ping backend_personal
dig postgres_personal
curl http://backend_personal:8080/health
curl http://graph-service-personal:9091/health
```

### Поднять отладочный контейнер с клиентами

```powershell
docker run --rm --network knowledge-graph_default -it alpine sh
# внутри:
apk add --no-cache postgresql-client redis mongodb-tools
psql -h postgres_personal -U personal -d knowledge_personal -c "SELECT 1;"
redis-cli -h redis_personal ping
mongosh mongodb://mongo_personal:27017/knowledge_graph --eval "db.users.countDocuments()"
```

### Скопировать файл в контейнер или из него

```powershell
# на хост → контейнер
docker cp ./my-script.sql kg-postgres-personal:/tmp/

# выполнить скрипт внутри контейнера
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -f /tmp/my-script.sql

# контейнер → хост
docker cp kg-postgres-personal:/tmp/result.csv .\result.csv
```

### Обратиться к бэкенду/граф-сервису изнутри сети

```bash
docker run --rm --network knowledge-graph_default curlimages/curl http://backend_personal:8080/health
docker run --rm --network knowledge-graph_default curlimages/curl http://graph-service-personal:9091/health
```

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

## Обновление между релизами

### Безопасный процесс обновления

1. Зафиксируй текущую версию:

```powershell
git log --oneline -1
```

2. Сделай бэкап Personal **перед** обновлением:

```powershell
.\scripts\devops\backup-personal.ps1 -Mode pre-upgrade
```

3. Обнови код:

```powershell
git pull origin ai-agents
```

4. Пересобери и перезапусти стек:

```powershell
docker compose -f docker-compose.personal.yml down
docker compose -f docker-compose.personal.yml up -d --build
```

5. Проверь, что миграции применились:

```powershell
docker logs -f kg-backend-personal
```

Ищи `Migrations applied successfully`. Если видишь `ERROR: Failed to run migrations`, действуй по разделу ниже.

6. Проверь health:

```powershell
curl http://127.0.0.1:18085/health
curl http://127.0.0.1:18082/health
```

7. Сделай smoke-тест: войди, открой заметку, открой граф.

### Миграции базы данных

Бэкенд при старте автоматически запускает миграции из `backend/migrations`. Если миграция зафейлилась, бэкенд падает с `log.Fatalf`, но при `MIGRATIONS_FAIL_ON_ERROR=false` может продолжить. В `.env` по умолчанию `MIGRATIONS_FAIL_ON_ERROR=false` — **не полагайся на это в продакшене**.

Проверить статус миграций вручную:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT id, version, applied_at FROM schema_migrations ORDER BY version;"
```

> Если таблица называется иначе, смотри `backend/migrations/*.sql`. Обычно `schema_migrations` создаётся `golang-migrate`.

### Если миграция не применилась

1. Читай лог бэкенда:

```powershell
docker logs kg-backend-personal | Select-String -Pattern "migration|migrations|ERROR"
```

2. Для **dev**-стека можно сбросить базу и пересоздать:

```powershell
docker compose down -v
docker compose up -d --build
```

> **Никогда** так не делай с Personal — там твои данные.

3. Для **Personal** восстановись из бэкапа:

```powershell
docker cp backups\backup-personal-pre-upgrade-....sql.gz kg-postgres-personal:/tmp/
docker exec -i kg-postgres-personal gunzip -c /tmp/backup-personal-pre-upgrade-....sql.gz | psql -U personal -d knowledge_personal
```

4. Если нужно откатить одну миграцию, используй `golang-migrate` (образ не в стеке, но есть на Docker Hub):

```powershell
docker run --rm --network knowledge-graph_default -v ${PWD}/backend/migrations:/migrations migrate/migrate -path /migrations -database "postgresql://personal:пароль@postgres_personal:5432/knowledge_personal?sslmode=disable" down 1
```

### Обновление с релиза на релиз (main)

Если работаешь с `main` вместо `ai-agents`, обычно достаточно:

```powershell
git fetch origin
git checkout main
git pull origin main
docker compose -f docker-compose.personal.yml down
docker compose -f docker-compose.personal.yml up -d --build
```

Но **перед этим** всегда — бэкап и проверка `CHANGELOG.md` / `docs/AI_HANDOFF.md` на ломающие изменения.

---

## Конфликты портов

Если при `docker compose up` появляется `Bind for 0.0.0.0:18082 failed: port is already allocated`, порт занят. Сначала выясни, кто.

### Windows

```powershell
# Найди PID
netstat -ano | findstr 18082
# или
Get-Process -Id (Get-NetTCPConnection -LocalPort 18082).OwningProcess

# Убедись, что это не твой Personal
Get-Process -Id <PID>
```

### Linux / macOS

```bash
lsof -i :18082
ss -ltnp | grep 18082
```

### Что может занимать порт

- Ранее запущенный Dev/Personal/Test-стек.
- Другой Docker-контейнер.
- Локальный сервис (IIS, SQL Server, PostgreSQL, Redis, Mongo).
- Antivirus / VPN / corporate proxy.

### Быстрое решение: сменить порты в `.env`

Найди переменные портов в `.env` и `docker-compose.personal.yml`:

```env
PERSONAL_API_PORT=18085
PERSONAL_VITE_API_URL=http://localhost:18085
```

и в `docker-compose.personal.yml`:

```yaml
nginx:
  ports:
    - "18082:80"       # API
    - "18084:8080"     # frontend
backend:
  ports:
    - "18085:8080"     # backend direct
postgres:
  ports:
    - "5433:5432"
redis:
  ports:
    - "16380:6379"
mongo:
  ports:
    - "27018:27017"
graph-service:
  ports:
    - "9092:9091"
nlp:
  ports:
    - "5001:5000"
```

Замени, например, `18082` → `28082`, `18084` → `28084`, `18085` → `28085`, `5433` → `15433`, `16380` → `26380`, `27018` → `37018`, `9092` → `19092`, `5001` → `15001.

> Важно: смени порты в обоих файлах (`.env` и `docker-compose.personal.yml`), иначе `VITE_API_URL` / `PERSONAL_VITE_API_URL` уедет.

После смены:

```powershell
docker compose -f docker-compose.personal.yml down
docker compose -f docker-compose.personal.yml up -d --build
```

### Если запущены сразу несколько стеков

Dev, Personal и Test могут работать одновременно, но каждому нужен свой набор портов. Сейчас в проекте они уже разнесены. Если ты менял порты вручную, проверь таблицу в начале документа и убедись, что нет дублей.

### Проверка

```powershell
docker ps --format "table {{.Names}}\t{{.Ports}}"
```

---

## Бэкап, восстановление и проверка

### Рекомендуемая стратегия

- **Раз в день** — автоматический SQL-бэкап PostgreSQL (`backup-scheduler`).
- **Перед обновлением** — ручной `pre-upgrade`.
- **Раз в неделю** — полный дамп Docker-томов на внешний диск.
- **После восстановления** — smoke-тест.

### SQL-бэкап

```powershell
.\scripts\devops\backup-personal.ps1 -Mode daily
```

Файл появится в `./backups/backup-personal-daily-<timestamp>.sql.gz`.

Внутри бэкапа:

```sql
-- public schema + data
-- pg_dump -F p --no-owner
```

Проверь размер:

```powershell
Get-ChildItem .\backups | Sort-Object Length -Descending | Select-Object -First 5
```

### Восстановление PostgreSQL из SQL-бэкапа

1. Останови стек:

```powershell
docker compose -f docker-compose.personal.yml down
```

2. (Опционально) пересоздай том, если он повреждён:

```powershell
docker volume rm pgdata_personal
```

> Делай это только если уверен, что бэкап рабочий.

3. Запусти только PostgreSQL:

```powershell
docker compose -f docker-compose.personal.yml up -d postgres_personal
```

4. Скопируй и распакуй бэкап:

```powershell
docker cp .\backups\backup-personal-daily-....sql.gz kg-postgres-personal:/tmp/backup.sql.gz
docker exec -i kg-postgres-personal gunzip -c /tmp/backup.sql.gz | psql -U personal -d knowledge_personal
```

5. Проверь:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT COUNT(*) FROM notes;"
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT COUNT(*) FROM users;"
```

6. Запусти остальной стек:

```powershell
docker compose -f docker-compose.personal.yml up -d
```

### Дамп Docker-томов

Полный образ томов Personal:

```powershell
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine tar czf /backup/pgdata_personal_$timestamp.tar.gz -C / data
docker run --rm -v redisdata_personal:/data -v "${PWD}:/backup" alpine tar czf /backup/redisdata_personal_$timestamp.tar.gz -C / data
docker run --rm -v mongodbdata_personal:/data -v "${PWD}:/backup" alpine tar czf /backup/mongodbdata_personal_$timestamp.tar.gz -C / data
```

### Восстановление из дампа томов

```powershell
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine sh -c "cd / && tar xzf /backup/pgdata_personal_....tar.gz"
docker run --rm -v redisdata_personal:/data -v "${PWD}:/backup" alpine sh -c "cd / && tar xzf /backup/redisdata_personal_....tar.gz"
docker run --rm -v mongodbdata_personal:/data -v "${PWD}:/backup" alpine sh -c "cd / && tar xzf /backup/mongodbdata_personal_....tar.gz"
```

### Smoke-тест после восстановления

```powershell
# 1. Все контейнеры Up (healthy)
docker compose -f docker-compose.personal.yml ps

# 2. Health endpoints
foreach ($p in 18082,18084,18085,9092,5001) { curl "http://127.0.0.1:$p/health" }

# 3. Login + note count
curl -X POST http://127.0.0.1:18082/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"you@example.com","password":"..."}'
curl http://127.0.0.1:18082/api/v1/notes
```

### Хранение бэкапов

- Держи как минимум 3 последних SQL-бэкапа.
- Копируй `backups/` на внешний диск / Yandex Disk.
- Для Yandex:

```env
BACKUP_CLOUD_ENABLED=true
BACKUP_CLOUD_PROVIDER=yandex
BACKUP_YANDEX_TOKEN=   # через env, не через .env
```

---

## WSL2 и Docker Desktop: ловушки Windows

### Почему `localhost` не работает, а `127.0.0.1` — да

Windows prefers IPv6 (`::1`) for `localhost`. Node, Playwright, Go и некоторые Docker-прокси могут либо не слушать `::1`, либо разрешать `localhost` в `::1`, хотя сервер на `127.0.0.1`. Всегда используй `127.0.0.1`.

- Playwright: `$env:FRONTEND_URL = "http://127.0.0.1:3002"`
- Frontend Vite: `VITE_API_URL` и `PERSONAL_VITE_API_URL` должны указывать на `127.0.0.1`.
- Backend health: `curl http://127.0.0.1:18085/health`.

### Диск WSL2 растёт

Docker Desktop с WSL2 хранит данные в `ext4.vhdx`. Со временем он разрастается.

Узнать размер:

```powershell
Get-ChildItem "$env:LOCALAPPDATA\Docker\wsl" -Recurse | Select-Object Name, @{N="SizeGB";E={[math]::Round($_.Length/1GB,2)}}
```

Очистить:

```powershell
wsl --shutdown
docker system prune -a
# Внутри WSL:
wsl -d docker-desktop-data -e sh -c "fstrim -av"
```

> `docker system prune -a` удаляет неиспользуемые образы и тома. Проверь, что Personal **остановлен**, иначе его тома могут пострадать.

### Docker Desktop не перезапускает WSL

После `wsl --shutdown` или обновления Docker Desktop:

```powershell
wsl --list --verbose
wsl --terminate docker-desktop
wsl --terminate docker-desktop-data
```

### Docker Desktop использует Hyper-V или WSL2

В настройках Docker Desktop:

- Settings → General → Use the WSL2 based engine (рекомендуется).
- Settings → Resources → WSL integration → включи для своего дистрибутива.

### Антивирус и файервол

- Windows Defender Controlled Folder Access может блокировать Docker bind-mount (`./backups`, `./huggingface_cache`).
- McAfee / Symantec / Kaspersky могут сканировать VHD и падать.
- Добавь папку `D:\knowledge-graph` в исключения.

### Пути Windows ↔ WSL

Bind-mount в `docker-compose.personal.yml` указывает на `./huggingface_cache` и `./backups`. Docker Desktop на WSL2 нормализует пути, но на Hyper-V это может сломаться.

### `wsl --shutdown` безопасный

```powershell
wsl --shutdown
```

Это безопасная команда — она останавливает WSL, но не удаляет данные. После `docker compose up` всё поднимется.

---

## Журнал логов: типичные сообщения и что они значат

### Как смотреть логи

```powershell
# посмотреть последние 100 строк
docker logs --tail 100 kg-backend-personal

# следить в реальном времени
docker logs -f kg-backend-personal

# логи всех сервисов одновременно
docker compose -f docker-compose.personal.yml logs -f

# логи с временем
docker logs -t --tail 50 kg-graph-service-personal
```

### Бэкенд

| Сообщение | Значение | Что делать |
|---|---|---|
| `Migrations applied successfully` | Всё норм. | — |
| `ERROR: Failed to run migrations` | Миграция упала. | Смотри `version`, `error`, правь или откатывай. |
| `server error` / `500` | Ошибка в handler. | Смотри stack trace. |
| `JWT token is missing` | Запрос без токена. | Проверь `Authorization` в `.env`/curl. |
| `connect: connection refused` | Бэкенд не может подключиться к Postgres/Redis/Mongo. | Проверь `DATABASE_URL`, `REDIS_URL`, `MONGO_URL`. |
| `failed to connect to `host=postgres_personal`:` | Postgres недоступен. | Проверь `docker ps`. |

### Graph service

| Сообщение | Значение | Что делать |
|---|---|---|
| `Graph service started` | Всё норм. | — |
| `JWT verification failed` | `JWT_SECRET` не совпадает с backend. | Синхронизируй `JWT_SECRET`. |
| `connection refused redis` | Не видит Redis. | Проверь `REDIS_URL`. |

### Worker

| Сообщение | Значение | Что делать |
|---|---|---|
| `worker started` | Всё норм. | — |
| `asynq: ready` | Подключился к очереди. | — |
| `error processing task` | Задача упала. | Проверь payload, NLP, DB. |

### NLP

| Сообщение | Значение | Что делать |
|---|---|---|
| `Model loaded` | Всё норм. | — |
| `Model not found` | Нет кэша. | Скачай модель. |
| `CUDA out of memory` | Не хватает VRAM. | Уменьши batch, не используй GPU. |

### PostgreSQL

| Сообщение | Значение | Что делать |
|---|---|---|
| `database system is ready to accept connections` | Всё норм. | — |
| `password authentication failed` | Неверный пароль. | Проверь `.env` и `docker-compose.personal.yml`. |
| `could not create lock file` | Нет прав на диск. | Проверь volume permissions. |

### Redis

| Сообщение | Значение | Что делать |
|---|---|---|
| `Ready to accept connections` | Всё норм. | — |
| `MISCONF Redis is configured to save RDB snapshots` | Redis не может сохранить dump. | Проверь права на `/data`. |

### MongoDB

| Сообщение | Значение | Что делать |
|---|---|---|
| `Waiting for connections` | Всё норм. | — |
| `Unrecognized option: --auth` | Возможна опечатка в compose. | Проверь `docker-compose.personal.yml`. |

---

## Харднинг и безопасность

### Секреты

- `JWT_SECRET` — минимум 32 случайных символа. Не используй `change_me_*`.
- `POSTGRES_PASSWORD` / `PERSONAL_POSTGRES_PASSWORD` — сложные, не в коммитах.
- `GRAPH_SERVICE_INTERNAL_TOKEN` — включи, если backend и graph-service в одной сети.
- `BACKUP_YANDEX_TOKEN` — передавай через env, не в `.env`.
- `SMTP_PASSWORD` — используй App Password, не основной пароль.

```powershell
# сгенерировать JWT_SECRET
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 } | ForEach-Object { [byte]$_ }))
```

### SKIP_AUTH

`SKIP_AUTH=true` **только** для локального теста. В Personal и production:

```env
SKIP_AUTH=false
```

### CORS

В продакшене указывай точные origin:

```env
CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
```

Не оставляй `http://localhost:*` в продакшене.

### HTTPS

- Локально можно HTTP.
- В продакшене всегда HTTPS + HSTS + CSP.
- См. раздел `Production: HTTPS и SSL`.

### Сетевой доступ

- Не выставяй порты Postgres/Redis/Mongo наружу.
- В `docker-compose.personal.yml` `127.0.0.1:5433:5432`, а не `5433:5432`.
- nginx — единственный публичный вход.

### Docker

- Обновляй базовые образы (`FROM ...`) регулярно.
- Запускай контейнеры не от root, где возможно:

```yaml
services:
  nlp:
    user: "1000:1000"
```

- Ограничивай ресурсы:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### Бэкапы и целостность

- Бэкапы на другой диск.
- Периодически проверяй восстановление на тестовом стеке.
- Шифруй чувствительные backup-токены.

---

## Перформанс-тюнинг

### Docker Desktop

Открой Settings → Resources и задай лимиты:

- **CPU**: 4+ cores.
- **Memory**: 8 GB minimum, 16 GB рекомендуется.
- **Swap**: 2 GB.
- **Disk image location**: SSD.

### Лимиты в compose

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
  nlp:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

> NLP модель съедает 1–2 GB RAM. Оставь запас для Postgres и Redis.

### PostgreSQL

Добавь в `docker-compose.personal.yml` или внешний `postgresql.conf`:

```env
POSTGRES_INITDB_ARGS="--encoding=UTF-8"
POSTGRES_EXTRA_ARGS="-c shared_buffers=512MB -c work_mem=64MB -c maintenance_work_mem=256MB -c max_connections=50"
```

или монтируй `postgresql.conf`:

```yaml
services:
  postgres_personal:
    volumes:
      - ./postgres/postgresql.conf:/etc/postgresql/postgresql.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
```

Проверь использование:

```sql
SELECT * FROM pg_stat_activity;
SELECT pg_size_pretty(pg_database_size('knowledge_personal'));
```

### Redis

Добавь persistence:

```yaml
services:
  redis_personal:
    command: redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
```

### NLP

- Первый запуск медленный — кэшируй `huggingface_cache`.
- Для CPU-inference отключи GPU:

```env
NLP_USE_GPU=false
```

- Если embedding занимает слишком много времени, уменьши `RECOMMENDATION_TOP_N` в `knowledge-graph.config.json`.

### Frontend

- В production используй `npm run build` (уже внутри Docker).
- `knowledge-graph.config.json` влияет на лимиты графа:
  - `frontend.graph.2d.max_nodes` — уменьши, если FPS падает.
  - `frontend.graph.2d.fog` — адаптивный туман помогает на слабом железе.

### Мониторинг нагрузки

```powershell
docker stats
# или
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT COUNT(*) FROM notes;"
```

### Признаки, что не хватает ресурсов

- `OOMKilled` в `docker ps` — увеличь лимит RAM.
- `context deadline exceeded` — backend не успевает, проверь CPU/DB.
- `Graph not responding` — graph-service не хватает RAM или виснет Redis.

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
