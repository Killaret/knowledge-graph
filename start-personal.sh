#!/bin/bash

# Скрипт для запуска личного экземпляра Knowledge Graph
# Запускает параллельный набор сервисов, не мешающий разработке

echo "Запуск личного экземпляра Knowledge Graph..."
docker compose -f docker-compose.personal.yml up -d --build

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Личный Knowledge Graph успешно запущен!"
    echo ""
    echo "🌐 Веб-интерфейс: http://localhost:3001"
    echo "🔌 API: http://localhost:8081"
    echo ""
    echo "Команды для управления:"
    echo "  Остановить:  docker compose -f docker-compose.personal.yml stop"
    echo "  Перезапустить: docker compose -f docker-compose.personal.yml restart"
    echo "  Логи:        docker compose -f docker-compose.personal.yml logs -f"
    echo "  Удалить:     docker compose -f docker-compose.personal.yml down"
else
    echo "❌ Ошибка при запуске. Проверьте логи командой выше."
    exit 1
fi
