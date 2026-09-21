# Скрипт для запуска личного экземпляра Knowledge Graph
# Запускает параллельный набор сервисов, не мешающий разработке

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
