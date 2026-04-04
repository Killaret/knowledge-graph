@echo off
REM Simple status check script

cd /d D:\knowledge-graph

cls
echo ========================================
echo Knowledge Graph - Quick Status Check
echo ========================================
echo.

echo [1] Docker version:
docker version --format "Client Version: {{.Client.Version}}"
echo.

echo [2] Services status:
docker-compose ps
echo.

echo [3] Backend service (is it running?):
docker-compose logs --tail=5 kg-backend
echo.

echo [4] Worker service (is it running?):
docker-compose logs --tail=5 kg-worker
echo.

echo [5] Redis queue size:
docker exec kg-redis redis-cli LLEN "asynq:queues:default" 2>nul
echo.

echo [6] Recent errors in logs:
docker-compose logs kg-worker | findstr /C:"error" /C:"ERROR" /C:"failed" || echo "No errors found"
echo.

echo Done!
pause
