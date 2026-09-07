#!/bin/bash

# Скрипт для запуска тестов с Docker окружением
echo "🐳 Запуск тестов с Docker окружением..."

# Убедимся что Docker запущен
if ! docker-compose ps | grep -q "kg-frontend.*Up"; then
    echo "❌ Frontend контейнер не запущен. Запустите: docker-compose up -d"
    exit 1
fi

echo "✅ Docker контейнеры запущены"
echo "🧪 Запуск smoke тестов..."

# Запускаем тесты с правильной конфигурацией
cd frontend
SKIP_AUTH=true npx playwright test \
    --config=playwright.config.docker.ts \
    --grep="@smoke" \
    --timeout=30000 \
    --reporter=list

echo "✅ Тесты завершены"
