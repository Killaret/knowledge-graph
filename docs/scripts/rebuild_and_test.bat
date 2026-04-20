@echo off
cd /d D:\knowledge-graph

echo Stopping old containers...
docker-compose down

echo Building and starting all services...
docker-compose up --build -d

echo Waiting 30 seconds for services to initialize...
timeout /t 30 /nobreak

echo.
echo ===== Backend logs =====
docker-compose logs --tail=20 kg-backend

echo.
echo ===== Worker logs =====
docker-compose logs --tail=50 kg-worker

echo.
echo ===== Checking container status =====
docker-compose ps

echo.
echo Rebuild complete! Now testing async task processing...
pause
