# Скрипт для запуска личного экземпляра Knowledge Graph
# Запускает параллельный набор сервисов, не мешающий разработке

# BACKUP-4 item 3: the stack must start from D:\knowledge-graph only.
# Compose derives the project name from the directory name, and named
# volumes get the <project>_ prefix: started from another worktree the
# stack silently runs on empty volumes and its backups overwrite the real
# ones with fresh empties (2026-09-30 incident).
$projectName = $env:COMPOSE_PROJECT_NAME
if (-not $projectName) { $projectName = Split-Path $PSScriptRoot -Leaf }
if ($projectName -ne "knowledge-graph") {
    Write-Host "❌ Personal-стек запускается только из D:\knowledge-graph." -ForegroundColor Red
    Write-Host "   Сейчас имя проекта Compose было бы '$projectName' — тома получили бы префикс '${projectName}_'," -ForegroundColor Red
    Write-Host "   и стек стартанул бы на пустых данных, а не на данных владельца." -ForegroundColor Red
    exit 1
}

Write-Host "Запуск личного экземпляра Knowledge Graph..." -ForegroundColor Cyan
docker compose -f docker-compose.personal.yml up -d --build

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Личный Knowledge Graph успешно запущен!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🌐 Веб-интерфейс: http://localhost:3001" -ForegroundColor Yellow
    Write-Host "🔌 API: http://localhost:8081" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Команды для управления:" -ForegroundColor Gray
    Write-Host "  Остановить:  docker compose -f docker-compose.personal.yml stop" -ForegroundColor Gray
    Write-Host "  Перезапустить: docker compose -f docker-compose.personal.yml restart" -ForegroundColor Gray
    Write-Host "  Логи:        docker compose -f docker-compose.personal.yml logs -f" -ForegroundColor Gray
    Write-Host "  Удалить:     docker compose -f docker-compose.personal.yml down" -ForegroundColor Gray

    # BACKUP-3: страховка на старте. Если свежайший бэкап старше
    # KG_BACKUP_MAX_AGE_HOURS (backup-policy.env), делаем daily — cron в 02:00
    # срабатывает только при поднятом стеке, а событийный путь воркера ловит
    # только новые изменения.
    Write-Host ""
    Write-Host "Проверка свежести бэкапа:" -ForegroundColor Cyan
    & "$PSScriptRoot\scripts\devops\check-personal-backup.ps1"
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  Свежего бэкапа нет — запускаю daily..." -ForegroundColor Yellow
        & "$PSScriptRoot\scripts\devops\backup-personal.ps1" -Mode daily
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  Страховочный бэкап создан." -ForegroundColor Green
        } else {
            Write-Host "  Страховочный бэкап не удался — сделайте вручную: scripts\devops\backup-personal.ps1" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "❌ Ошибка при запуске. Проверьте логи командой выше." -ForegroundColor Red
    exit 1
}
