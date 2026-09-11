# Развёртывание Knowledge Graph на новой машине

> Практическое руководство по запуску трёх Docker-стеков: **dev** (разработка), **personal** (личные данные) и **test** (E2E/BDD).  
> Для production, Kubernetes и CI/CD детали смотри [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md).  
> Тестирование — [`docs/TESTING.md`](docs/TESTING.md), бэкапы — [`docs/BACKUP.md`](docs/BACKUP.md), конфигурация — [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md).

---

## Содержание

1. [Что понадобится](#что-понадобится)
2. [Быстрый старт (TL;DR)](#быстрый-старт-tldr)
3. [Клонирование и ветка](#клонирование-и-ветка)
4. [Настройка `.env`](#настройка-env)
5. [NLP-модель и `huggingface_cache`](#nlp-модель-и-huggingface_cache)
6. [Три стека: как запустить](#три-стека-как-запустить)
   - [Personal](#personal)
   - [Dev](#dev)
   - [Test](#test)
7. [Порты и URL после старта](#порты-и-url-после-старта)
8. [Проверка, что всё поднялось](#проверка-что-всё-поднялось)
9. [Первая регистрация и вход](#первая-регистрация-и-вход)
10. [Частые проблемы](#частые-проблемы)
11. [Обновление кода](#обновление-кода)
12. [Перенос Personal-данных на другую машину](#перенос-personal-данных-на-другую-машину)
13. [Что не надо трогать](#что-не-надо-трогать)

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

> Сейчас в корне репозитория `D:\knowledge-graph`, текущая ветка `ai-agents`.

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

Остальные переменные можно оставить по умолчанию.  
Если планируешь облачный бэкап, добавь `BACKUP_YANDEX_TOKEN` **в окружение**, а не в `.env` — см. [`docs/BACKUP.md`](docs/BACKUP.md) и `SEC-2` в [`docs/AI_HANDOFF.md`](docs/AI_HANDOFF.md).

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

## Частые проблемы

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
