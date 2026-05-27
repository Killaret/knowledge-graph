# Scripts Reorganization

## Overview
Скрипты проекта были реорганизованы по семантическим категориям для улучшения структуры и удобства навигации.

## New Structure

```
scripts/
├── cleanup/          # Скрипты очистки и сжатия
├── diagnostics/      # Диагностические и проверочные скрипты
├── testing/          # Тестовые скрипты и эксперименты
├── utility/          # Вспомогательные скрипты
├── database/         # Скрипты для работы с базой данных
├── docs/             # Документация скриптов
├── build-config.js   # Конфигурация сборки (корневой)
└── clean_and_compress_lunix.sh # Bash версия (корневой)
```

## Categories

### Cleanup (`scripts/cleanup/`)
Основные скрипты для очистки и оптимизации системы:
- `diskpart_compress_admin.ps1` - VHDX сжатие через DiskPart с авто-разблокировкой
- `clean_and_compress_lunix.ps1` - Сжатие lunix образов
- `clean_and_compress_lunix.sh` - Bash версия сжатия lunix
- `cleanup_and_compress.ps1` - Полная очистка Docker и сжатие
- `cleanup-docker.ps1` - Очистка Docker
- `cleanup-docker.sh` - Bash версия очистки Docker
- `register_cleanup_task.ps1` - Регистрация задачи очистки
- `register_cron.sh` - Настройка cron для очистки
- `run_cleanup.bat` - Bat файл для запуска очистки

### Diagnostics (`scripts/diagnostics/`)
Скрипты для диагностики и проверки системы:
- `check_all_vhdx.ps1` - Проверка размеров всех VHDX файлов
- `check_disk_lock.ps1` - Проверка блокировок диска и процессов
- `check_docker_usage.ps1` - Проверка использования диска Docker
- `check_file_lock.ps1` - Простая проверка блокировки файла
- `check_vhdx_size_simple.ps1` - Простая проверка размера VHDX
- `find_docker.ps1` - Поиск Docker процессов

### Testing (`scripts/testing/`)
Тестовые скрипты и экспериментальные решения:
- `try_diskpart.ps1` - Тест DiskPart сжатия
- `try_diskpart_advanced.ps1` - Расширенный тест DiskPart
- `try_diskpart_shrink.ps1` - Тест DiskPart shrink
- `try_diskpart_volume.ps1` - Тест DiskPart volume
- `try_optimize_vhd.ps1` - Тест Optimize-VHD
- `try_vhdx_compression.ps1` - Тест VHDX сжатия
- `try_windows_tools.ps1` - Тест Windows инструментов
- `test_cleanup.ps1` - Тест скриптов очистки
- `open_diskpart_manual.ps1` - Ручное открытие DiskPart
- `run-full-tests.sh` - Запуск всех тестов

### Utility (`scripts/utility/`)
Вспомогательные скрипты:
- `stop_docker.ps1` - Остановка Docker
- `force_stop_docker.ps1` - Принудительная остановка Docker
- `fix_vhdx_attributes.ps1` - Исправление атрибутов VHDX
- `backup-personal.ps1` - Персональное резервирование
- `backup-personal.sh` - Bash версия резервирования
- `lint-scripts.ps1` - Линтинг PowerShell скриптов
- `lint-scripts.sh` - Bash версия линтинга

### Database (`scripts/database/`)
Скрипты для работы с базой данных:
- `seed_data.py` - Заполнение тестовыми данными
- `setup_test_db.sql` - Настройка тестовой базы данных

### Docs (`scripts/docs/`)
Документация скриптов и процессов:
- `VHDX_COMPRESSION_FINAL.md` - Финальная документация VHDX сжатия
- `VHDX_COMPRESSION_SOLUTION.md` - Решение для VHDX сжатия
- `CLEANUP_FIXES.md` - Исправления скриптов очистки
- `RECOMMENDATION.md` - Рекомендации по использованию
- `SCRIPT_READY.md` - Статус готовности скриптов
- `TEST_RESULTS.md` - Результаты тестирования

## Removed Files

Удалены следующие временные и дублирующиеся файлы:
- `check_vhdx_before.ps1` - дубликат
- `check_vhdx_size.ps1` - дубликат
- `check_vhdx_attributes.ps1` - устаревший
- `fix_attrib.ps1` - устаревший
- `fix_compression.ps1` - устаревший
- `stop_and_fix.ps1` - временный
- `stop_and_sdelete.ps1` - временный
- `check_hyper_v.ps1` - не используется
- `cleanup_with_log.ps1` - временный
- `diskpart_admin.ps1` - старая версия
- `try_diskpart_vdisk*.ps1` - старые версии
- `cleanup_log.txt` - временный лог
- `diskpart_vdisk.txt` - временный файл

## Updated Paths

### Package Scripts
Обновлены пути в `package.json`:
```json
"clean:lunix": "./scripts/cleanup/clean_and_compress_lunix.ps1"
"clean:lunix:vhd": "./scripts/cleanup/clean_and_compress_lunix.ps1"
"clean:lunix:dry": "./scripts/cleanup/clean_and_compress_lunix.ps1"
"clean:lunix:sh": "./scripts/cleanup/clean_and_compress_lunix.sh"
"clean:lunix:sh:dry": "./scripts/cleanup/clean_and_compress_lunix.sh"
"clean:docker:vhdx": "./scripts/cleanup/diskpart_compress_admin.ps1"
```

### Documentation
Обновлены пути в `COMMANDS.md`:
- Все скрипты cleanup теперь в `scripts/cleanup/`
- Все диагностические скрипты в `scripts/diagnostics/`
- Основной скрипт VHDX сжатия: `scripts/cleanup/diskpart_compress_admin.ps1`

## GitIgnore Updates

Добавлены правила для игнорирования временных файлов скриптов:
```
# Scripts temporary files
scripts/*.txt
scripts/*.log
scripts/diskpart_*.txt
scripts/cleanup_*.txt
scripts/temp_*
scripts/tmp_*

# Script test artifacts
scripts/testing/temp/
scripts/testing/*.log
scripts/diagnostics/*.log

# Database seed files (keep structure, ignore generated data)
scripts/database/*.db
scripts/database/*.sqlite
scripts/database/generated/
```

## Migration Guide

### Для пользователей:
1. Обновите команды в `COMMANDS.md` - пути изменились
2. Используйте npm скрипты - они автоматически обновлены
3. Прямые вызовы скриптов требуют обновления путей

### Для разработчиков:
1. Новые скрипты размещайте в соответствующих категориях
2. Временные скрипты в `scripts/testing/`
3. Диагностические скрипты в `scripts/diagnostics/`
4. Основные скрипты очистки в `scripts/cleanup/`

## Benefits

1. **Организация** - скрипты легко найти по назначению
2. **Поддержка** - понятная структура для разработки
3. **Чистота** - удалены временные и дублирующиеся файлы
4. **Git** - правила игнорирования для временных файлов
5. **Документация** - централизованная документация скриптов