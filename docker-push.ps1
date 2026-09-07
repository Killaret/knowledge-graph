# docker-push.ps1
# Скрипт для публикации образов Knowledge Graph в Docker Hub

$DOCKER_USER = "killaret"
$REPO_PREFIX = "$DOCKER_USER/knowledge-graph"

Write-Host "=== Docker Push Script ===" -ForegroundColor Cyan
Write-Host "Repository: $REPO_PREFIX" -ForegroundColor Yellow
Write-Host ""

# Проверка логина
Write-Host "Проверка логина в Docker Hub..." -ForegroundColor Gray
try {
    $test = docker info | Select-String "Username"
    if (-not $test) {
        Write-Host "Вы не авторизованы. Выполните: docker login" -ForegroundColor Red
        exit 1
    }
    Write-Host "Авторизация: $test" -ForegroundColor Green
} catch {
    Write-Host "Ошибка проверки авторизации" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Тегирование и пуш образов ===" -ForegroundColor Cyan

# Backend
Write-Host "`n[1/5] Backend..." -ForegroundColor Yellow
docker tag knowledge-graph-backend:latest "${REPO_PREFIX}-backend:latest" -ErrorAction Stop
docker push "${REPO_PREFIX}-backend:latest" -ErrorAction Stop
Write-Host "✓ Backend запушен" -ForegroundColor Green

# Frontend
Write-Host "`n[2/5] Frontend..." -ForegroundColor Yellow
docker tag knowledge-graph-frontend:latest "${REPO_PREFIX}-frontend:latest" -ErrorAction Stop
docker push "${REPO_PREFIX}-frontend:latest" -ErrorAction Stop
Write-Host "✓ Frontend запушен" -ForegroundColor Green

# NLP
Write-Host "`n[3/5] NLP..." -ForegroundColor Yellow
docker tag knowledge-graph-nlp:latest "${REPO_PREFIX}-nlp:latest" -ErrorAction Stop
docker push "${REPO_PREFIX}-nlp:latest" -ErrorAction Stop
Write-Host "✓ NLP запушен" -ForegroundColor Green

# Graph-service
Write-Host "`n[4/5] Graph-service..." -ForegroundColor Yellow
docker tag knowledge-graph-graph-service:latest "${REPO_PREFIX}-graph-service:latest" -ErrorAction Stop
docker push "${REPO_PREFIX}-graph-service:latest" -ErrorAction Stop
Write-Host "✓ Graph-service запушен" -ForegroundColor Green

# Worker
Write-Host "`n[5/5] Worker..." -ForegroundColor Yellow
docker tag knowledge-graph-worker:latest "${REPO_PREFIX}-worker:latest" -ErrorAction Stop
docker push "${REPO_PREFIX}-worker:latest" -ErrorAction Stop
Write-Host "✓ Worker запушен" -ForegroundColor Green

Write-Host "`n=== Готово! ===" -ForegroundColor Green
Write-Host "`nОбразы доступны по адресам:" -ForegroundColor Cyan
Write-Host "  https://hub.docker.com/r/$DOCKER_USER/knowledge-graph-backend"
Write-Host "  https://hub.docker.com/r/$DOCKER_USER/knowledge-graph-frontend"
Write-Host "  https://hub.docker.com/r/$DOCKER_USER/knowledge-graph-nlp"
Write-Host "  https://hub.docker.com/r/$DOCKER_USER/knowledge-graph-graph-service"
Write-Host "  https://hub.docker.com/r/$DOCKER_USER/knowledge-graph-worker"
Write-Host ""
