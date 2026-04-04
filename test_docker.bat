@echo off
cd D:\knowledge-graph

REM Step 1: Stop old containers
echo.
echo === Stopping old containers ===
docker-compose down

REM Step 2: Clean rebuild and start
echo.
echo === Building and starting services ===
docker-compose up --build -d

REM Step 3: Wait for services
echo.
echo === Waiting 45 seconds for services to be ready ===
timeout /t 45 /nobreak

REM Step 4: Check status
echo.
echo === Checking service status ===
docker-compose ps

REM Step 5: Check backend logs
echo.
echo === Backend Logs (checking Asynq client) ===
docker-compose logs --tail=20 kg-backend

REM Step 6: Check worker logs  
echo.
echo === Worker Logs (checking if listening) ===
docker-compose logs --tail=20 kg-worker

REM Step 7: Check Redis
echo.
echo === Redis Health ===
docker-compose logs --tail=10 kg-redis

REM Step 8: Check Redis queue
echo.
echo === Checking Redis queue before test ===
docker exec kg-redis redis-cli KEYS "*asynq*" 2>nul || echo Redis not ready yet

REM Step 9: Create test note via curl
echo.
echo === Creating test note ===
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c%%a%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)
curl -X POST http://localhost:8080/notes ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Async Test %mydate%_%mytime%\",\"content\":\"Testing async task processing - keywords extraction and embedding computation\"}"

echo.
echo === Waiting 10 seconds for worker to process ===
timeout /t 10 /nobreak

REM Step 10: Check Redis queue after processing
echo.
echo === Redis queue after processing ===
docker exec kg-redis redis-cli KEYS "*asynq*" 2>nul || echo Checking Redis...
docker exec kg-redis redis-cli LLEN "asynq:queues:default" 2>nul || echo Queue check...

REM Step 11: Check worker logs for successful processing
echo.
echo === Worker Logs (checking for successful processing) ===
docker-compose logs --tail=50 kg-worker | findstr /C:"HandleExtractKeywords" /C:"HandleComputeEmbedding" /C:"successfully processed"

echo.
echo === DETAILED WORKER LOGS ===
docker-compose logs --tail=100 kg-worker

echo.
echo === Test Complete ===
