@echo off
setlocal enabledelayedexpansion

title Knowledge Graph - Async Testing

echo.
echo ========================================
echo Knowledge Graph - Full Async Test
echo ========================================
echo.

cd /d D:\knowledge-graph

echo [1] Stopping old containers...
docker-compose down
echo.

echo [2] Building and starting services (this may take 3-5 minutes)...
docker-compose up --build -d
echo.

echo [3] Waiting 45 seconds for services to initialize...
echo Please wait...
timeout /t 45 /nobreak

echo.
echo [4] Service Status:
echo ================
docker-compose ps
echo.

echo [5] Backend Logs (checking Asynq initialization):
echo ==================================================
docker-compose logs --tail=25 kg-backend
echo.

echo [6] Worker Logs (checking if listening for tasks):
echo ===================================================
docker-compose logs --tail=25 kg-worker
echo.

echo [7] Redis Status:
echo =================
docker-compose logs --tail=10 kg-redis
echo.

echo [8] Checking Redis queue BEFORE test:
echo ======================================
docker exec kg-redis redis-cli KEYS "*asynq*"
echo.

echo [9] Creating test note (triggering async tasks)...
echo ===================================================
curl -X POST http://localhost:8080/notes ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Async Test Note\",\"content\":\"This note triggers keyword extraction and embedding computation async tasks\"}"
echo.
echo.

echo [10] Waiting 15 seconds for worker to process tasks...
timeout /t 15 /nobreak
echo.

echo [11] Checking Redis queue AFTER processing:
echo =============================================
docker exec kg-redis redis-cli KEYS "*asynq*"
echo.
docker exec kg-redis redis-cli LLEN "asynq:queues:default"
echo.

echo [12] Worker Logs - Looking for successful task processing:
echo ============================================================
docker-compose logs kg-worker | findstr "HandleExtractKeywords HandleComputeEmbedding successfully"
echo.

echo [13] FULL Worker Logs (last 100 lines):
echo ======================================
docker-compose logs --tail=100 kg-worker
echo.

echo ========================================
echo TEST COMPLETE
echo ========================================
echo.
echo Summary:
echo - Check above for "HandleExtractKeywords: successfully processed"
echo - Check above for "HandleComputeEmbedding: successfully processed"
echo - These indicate successful async task processing
echo.
echo To keep watching worker logs, run:
echo   docker-compose logs -f kg-worker
echo.
pause
