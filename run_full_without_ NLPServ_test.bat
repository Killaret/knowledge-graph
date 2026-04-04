@echo off
setlocal enabledelayedexpansion

title Knowledge Graph - Async Testing

echo.
echo ========================================
echo Knowledge Graph - Full Async Test
echo ========================================
echo.

cd /d D:\knowledge-graph

:: Создаём папку для кэша, если её нет
:: if not exist "huggingface_cache" mkdir huggingface_cache

:: Копируем локальную модель из пользовательского кэша (если есть)
:: if exist "%USERPROFILE%\.cache\huggingface\models--sentence-transformers--all-MiniLM-L6-v2" (
    :: echo [0] Copying cached model from local Hugging Face cache...
   ::  xcopy /E /I /Y "%USERPROFILE%\.cache\huggingface\models--sentence-transformers--all-MiniLM-L6-v2" "huggingface_cache\models--sentence-transformers--all-MiniLM-L6-v2"
:: ) else (
   ::  echo [0] No local cache found. Model will be downloaded during first start.
:: )

:: echo [1] Stopping old containers...
:: docker-compose down
:: echo.

:: echo [2] Building and starting services...
:: docker-compose build
:: docker-compose up -d
:: echo.

:: echo [3] Waiting for services to initialize (60 seconds)...
:: timeout /t 60 /nobreak

:: Ожидание готовности NLP
echo [3a] Checking NLP service status...
set NLP_READY=0
for /l %%i in (1,1,100) do (
    curl -s http://localhost:5000/health >nul 2>&1
    if !errorlevel! equ 0 (
        set NLP_READY=1
        echo NLP is ready.
        goto :nlp_ready
    )
    echo Waiting for NLP... %%i/100
    timeout /t 3 /nobreak >nul
)
:nlp_ready

if !NLP_READY! equ 1 (
    :: Копируем модель из контейнера в хост (если она ещё не скопирована)
    echo Copying model from container to host cache...
    docker cp kg-nlp:/root/.cache/huggingface/. D:\knowledge-graph\huggingface_cache\ 2>nul
    echo Model cache updated.
) else (
    echo WARNING: NLP service did not become ready within timeout.
    echo Showing NLP logs:
    docker-compose logs --tail=50 kg-nlp
)

:: Вывод информации о кэше модели
echo.
echo [3b] Model cache information:
dir D:\knowledge-graph\huggingface_cache
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

if !NLP_READY! equ 1 (
    echo [9] Creating test note (triggering async tasks)...
    echo ===================================================
    curl -X POST http://localhost:8080/notes ^
      -H "Content-Type: application/json" ^
      -d "{\"title\":\"Async Test Note\",\"content\":\"This note triggers keyword extraction and embedding computation async tasks\"}"
    echo.
    echo.
) else (
    echo [9] Skipping note creation because NLP is not ready.
    goto :skip_create
)

echo [10] Waiting 20 seconds for worker to process tasks...
timeout /t 20 /nobreak
echo.

:skip_create
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