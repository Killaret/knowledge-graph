# Эффективное сжатие VHDX через Docker настройки

Write-Output "=== РЕКОМЕНДАЦИЯ: Сжатие VHDX через Docker настройки ==="

Write-Output "Проблема: VHDX файлы Docker/WSL2 не сжимаются обычными методами:"
Write-Output "- Compact.exe: работает только на уровне NTFS"
Write-Output "- Sparse файлы: VHDX уже оптимизирован"
Write-Output "- Системные инструменты: не работают с WSL2 VHDX"

Write-Output ""
Write-Output "РЕШЕНИЕ: Уменьшить лимит диска в Docker Desktop"
Write-Output ""
Write-Output "1. Откройте Docker Desktop"
Write-Output "2. Settings -> Resources -> Advanced"
Write-Output "3. Disk image size: уменьшить (например с 64GB до 32GB)"
Write-Output "4. Docker автоматически сожмет VHDX файл"
Write-Output ""
Write-Output "Ожидаемое сжатие: 16.76 GB -> ~8-12 GB (50-75% экономия)"
Write-Output ""
Write-Output "Это единственный РАБОТАЮЩИЙ метод для WSL2 VHDX файлов!"