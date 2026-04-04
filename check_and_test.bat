@echo off
setlocal enabledelayedexpansion

title Knowledge Graph - Check & Test

echo.
echo ========================================
echo Knowledge Graph - Service Check & Async Test
echo ========================================
echo.

cd /d D:\knowledge-graph

:: Проверка статуса контейнеров
echo [1] Checking service status...
docker-compose ps
echo.

:: Проверка, что все нужные сервисы запущены
set ALL_UP=1
docker-compose ps --quiet backend >nul 2>&1 || set ALL_UP=0
docker-compose ps --quiet worker >nul 2>&1 || set ALL_UP=0
docker-compose ps --quiet nlp >nul 2>&1 || set ALL_UP=0
docker-compose ps --quiet postgres >nul 2>&1 || set ALL_UP=0
docker-compose ps --quiet redis >nul 2>&1 || set ALL_UP=0

if %ALL_UP% equ 0 (
    echo ERROR: Not all services are running.
    echo Please run 'docker-compose up -d' first.
    goto :end
)

:: Проверка доступности бэкенда
echo [2] Checking backend health...
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo ERROR: Backend not responding.
    goto :end
)
echo Backend OK.

:: Проверка доступности NLP
echo [3] Checking NLP health...
set NLP_READY=0
for /l %%i in (1,1,15) do (
    curl -s http://localhost:5000/health >nul 2>&1
    if !errorlevel! equ 0 (
        set NLP_READY=1
        echo NLP is ready.
        goto :nlp_ready
    )
    echo Waiting for NLP... %%i/15
    timeout /t 2 /nobreak >nul
)
:nlp_ready

if !NLP_READY! equ 0 (
    echo ERROR: NLP service not ready.
    goto :end
)

:: Проверка Redis очереди ДО теста
echo.
echo [4] Redis queue BEFORE test:
docker exec kg-redis redis-cli KEYS "*asynq*" 2>nul
echo.

:: Создание тестовой заметки
echo [5] Creating test note...
curl -X POST http://localhost:8080/notes ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Async Test Note\",\"content\":\"This note triggers keyword extraction and embedding computation async tasks\"}"
echo.

:: Ожидание обработки воркером
echo.
echo [6] Waiting 15 seconds for worker to process...
timeout /t 15 /nobreak

:: Проверка Redis очереди ПОСЛЕ теста
echo.
echo [7] Redis queue AFTER test:
docker exec kg-redis redis-cli KEYS "*asynq*"
docker exec kg-redis redis-cli LLEN "asynq:queues:default"

:: Логи воркера (последние 30 строк)
echo.
echo [8] Worker logs (last 30 lines):
docker-compose logs --tail=30 worker

:: Проверка успешной обработки
echo.
echo [9] Checking for successful task processing:
docker-compose logs worker | findstr "successfully processed"
if errorlevel 1 (
    echo WARNING: No successful task processing found.
) else (
    echo SUCCESS: Tasks were processed.
)

:end
echo.
echo ========================================
echo Test complete.
echo ========================================
pause