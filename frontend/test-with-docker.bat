@echo off
echo 🐳 Запуск тестов с Docker окружением...

REM Проверяем что Docker запущен
docker-compose ps | findstr "kg-frontend.*Up" >nul
if errorlevel 1 (
    echo ❌ Frontend контейнер не запущен. Запустите: docker-compose up -d
    pause
    exit /b 1
)

echo ✅ Docker контейнеры запущены
echo 🧪 Запуск тестов Playwright...

cd frontend
set SKIP_AUTH=true
npx playwright test --config=playwright.config.docker.ts --reporter=list

echo ✅ Тесты завершены
pause
