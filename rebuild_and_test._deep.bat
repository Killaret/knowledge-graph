@echo off
setlocal enabledelayedexpansion

title Knowledge Graph - Rebuild & Test

echo.
echo ========================================
echo Knowledge Graph - Rebuild and Async Test
echo ========================================
echo.

cd /d D:\knowledge-graph

echo [1] Stopping old containers...
docker-compose down
echo.

echo [2] Building and starting services (this may take 5-10 minutes)...
docker-compose up --build -d
echo.

echo [3] Waiting for services to initialize...
timeout /t 30 /nobreak >nul

:: Ожидание готовности NLP (проверяем через curl, но контейнер называется kg-nlp)
echo [3a] Checking NLP service status...
set NLP_READY=0
for /l %%i in (1,1,30) do (
    curl -s http://localhost:5000/health >nul 2>&1
    if !errorlevel! equ 0 (
        set NLP_READY=1
        echo NLP is ready.
        goto :nlp_ready
    )
    echo Waiting for NLP... %%i/30
    timeout /t 2 /nobreak >nul
)
:nlp_ready

if !NLP_READY! equ 0 (
    echo WARNING: NLP service not ready. Check logs.
    docker-compose logs --tail=50 nlp
)

echo.
echo [4] Service Status:
docker-compose ps

echo.
echo [5] Backend logs (Asynq init):
docker-compose logs --tail=20 backend

echo.
echo [6] Redis queue BEFORE test:
docker exec kg-redis redis-cli KEYS "*asynq*"

echo.
echo [7] Creating test note...
curl -X POST http://localhost:8080/notes ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Async Test Note\",\"content\":\"This note triggers keyword extraction and embedding computation async tasks\"}"
echo.

echo.
echo [8] Waiting 15 seconds for worker to process...
timeout /t 15 /nobreak

echo.
echo [9] Redis queue AFTER test:
docker exec kg-redis redis-cli KEYS "*asynq*"
docker exec kg-redis redis-cli LLEN "asynq:queues:default"

echo.
echo [10] Worker logs (last 30 lines):
docker-compose logs --tail=30 worker

echo.
echo [11] Checking for successful task processing:
docker-compose logs worker | findstr "successfully processed"

echo.
echo ========================================
echo Test complete.
echo ========================================
pause