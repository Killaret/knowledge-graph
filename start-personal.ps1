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
} else {
    Write-Host "❌ Ошибка при запуске. Проверьте логи командой выше." -ForegroundColor Red
    exit 1
}
