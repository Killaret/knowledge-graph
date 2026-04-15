# Руководство по развёртыванию Knowledge Graph

> **Версия:** 1.0  
> **Дата:** Апрель 2026  
> **Статус:** Production Ready

---

## 📋 Содержание

1. [Требования](#требования)
2. [Локальное развёртывание](#локальное-развёртывание)
3. [Production развёртывание](#production-развёртывание)
4. [Kubernetes (K8s)](#kubernetes-k8s)
5. [Обновление](#обновление)
6. [Откат](#откат)
7. [Мониторинг](#мониторинг)

---

## Требования

### Минимальные (MVP)

| Компонент | Версия | CPU | RAM | Диск |
|-----------|--------|-----|-----|------|
| **Backend** | Go 1.21+ | 0.5 core | 512 MB | 100 MB |
| **Frontend** | Node 20+ | 0.5 core | 256 MB | 50 MB |
| **PostgreSQL** | 16+ | 1 core | 1 GB | 10 GB |
| **Redis** | 7+ | 0.25 core | 256 MB | 1 GB |
| **NLP** | Python 3.11+ | 1 core | 2 GB | 2 GB |

### Рекомендуемые (Production)

| Компонент | CPU | RAM | Диск | Примечания |
|-----------|-----|-----|------|------------|
| **Backend** | 2 cores | 2 GB | 1 GB | Можно масштабировать |
| **Frontend** | 1 core | 512 MB | 100 MB | Static files |
| **PostgreSQL** | 4 cores | 4 GB | 100 GB | SSD обязательно |
| **Redis** | 1 core | 1 GB | 5 GB | Persistence включена |
| **NLP** | 4 cores | 8 GB | 5 GB | GPU опционально |

### GPU требования (для NLP)

NLP сервис работает на CPU, но GPU значительно ускоряет генерацию эмбеддингов:

- **Минимум**: CPU-only (медленно для больших текстов)
- **Рекомендуется**: NVIDIA GPU с CUDA 11.8+
- **Модели**: all-MiniLM-L6-v2 (384 dim), использует ~500MB VRAM

---

## Локальное развёртывание

### 1. Подготовка окружения

```bash
# Клонируйте репозиторий
git clone https://github.com/your-org/knowledge-graph.git
cd knowledge-graph

# Проверьте Docker и Docker Compose
docker --version  # 24.0+
docker-compose --version  # 2.20+

# Создайте .env файл
cp .env.example .env
```

### 2. Конфигурация (.env)

```env
# === Обязательные ===
DATABASE_URL=postgresql://kg_user:kg_password@postgres:5432/knowledge_graph?sslmode=disable

# === Опциональные ===
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000

# === Рекомендации (оставьте дефолтные для начала) ===
RECOMMENDATION_ALPHA=0.7
RECOMMENDATION_BETA=0.3
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_CACHE_TTL_SECONDS=300
EMBEDDING_SIMILARITY_LIMIT=30

# === Визуализация ===
GRAPH_LOAD_DEPTH=2
```

### 3. Запуск

```bash
# Сборка и запуск всех сервисов
docker-compose up --build -d

# Проверка статуса
docker-compose ps

# Логи
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f worker
docker-compose logs -f nlp
```

### 4. Инициализация базы данных

```bash
# Применение миграций
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up

# Сидинг тестовых данных (опционально)
docker-compose exec backend go run /app/scripts/seed.go
```

### 5. Проверка работоспособности

```bash
# Health check backend
curl http://localhost:8080/health
# → {"status":"ok"}

# Health check DB
curl http://localhost:8080/db-check
# → {"status":"db ok"}

# NLP сервис
curl http://localhost:5000/health
# → {"status":"ok"}

# Frontend
curl http://localhost:5173
# → HTML страница
```

### 6. Доступ к приложению

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Docs** (если настроено): http://localhost:8080/swagger

---

## Production развёртывание

### Вариант A: Docker Compose на сервере

#### 1. Подготовка сервера

```bash
# Ubuntu 22.04 LTS
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo apt install docker-compose-plugin -y
```

#### 2. Production конфигурация

Создайте `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - NLP_SERVICE_URL=${NLP_SERVICE_URL}
      - RECOMMENDATION_ALPHA=${RECOMMENDATION_ALPHA:-0.7}
      - RECOMMENDATION_BETA=${RECOMMENDATION_BETA:-0.3}
      - SERVER_PORT=8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  worker:
    build:
      context: ./backend
      dockerfile: Dockerfile.worker
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - NLP_SERVICE_URL=${NLP_SERVICE_URL}
    depends_on:
      - postgres
      - redis
      - nlp
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        - PUBLIC_API_URL=/api
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
    restart: unless-stopped

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      - POSTGRES_USER=${DB_USER:-kg_user}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME:-knowledge_graph}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nlp:
    build:
      context: ./nlp-service
      dockerfile: Dockerfile
    environment:
      - MODEL_NAME=all-MiniLM-L6-v2
      - CACHE_DIR=/app/cache
    volumes:
      - nlp_cache:/app/cache
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  nlp_cache:
```

#### 3. SSL/TLS с Let's Encrypt

```bash
# Установка certbot
sudo apt install certbot -y

# Получение сертификата
sudo certbot certonly --standalone -d your-domain.com

# Настройка nginx (в Docker или на хосте)
```

#### 4. Запуск

```bash
# Production запуск
docker-compose -f docker-compose.prod.yml up -d

# Масштабирование worker'ов
docker-compose -f docker-compose.prod.yml up -d --scale worker=3
```

---

## Kubernetes (K8s)

### Структура манифестов

```
k8s/
├── namespace.yaml
├── configmap.yaml          # Не-чувствительные настройки
├── secret.yaml             # Чувствительные данные (base64)
├── postgres/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── pvc.yaml
├── redis/
│   ├── deployment.yaml
│   └── service.yaml
├── backend/
│   ├── deployment.yaml
│   └── service.yaml
├── worker/
│   └── deployment.yaml
├── frontend/
│   ├── deployment.yaml
│   └── service.yaml
└── ingress.yaml            # Вход с SSL
```

### Быстрый старт с kubectl

```bash
# Создание namespace
kubectl apply -f k8s/namespace.yaml

# Config и Secrets
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# База данных
kubectl apply -f k8s/postgres/

# Кэш
kubectl apply -f k8s/redis/

# Приложение
kubectl apply -f k8s/backend/
kubectl apply -f k8s/worker/
kubectl apply -f k8s/frontend/

# Ingress (требуется ingress-nginx)
kubectl apply -f k8s/ingress.yaml
```

### Масштабирование

```bash
# Масштабирование backend
kubectl scale deployment backend --replicas=3

# Масштабирование workers
kubectl scale deployment worker --replicas=5

# HPA (Horizontal Pod Autoscaler)
kubectl autoscale deployment backend --min=2 --max=10 --cpu-percent=70
```

---

## Обновление

### Docker Compose обновление

```bash
# 1. Бэкап базы данных
docker-compose exec postgres pg_dump -U kg_user knowledge_graph > backup_$(date +%Y%m%d).sql

# 2. Pull новых образов
docker-compose pull

# 3. Применение миграций
docker-compose run --rm backend migrate -path /app/migrations -database "$DATABASE_URL" up

# 4. Перезапуск с zero-downtime (если настроено)
docker-compose up -d

# 5. Проверка
./scripts/health-check.sh
```

### Rolling Update в Kubernetes

```bash
# Обновление образа
kubectl set image deployment/backend backend=kg-backend:v1.1.0

# Отслеживание статуса
kubectl rollout status deployment/backend

# Откат при проблемах
kubectl rollout undo deployment/backend
```

---

## Откат

### Быстрый откат (Docker Compose)

```bash
# Восстановление из бэкапа
docker-compose exec postgres psql -U kg_user -d knowledge_graph < backup_20250415.sql

# Откат к предыдущей версии
git checkout v1.0.0
docker-compose up --build -d
```

### Откат миграций

```bash
# Откат на N версий
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" down 3

# Откат всего
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" down
```

---

## Мониторинг

### Health Checks

```bash
# Автоматическая проверка
./scripts/health-check.sh

# Ручная проверка всех компонентов
curl http://localhost:8080/health
curl http://localhost:8080/db-check
curl http://localhost:5000/health
docker-compose exec redis redis-cli ping
```

### Логи

```bash
# Все логи
docker-compose logs --tail=100 -f

# Конкретный сервис
docker-compose logs -f backend

# JSON формат для анализа
docker-compose logs backend --format json
```

### Метрики (опционально)

```bash
# Prometheus + Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Доступ к Grafana
curl http://localhost:3000  # admin/admin
```

---

## Troubleshooting

### Проблема: Backend не стартует

```bash
# Проверка логов
docker-compose logs backend

# Частые причины:
# 1. Нет подключения к БД
docker-compose exec backend nc -zv postgres 5432

# 2. Неприменённые миграции
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up
```

### Проблема: NLP сервис медленный

```bash
# Проверка ресурсов
docker stats kg-nlp

# Загрузка модели занимает время при первом старте
# Проверка кэша модели
docker-compose exec nlp ls -la /app/cache/
```

### Проблема: Worker не обрабатывает задачи

```bash
# Проверка очереди
docker-compose exec redis redis-cli LLEN asynq:{default}

# Перезапуск worker'а
docker-compose restart worker
```

---

## Чек-лист Production развёртывания

- [ ] Создан `.env` файл с production значениями
- [ ] Настроены сложные пароли (DB, Redis)
- [ ] Настроен SSL/TLS сертификат
- [ ] Бэкап БД настроен (cron + pg_dump)
- [ ] Мониторинг health checks настроен
- [ ] Логи отправляются в централизованное хранилище
- [ ] Масштабирование backend/worker настроено
- [ ] Размеры ресурсов (limits/requests) настроены
- [ ] Readiness/Liveness probes настроены (K8s)
- [ ] PDB (Pod Disruption Budget) настроен (K8s)

---

## Полезные ссылки

- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Kubernetes Docs](https://kubernetes.io/docs/)
- [pgvector](https://github.com/pgvector/pgvector)
- [Asynq](https://github.com/hibiken/asynq)
