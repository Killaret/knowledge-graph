# Резервное копирование Knowledge Graph

Система резервного копирования для личного инстанса Knowledge Graph с поддержкой локального хранения и облачного бэкапа на Яндекс.Диск.

## 📋 Обзор системы

Система бэкапа включает:

- **Локальные скрипты**: `scripts/utility/backup-personal.sh` (Linux/Mac) и `scripts/utility/backup-personal.ps1` (Windows)
- **Go-сервис**: `backend/internal/infrastructure/cloud/yandex_backup.go` для работы с Яндекс.Диск через WebDAV
- **Asynq задача**: `TypeBackupToCloud` для асинхронной загрузки бэкапов
- **Docker сервис**: `backup_scheduler` в `docker-compose.personal.yml` для автоматического бэкапа
- **Конфигурация**: `knowledge-graph.config.json` секция `backup`

### Архитектура бэкапа

```
┌─────────────────────────────────────────────────────────────┐
│                   Backup Scheduler (Docker)                   │
│              - Ежедневный запуск по расписанию               │
│              - Выполнение backup-personal.sh                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Backup Script (backup-personal.sh/ps1)          │
│              - pg_dump базы PostgreSQL                        │
│              - Сжатие gzip                                   │
│              - Загрузка на Яндекс.Диск (опционально)          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│               Локальное хранение                             │
│               ./backups/backup-personal-YYYY-MM-DD.sql.gz    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼ (если включен облачный бэкап)
┌─────────────────────────────────────────────────────────────┐
│               Яндекс.Диск (WebDAV)                          │
│               /KnowledgeGraphBackups/                       │
└─────────────────────────────────────────────────────────────┘
```

## ⚙️ Настройка бэкапов

### Шаг 1: Настройка конфигурации

Отредактируйте файл `knowledge-graph.config.json`:

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": true,
      "provider": "yandex",
      "yandex": {
        "oauth_token": "your_oauth_token_here",
        "backup_folder": "/KnowledgeGraphBackups",
        "max_backups": 10
      }
    },
    "schedule": "0 2 * * *",
    "retention_days": 7,
    "draft_ttl_hours": 168
  }
}
```

**Параметры конфигурации:**

- `local_path` — локальная папка для хранения бэкапов
- `cloud.enabled` — включить облачный бэкап
- `cloud.provider` — провайдер облачного хранилища (только `yandex`)
- `cloud.yandex.oauth_token` — OAuth токен Яндекс.Диска
- `cloud.yandex.backup_folder` — папка на Яндекс.Диске для бэкапов
- `cloud.yandex.max_backups` — максимальное количество бэкапов в облаке
- `schedule` — расписание в формате cron (по умолчанию `0 2 * * *` — каждый день в 2:00)
- `retention_days` — количество дней для хранения локальных бэкапов
- `draft_ttl_hours` — время жизни черновиков в часах

### Шаг 2: Получение OAuth токена Яндекс.Диска

#### Создание приложения

1. Перейдите на [Yandex OAuth](https://oauth.yandex.ru/client/new)
2. Заполните форму:
   - **Название приложения**: Knowledge Graph Backup
   - **Описание**: Резервное копирование базы данных
   - **Ссылки на сайт**: `http://localhost`
   - **Redirect URI**: `https://oauth.yandex.ru/verification_code`
   - **Разрешения (DLS)**: выберите "Яндекс.Диск WebDAV API"
3. Получите **Client ID**

#### Получение токена через браузер

1. Откройте в браузере:
   ```
   https://oauth.yandex.ru/authorize?response_type=token&client_id=YOUR_CLIENT_ID
   ```
   Замените `YOUR_CLIENT_ID` на ваш Client ID
2. Авторизуйтесь в Яндекс
3. Разрешите доступ к Яндекс.Диску
4. В URL после перенаправления будет токен в формате:
   ```
   https://oauth.yandex.ru/verification_code#access_token=YOUR_TOKEN&...
   ```
5. Скопируйте токен (часть после `access_token=`)

#### Получение токена через curl

```bash
curl -X POST "https://oauth.yandex.ru/token" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

### Шаг 3: Настройка переменных окружения (опционально)

Добавьте в файл `.env`:

```bash
# Включить облачный бэкап
BACKUP_CLOUD_ENABLED=true

# OAuth токен Яндекс.Диска
BACKUP_YANDEX_TOKEN=your_oauth_token_here

# Папка на Яндекс.Диске
BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups

# Локальная папка для бэкапов
BACKUP_DIR=./backups

# Очистка старых бэкапов (true/false)
CLEANUP_OLD_BACKUPS=true
```

## 🚀 Ручной запуск бэкапа

### Windows (PowerShell)

```powershell
# Установите переменные окружения
$env:BACKUP_CLOUD_ENABLED = "true"
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token_here"
$env:BACKUP_YANDEX_FOLDER = "/KnowledgeGraphBackups"

# Запустите скрипт бэкапа
.\scripts\utility\backup-personal.ps1
```

### Linux/Mac

```bash
# Установите переменные окружения
export BACKUP_CLOUD_ENABLED=true
export BACKUP_YANDEX_TOKEN="your_oauth_token_here"
export BACKUP_YANDEX_FOLDER="/KnowledgeGraphBackups"

# Запустите скрипт бэкапа
./scripts/utility/backup-personal.sh
```

### Через Docker Compose (личный инстанс)

```bash
# Запуск backup_scheduler сервиса
docker-compose -f docker-compose.personal.yml up backup_scheduler

# Или единоразовый запуск бэкапа
docker-compose -f docker-compose.personal.yml run --rm backup_scheduler
```

## 📅 Автоматические бэкапы

### Через Docker Compose (рекомендуется)

Сервис `backup_scheduler` в `docker-compose.personal.yml` автоматически запускает бэкап каждые 24 часа:

```yaml
backup_scheduler:
  image: postgres:16-alpine
  container_name: kg-backup-scheduler
  volumes:
    - ./backups:/backups
    - ./scripts/utility:/scripts:ro
    - ./knowledge-graph.config.json:/config/config.json:ro
  environment:
    BACKUP_DIR: /backups
    BACKUP_CLOUD_ENABLED: ${BACKUP_CLOUD_ENABLED:-false}
    BACKUP_YANDEX_TOKEN: ${BACKUP_YANDEX_TOKEN:-}
    BACKUP_YANDEX_FOLDER: ${BACKUP_YANDEX_FOLDER:-/KnowledgeGraphBackups}
    PERSONAL_POSTGRES_HOST: postgres_personal
    PERSONAL_POSTGRES_PORT: 5432
    PERSONAL_POSTGRES_USER: ${PERSONAL_POSTGRES_USER:-personal}
    PERSONAL_POSTGRES_PASSWORD: ${PERSONAL_POSTGRES_PASSWORD:-personal_password}
    PERSONAL_POSTGRES_DB: ${PERSONAL_POSTGRES_DB:-knowledge_personal}
  command: >
    sh -c "
      apk add --no-cache curl gzip &&
      while true; do
        echo Running scheduled backup at $$(date) &&
        /scripts/backup-personal.sh &&
        echo Next backup in 24 hours &&
        sleep 86400
      done
    "
  depends_on:
    postgres_personal:
      condition: service_healthy
  restart: unless-stopped
```

### Через cron (Linux/Mac)

Добавьте в crontab (`crontab -e`):

```bash
# Ежедневный бэкап в 2:00 ночи
0 2 * * * cd /path/to/knowledge-graph && BACKUP_CLOUD_ENABLED=true BACKUP_YANDEX_TOKEN=your_token ./scripts/utility/backup-personal.sh >> /var/log/kg-backup.log 2>&1
```

### Через Task Scheduler (Windows)

1. Откройте Task Scheduler
2. Create Task → Triggers → Daily at 2:00 AM
3. Action: Start a program
   - Program: `powershell.exe`
   - Arguments: `-ExecutionPolicy Bypass -File "D:\knowledge-graph\scripts\utility\backup-personal.ps1"`
   - Add environment variables:
     - `BACKUP_CLOUD_ENABLED=true`
     - `BACKUP_YANDEX_TOKEN=your_token`
     - `BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups`

## 🔄 Восстановление из бэкапа

### Шаг 1: Скачивание бэкапа

**С Яндекс.Диска (веб-интерфейс):**

1. Откройте [Яндекс.Диск](https://disk.yandex.ru)
2. Перейдите в папку `KnowledgeGraphBackups`
3. Скачайте нужный бэкап: `backup-personal-YYYY-MM-DD.sql.gz`

**С Яндекс.Диска (через WebDAV):**

```bash
# Скачивание через curl
TOKEN="your_oauth_token"
BACKUP_DATE="2024-05-17"
curl -X GET "https://webdav.yandex.ru/KnowledgeGraphBackups/backup-personal-${BACKUP_DATE}.sql.gz" \
  -H "Authorization: OAuth ${TOKEN}" \
  --output backup-personal-${BACKUP_DATE}.sql.gz
```

**Из локального хранилища:**

```bash
# Бэкапы находятся в папке ./backups/
ls ./backups/
```

### Шаг 2: Распаковка бэкапа

```bash
gunzip backup-personal-2024-05-17.sql.gz
```

### Шаг 3: Восстановление в PostgreSQL

```bash
# Восстановление в базу personal
psql -h localhost -p 5433 -U personal -d knowledge_personal < backup-personal-2024-05-17.sql

# Или через Docker
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal < backup-personal-2024-05-17.sql
```

### Шаг 4: Проверка восстановления

```bash
# Подключитесь к базе и проверьте данные
docker exec -it kg-postgres-personal psql -U personal -d knowledge_personal

# Внутри psql
SELECT COUNT(*) FROM notes;
SELECT COUNT(*) FROM links;
\q
```

## 📊 Управление бэкапами

### Просмотр списка бэкапов

**Локальные бэкапы:**

```bash
ls -lh ./backups/
```

**Бэкапы на Яндекс.Диске (веб-интерфейс):**

1. Откройте [Яндекс.Диск](https://disk.yandex.ru)
2. Перейдите в папку `KnowledgeGraphBackups`

**Бэкапы на Яндекс.Диске (через API):**

Go-сервис `YandexBackupService` поддерживает метод `ListBackups` для получения списка файлов:

```go
files, err := service.ListBackups(ctx, "")
// files содержит список имен файлов бэкапов
```

### Удаление старых бэкапов

**Автоматическая очистка:**

Скрипты бэкапа автоматически удаляют локальные бэкапы старше 7 дней (настраивается через `CLEANUP_OLD_BACKUPS` и `retention_days`).

**Ручное удаление локальных бэкапов:**

```bash
# Удаление бэкапов старше 30 дней
find ./backups -name "backup-personal-*.sql.gz" -mtime +30 -delete
```

**Удаление бэкапов с Яндекс.Диска:**

1. Откройте [Яндекс.Диск](https://disk.yandex.ru)
2. Перейдите в папку `KnowledgeGraphBackups`
3. Удалите ненужные файлы

Go-сервис поддерживает метод `DeleteBackup` для программного удаления:

```go
err := service.DeleteBackup(ctx, "backup-personal-2024-05-17.sql.gz")
```

## 🔍 Устранение проблем

### Ошибка "BACKUP_YANDEX_TOKEN not set"

**Причина:** Переменная окружения `BACKUP_YANDEX_TOKEN` не установлена.

**Решение:**

```bash
# Linux/Mac
export BACKUP_YANDEX_TOKEN="your_token"
echo $BACKUP_YANDEX_TOKEN

# Windows PowerShell
$env:BACKUP_YANDEX_TOKEN = "your_token"
echo $env:BACKUP_YANDEX_TOKEN
```

### Ошибка "Yandex.Disk upload failed"

**Возможные причины:**

1. Токен неверный или истек
2. Недостаточно места на Яндекс.Диске
3. Проблемы с подключением к интернету
4. Неверная папка на Яндекс.Диске

**Решение:**

1. Проверьте токен и получите новый при необходимости
2. Проверьте свободное место на Яндекс.Диске (бесплатный тариф — 10 ГБ)
3. Проверьте подключение к интернету
4. Убедитесь, что папка `KnowledgeGraphBackups` существует или может быть создана

### Ошибка "pg_dump: command not found"

**Причина:** Утилита `pg_dump` не установлена или не доступна в PATH.

**Решение:**

```bash
# Установка PostgreSQL (включает pg_dump)
# Ubuntu/Debian
sudo apt-get install postgresql-client

# macOS
brew install postgresql

# Windows
# Скачайте PostgreSQL с https://www.postgresql.org/download/windows/
```

### Ошибка "Permission denied" при записи бэкапа

**Причина:** Недостаточно прав для записи в папку бэкапов.

**Решение:**

```bash
# Создайте папку бэкапов с нужными правами
mkdir -p ./backups
chmod 755 ./backups
```

### Ошибка 401 Unauthorized при загрузке на Яндекс.Диск

**Причина:** Токен неверный или истек.

**Решение:**

1. Получите новый OAuth токен по инструкции выше
2. Обновите `BACKUP_YANDEX_TOKEN` в `.env` файле
3. Перезапустите сервис бэкапа

### Бэкап создается локально, но не загружается в облако

**Причина:** Облачный бэкап отключен или настроен неправильно.

**Решение:**

```bash
# Проверьте, что облачный бэкап включен
echo $BACKUP_CLOUD_ENABLED  # должно быть "true"

# Проверьте токен
echo $BACKUP_YANDEX_TOKEN

# Проверьте конфигурацию в knowledge-graph.config.json
cat knowledge-graph.config.json | grep -A 10 backup
```

## 🔒 Безопасность

- **Никогда не коммитьте** `.env` файл с реальными токенами
- Храните токен в безопасном месте (менеджер паролей)
- Регулярно обновляйте OAuth токен
- Ограничьте права приложения только к Яндекс.Диску
- Используйте отдельные токены для разных окружений (dev, staging, prod)
- Шифруйте бэкапы при хранении в чувствительных окружениях
- Регулярно тестируйте восстановление из бэкапов

## 📈 Мониторинг и логирование

### Просмотр логов бэкапа

**Docker сервис:**

```bash
docker logs kg-backup-scheduler
```

**Локальный запуск:**

```bash
# Перенаправление вывода в файл
./scripts/utility/backup-personal.sh >> /var/log/kg-backup.log 2>&1

# Просмотр логов
tail -f /var/log/kg-backup.log
```

### Метрики бэкапа

Рекомендуется отслеживать:

- Частоту успешных бэкапов
- Размер бэкапов
- Время выполнения бэкапа
- Свободное место на Яндекс.Диске
- Количество хранимых бэкапов


## 📚 Дополнительная документация

- [Настройка Яндекс.Диск бэкапа (детально)](docs/YANDEX_DISK_BACKUP.md) — подробная инструкция по настройке Яндекс.Диска
- [Настройка облачного бэкапа](docs/CLOUD_BACKUP_SETUP.md) — общая информация о облачных бэкапах
- [Конфигурация системы](docs/CONFIGURATION.md) — полное руководство по настройке
- [Архитектура](docs/ARCHITECTURE.md) — архитектура системы бэкапа

## 🆘 Поддержка

Если вы столкнулись с проблемами при настройке или использовании системы бэкапа:

1. Проверьте этот документ на наличие решения
2. Изучите [docs/YANDEX_DISK_BACKUP.md](docs/YANDEX_DISK_BACKUP.md)
3. Проверьте логи системы
4. Создайте issue в репозитории проекта с описанием проблемы
