# Настройка облачного резервного копирования

> **⚠️ Устаревший документ:** Данный документ описывает настройку Cloudflare R2, которая является устаревшей системой бэкапа. Для личного инстанса теперь используется Яндекс.Диск. См. актуальную документацию: [`docs/BACKUP.md`](docs/BACKUP.md) и [`docs/YANDEX_DISK_BACKUP.md`](docs/YANDEX_DISK_BACKUP.md).

В проекте уже реализована инфраструктура для резервного копирования базы данных в облачное хранилище Cloudflare R2.

## Обзор

Система резервного копирования включает:
- **Локальные скрипты**: `scripts/devops/backup-personal.ps1` (Windows) и `scripts/devops/backup-personal.sh` (Linux/Mac)
- **Backend сервис**: `backend/internal/infrastructure/cloud/r2_backup.go` для работы с Cloudflare R2
- **API endpoint**: `/api/v1/admin/backup/cloud` для запуска облачного бэкапа
- **Асинхронная очередь**: задачи загружаются в очередь через Asynq

## Способы настройки

### Способ 1: Через knowledge-graph.config.json (рекомендуется)

Отредактируйте файл `knowledge-graph.config.json`:

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": true,
      "provider": "r2",
      "r2": {
        "account_id": "your_account_id",
        "access_key_id": "your_access_key_id",
        "secret_access_key": "your_secret_access_key",
        "bucket": "your_bucket_name",
        "region": "auto"
      }
    },
    "schedule": "0 2 * * *",
    "retention_days": 7,
    "draft_ttl_hours": 168
  }
}
```

### Способ 2: Через переменные окружения

Добавьте в файл `.env`:

```bash
# Включить облачный бэкап
BACKUP_CLOUD_ENABLED=true

# Провайдер (пока только r2)
BACKUP_CLOUD_PROVIDER=r2

# Локальный путь для бэкапов
BACKUP_LOCAL_PATH=./backups

# Расписание в формате cron (0 2 * * * = каждый день в 2:00)
BACKUP_SCHEDULE=0 2 * * *

# Хранить бэкапы 7 дней
BACKUP_RETENTION_DAYS=7

# Cloudflare R2 настройки
BACKUP_R2_ACCOUNT_ID=your_account_id
BACKUP_R2_ACCESS_KEY_ID=your_access_key_id
BACKUP_R2_SECRET_ACCESS_KEY=your_secret_access_key
BACKUP_R2_BUCKET=your_bucket_name
BACKUP_R2_REGION=auto
```

## Настройка Cloudflare R2

### 1. Создание аккаунта и bucket

1. Зарегистрируйтесь на [Cloudflare](https://dash.cloudflare.com/sign-up)
2. Перейдите в раздел **R2** (Zero Trust → R2)
3. Создайте новый bucket (например, `knowledge-graph-backups`)

### 2. Получение API ключей

1. В разделе R2 нажмите **Manage R2 API Tokens**
2. Создайте новый API Token с правами:
   - **Object Read & Write**
3. Сохраните полученные данные:
   - **Account ID** (виден в правом верхнем углу дашборда R2)
   - **Access Key ID**
   - **Secret Access Key**

### 3. Настройка CORS (опционально)

Если нужно доступ к бэкапам из браузера, настройте CORS в настройках bucket:

```json
{
  "AllowedOrigins": [
    "https://your-domain.com"
  ],
  "AllowedMethods": [
    "GET",
    "HEAD"
  ],
  "AllowedHeaders": [
    "*"
  ],
  "MaxAgeSeconds": 86400
}
```

## Использование

### Ручной запуск бэкапа (Windows)

```powershell
# Установите переменную окружения для включения облачного бэкапа
$env:CLOUD_BACKUP_ENABLED = "true"

# Запустите скрипт бэкапа
.\scripts\devops\backup-personal.ps1
```

### Ручной запуск бэкапа (Linux/Mac)

```bash
# Установите переменную окружения для включения облачного бэкапа
export CLOUD_BACKUP_ENABLED=true

# Запустите скрипт бэкапа
./scripts/devops/backup-personal.sh
```

### Через API endpoint

После настройки backend, можно вызвать API для запуска облачного бэкапа:

```bash
curl -X POST http://localhost:18085/api/v1/admin/backup/cloud \
  -H "Content-Type: application/json" \
  -d '{"local_path": "./backups/backup-personal-2024-05-17.sql.gz"}'
```

## Автоматизация через cron

### Linux/Mac

Добавьте в crontab (`crontab -e`):

```bash
# Ежедневный бэкап в 2:00 ночи
0 2 * * * cd /path/to/knowledge-graph && CLOUD_BACKUP_ENABLED=true ./scripts/devops/backup-personal.sh
```

### Windows (Task Scheduler)

1. Откройте Task Scheduler
2. Create Task → Triggers → Daily at 2:00 AM
3. Action: Start a program
   - Program: `powershell.exe`
   - Arguments: `-ExecutionPolicy Bypass -File "D:\knowledge-graph\scripts\devops\backup-personal.ps1"`
   - Add environment variable: `CLOUD_BACKUP_ENABLED=true`

## Проверка бэкапов

### Просмотр списка бэкапов в R2

Используйте AWS CLI или Cloudflare R2 CLI:

```bash
# Установка wrangler (Cloudflare CLI)
npm install -g wrangler

# Авторизация
wrangler login

# Просмотр файлов в bucket
wrangler r2 object list knowledge-graph-backups
```

### Восстановление из бэкапа

```bash
# Скачивание бэкапа из R2
wrangler r2 object get knowledge-graph-backups/backups/backup-personal-2024-05-17.sql.gz --file=backup.sql.gz

# Распаковка
gunzip backup.sql.gz

# Восстановление в PostgreSQL
psql -h localhost -p 5433 -U personal -d knowledge_personal < backup.sql
```

## Устранение проблем

### Ошибка "R2 configuration is incomplete"

Убедитесь, что все обязательные поля заполнены:
- `BACKUP_R2_ACCOUNT_ID`
- `BACKUP_R2_ACCESS_KEY_ID`
- `BACKUP_R2_SECRET_ACCESS_KEY`
- `BACKUP_R2_BUCKET`

### Ошибка подключения к R2

- Проверьте, что Account ID правильный (виден в дашборде R2)
- Убедитесь, что API Token имеет нужные права
- Проверьте, что регион установлен в "auto" или правильный регион

### Бэкап создается локально, но не загружается в облако

- Убедитесь, что `BACKUP_CLOUD_ENABLED=true`
- Проверьте логи backend на наличие ошибок
- Убедитесь, что backend запущен и worker обрабатывает задачи

## Безопасность

- **Никогда не коммитьте** `.env` файл с реальными ключами
- Используйте отдельные API токены для разных окружений
- Ограничьте права API токена только необходимыми операциями
- Регулярно ротируйте ключи доступа
- Включите шифрование на стороне R2 (опция в настройках bucket)

## Альтернативные провайдеры

В будущем можно добавить поддержку других S3-совместимых хранилищ:
- AWS S3
- Google Cloud Storage
- Azure Blob Storage
- MinIO (self-hosted)

Для этого нужно создать новый сервис в `backend/internal/infrastructure/cloud/` по аналогии с `r2_backup.go`.
