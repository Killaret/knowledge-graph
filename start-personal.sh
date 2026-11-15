#!/bin/bash

# Скрипт для запуска личного экземпляра Knowledge Graph
# Запускает параллельный набор сервисов, не мешающий разработке

# BACKUP-4 item 3: the stack must start from D:\knowledge-graph only.
# Compose derives the project name from the directory name, and named
# volumes get the <project>_ prefix: started from another worktree the
# stack silently runs on empty volumes and its backups overwrite the real
# ones with fresh empties (2026-09-30 incident).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="${COMPOSE_PROJECT_NAME:-$(basename "$SCRIPT_DIR")}"
if [ "$PROJECT_NAME" != "knowledge-graph" ]; then
    echo "❌ Personal-стек запускается только из D:\\knowledge-graph." >&2
    echo "   Сейчас имя проекта Compose было бы '$PROJECT_NAME' — тома получили бы префикс '${PROJECT_NAME}_'," >&2
    echo "   и стек стартанул бы на пустых данных, а не на данных владельца." >&2
    exit 1
fi

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
