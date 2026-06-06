# Инструменты DevOps Агента

## 🛠️ Набор инструментов

### 1. Docker Management

#### Проверка здоровья сервисов
```powershell
docker compose ps
```

#### Перезапуск сервиса
```powershell
docker compose restart <service-name>
```

#### Просмотр логов
```powershell
docker compose logs -f <service-name>
docker compose logs --tail=100 <service-name>
```

#### Сброс кэша и очистка
```powershell
docker system prune -f
docker image prune -a -f
docker volume prune -f
```

### 2. Database Operations

#### Выполнение миграций
```powershell
docker compose exec backend go run ./cmd/server migrate
```

#### Бэкап БД
```powershell
docker compose exec postgres pg_dump -U postgres knowledge_base > backup.sql
```

#### Восстановление БД
```powershell
docker compose exec -T postgres psql -U postgres knowledge_base < backup.sql
```

#### Проверка подключения
```powershell
docker compose exec postgres pg_isready -U postgres
```

### 3. Redis Operations

#### Очистка кэша
```powershell
docker compose exec redis redis-cli FLUSHALL
docker compose exec redis redis-cli FLUSHDB
```

#### Проверка статуса
```powershell
docker compose exec redis redis-cli INFO
docker compose exec redis redis-cli PING
```

### 4. Testing Tools

#### Backend tests
```powershell
cd backend
go test -race -cover ./...
go test -v ./internal/application/graph
```

#### Frontend tests
```powershell
cd frontend
npm run test:unit
npm run test:coverage
npm run test:e2e
```

#### Linting
```powershell
# Backend
golangci-lint run ./...

# Frontend
npm run lint
npm run format:check
```

### 5. Monitoring & Metrics

#### Проверка API endpoints
```powershell
# Health check
curl http://localhost:8080/health

# Graph API
docker exec kg-graph-service curl -s "http://localhost:9091/api/v1/graph/full" | jq '.meta'

# Auth API
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password"}'
```

#### Проверка производительности
```powershell
# Load test
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/graph/full

# Simple benchmark
for i in {1..100}; do
  curl -s http://localhost:8080/api/v1/graph/full > /dev/null
done
```

### 6. Deployment Tools

#### Build Docker images
```powershell
# Backend
docker build -t kg-backend:latest ./backend

# Frontend
docker build -t kg-frontend:latest ./frontend

# Multi-stage
docker buildx build --platform linux/amd64,linux/arm64 \
  -t kg-backend:latest --push ./backend
```

#### Deploy to Kubernetes
```powershell
kubectl apply -f k8s/deployment.yaml
kubectl rollout status deployment/backend
kubectl get pods -l app=backend
```

#### Rollback
```powershell
kubectl rollout undo deployment/backend
kubectl rollout status deployment/backend
```

### 7. Backup & Recovery

#### Полный бэкап
```powershell
./scripts/backup-database.sh
./scripts/backup-redis.sh
./scripts/backup-volumes.sh
```

#### Проверка бэкапов
```powershell
ls -lh /backups/postgres/
find /backups -name "*.dump.gz" -mtime -7
```

### 8. Security Scanning

#### Trivy сканирование
```powershell
trivy fs .
trivy image kg-backend:latest
trivy k8s --namespace default deployment/backend
```

#### Checkov для IaC
```powershell
checkov -d .k8s/
checkov -d docker-compose.yml
```

---

## 📋 Чеклисты

### Pre-deployment Checklist
- [ ] Все тесты passed
- [ ] Coverage > 60%
- [ ] Security scan clean
- [ ] Миграции протестированы
- [ ] Бэкап выполнен
- [ ] Health checks green

### Post-deployment Checklist
- [ ] Все сервисы healthy
- [ ] Smoke tests passed
- [ ] Error rate < 1%
- [ ] Latency p95 < 500ms
- [ ] Logs clean (no errors)
- [ ] Metrics collecting

### Rollback Checklist
- [ ] Откат выполнен
- [ ] Сервисы восстановлены
- [ ] Данные восстановлены из бэкапа
- [ ] Уведомлены стейкхолдеры
- [ ] Post-mortem создан

---

## 🔧 Скрипты автоматизации

### Health Check Script
```powershell
# health-check.ps1
$services = @("backend", "frontend", "graph-service", "postgres", "redis")

foreach ($service in $services) {
    $status = docker compose ps $service
    if ($status -notlike "*Up*") {
        Write-Host "❌ $service is down!" -ForegroundColor Red
        exit 1
    }
}

Write-Host "✅ All services healthy" -ForegroundColor Green
```

### Full Test Suite
```powershell
# run-all-tests.ps1
Write-Host "Running backend tests..." -ForegroundColor Cyan
cd backend
go test -race -cover ./...
cd ..

Write-Host "Running frontend tests..." -ForegroundColor Cyan
cd frontend
npm run test:unit
cd ..

Write-Host "All tests completed!" -ForegroundColor Green
```

### Performance Benchmark
```powershell
# benchmark.ps1
Write-Host "Starting performance benchmark..." -ForegroundColor Cyan

$endpoints = @(
    "/api/v1/graph/full",
    "/api/v1/graph/note/1?depth=2",
    "/api/v1/auth/login"
)

foreach ($endpoint in $endpoints) {
    Write-Host "Testing $endpoint" -ForegroundColor Gray
    $result = wrk -t4 -c100 -d10s "http://localhost:8080$endpoint"
    Write-Host $result
}
```
