# Docker Cleanup Fixes

## Исправленные проблемы в clean_and_compress_lunix скриптах

### 🔧 Что было исправлено:

#### 1. **Поиск нескольких файлов вместо одного**
**Проблема:** Скрипты находили только первый lunix файл и прекращали поиск.

**Решение:**
- Теперь скрипты находят **все** lunix файлы в разных директориях
- Поддерживается поиск в нескольких типовых директориях:
  - `$env:USERPROFILE` (пользовательская папка)
  - `$env:LOCALAPPDATA\Docker\wsl\`
  - `$env:LOCALAPPDATA\Packages\`
  - `D:\`, `C:\` диски
  - Произвольные директории через параметр `-ImagePath`

#### 2. **Поддержка директорий**
**Проблема:** Нельзя было указать директорию для поиска всех файлов внутри.

**Решение:**
- Параметр `-ImagePath` теперь принимает как файл, так и директорию
- Если указана директория, скрипт ищет все `*lunix*` файлы внутри
- Пример: `.\clean_and_compress_lunix.ps1 -ImagePath "D:\images\" -Compress`

#### 3. **Сжатие диска Windows Compact.exe**
**Проблема:** Сжатие не работало на системах без Hyper-V.

**Решение:**
- Добавлен параметр `-UseCompact` для использования **Windows Compact.exe**
- Compact.exe — встроенная утилита Windows, не требует Hyper-V
- Сжимает файлы на уровне файловой системы
- Теперь используется по умолчанию в `npm run clean:lunix`

#### 4. **Sparse файлы оптимизация**
**Проблема:** Не использовались возможности sparse файлов для экономии диска.

**Решение:**
- Добавлена поддержка `fsutil sparse` для Windows
- Автоматически пытается включить sparse атрибут на lunix файлах
- Для Linux/WSL используется `chattr +s` для sparse файлов
- Sparse файлы занимают меньше места на диске при наличии пустых блоков

#### 5. **Улучшенный обработчик ошибок**
**Проблема:** Скрипты могли падать при ошибке с одним файле.

**Решение:**
- Каждый файл обрабатывается независимо
- Ошибка с одним файлом не останавливает обработку остальных
- Детальные статусы для каждого этапа обработки

## 🚀 Новые возможности

### PowerShell скрипт

```powershell
# Базовое использование с Compact.exe (рекомендуется)
.\clean_and_compress_lunix.ps1 -Search -Compress -UseCompact

# С использованием Optimize-VHD (требует Hyper-V)
.\clean_and_compress_lunix.ps1 -Search -Compress

# Обработка всех файлов в директории
.\clean_and_compress_lunix.ps1 -ImagePath "D:\images\" -Compress -Force

# Dry run для проверки
.\clean_and_compress_lunix.ps1 -Search -Compress -UseCompact -DryRun
```

### Bash скрипт

```bash
# Поиск и сжатие всех lunix файлов
./clean_and_compress_lunix.sh -s -c

# Обработка директории
./clean_and_compress_lunix.sh -p /path/to/dir -c -f

# Dry run
./clean_and_compress_lunix.sh -s -c --dry-run
```

### NPM команды

```bash
# С Compact.exe (рекомендуется, работает без Hyper-V)
npm run clean:lunix

# С Optimize-VHD (требует Hyper-V)
npm run clean:lunix:vhd

# Dry run
npm run clean:lunix:dry

# Bash версия
npm run clean:lunix:sh
```

## 📋 Сравнение методов сжатия

| Метод | Требования | Преимущества | Недостатки |
|-------|------------|--------------|------------|
| **Compact.exe** | Windows встроенный | Работает везде, не требует Hyper-V | Менее эффективно для VHD |
| **Optimize-VHD** | Hyper-V | Максимальное сжатие VHD | Требует Hyper-V Pro/Enterprise |
| **qemu-img** | qemu-utils | Кроссплатформенный, мощный | Требует установки |
| **Sparse файлы** | fsutil/chattr | Автоматическая оптимизация | Зависит от типа данных |

## 🔍 Поисковые директории

### Windows (PowerShell)
Скрипт ищет в:
- `%USERPROFILE%\lunix.vhdx`
- `%USERPROFILE%\lunix.img`
- `%LOCALAPPDATA%\Docker\wsl\lunix.vhdx`
- `%LOCALAPPDATA%\Docker\wsl\disk\lunix.vhdx`
- `%LOCALAPPDATA%\Docker\wsl\data\`
- `%LOCALAPPDATA%\Packages\`
- `D:\`, `C:\` диски
- Рекурсивно в пользовательском профиле при необходимости

### Linux/WSL (Bash)
Скрипт ищет в:
- `~/lunix.vhdx`, `~/lunix.img`
- `~/.docker/wsl/lunix.vhdx`
- `/mnt/wsl/docker-desktop/lunix.vhdx`
- `/mnt/c/Users/$USER/lunix.vhdx`
- `/d/`, `/c/` монтированные диски
- Docker WSL директории

## 🎯 Рекомендации

### Для Windows пользователей:
1. **Используйте `npm run clean:lunix`** (Compact.exe по умолчанию)
2. Если есть Hyper-V Pro/Enterprise, попробуйте `npm run clean:lunix:vhd`
3. Для проверки используйте `npm run clean:lunix:dry`

### Для Linux/WSL пользователей:
1. Используйте `npm run clean:lunix:sh`
2. Установите `qemu-utils` для лучшего сжатия: `sudo apt-get install qemu-utils`
3. Используйте `--dry-run` для предварительной проверки

### Для автоматизации:
1. Регистрация через `scripts/register_cleanup_task.ps1` (Windows)
2. Добавление в cron через `scripts/register_cron.sh` (Linux/WSL)

## 📊 Типичная экономия места

- **Compact.exe:** 5-15% для обычных файлов
- **Optimize-VHD:** 15-40% для VHD файлов с пустыми блоками
- **Sparse файлы:** 10-30% в зависимости от контента
- **qemu-img:** 20-50% при конвертации в qcow2 с сжатием

## ⚠️ Важные замечания

1. **Сначала dry run:** Всегда используйте `-DryRun` или `--dry-run` для проверки
2. **Резервные копии:** Создайте бэкап перед сжатием важных данных
3. **Время выполнения:** Сжатие больших файлов (>10GB) может занять время
4. **Права доступа:** Убедитесь что есть права на запись к файлам
5. **Hyper-V ограничения:** Optimize-VHD работает только с Windows Pro/Enterprise с Hyper-V

## 🐛 Устранение проблем

### "Hyper-V feature not enabled"
**Решение:** Используйте `-UseCompact` параметр или `npm run clean:lunix`

### "File not accessible"
**Решение:** Запустите PowerShell от администратора или проверьте права доступа

### "No lunix images found"
**Решение:** Укажите конкретный путь через `-ImagePath` параметр

### "Compact.exe failed"
**Решение:** Проверьте что файл не используется другими процессами (Docker, WSL)

### "qemu-img not found"
**Решение:** Установите qemu-utils: `sudo apt-get install qemu-utils`