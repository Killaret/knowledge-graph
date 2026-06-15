# Настройка резервного копирования на Яндекс.Диск

В проекте реализована поддержка резервного копирования базы данных на Яндекс.Диск через REST API.

## Обзор

Система резервного копирования на Яндекс.Диск включает:
- **Локальные скрипты**: `scripts/backup-personal.ps1` (Windows) и `scripts/backup-personal.sh` (Linux/Mac)
- **REST API**: прямая загрузка файлов на Яндекс.Диск через OAuth токен
- **Автоматическая очистка**: удаление старых бэкапов (по умолчанию 7 дней)

## Требования к месту

Для Knowledge Graph обычно требуется:
- **Текстовые данные (notes)**: несколько МБ
- **Embeddings**: может занимать больше места, обычно до 100-500 МБ
- **Ежедневные бэкапы за неделю**: ~1-3 ГБ

**Бесплатный тариф Яндекс.Диска**: 10 ГБ - этого более чем достаточно для личного использования.

## Настройка

### Способ 1: Через knowledge-graph.config.json (рекомендуется)

Отредактируйте файл `knowledge-graph.config.json`:

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": true,
      "provider": "yandex",
      "yandex": {
        "oauth_token": "your_oauth_token_here"
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

# Провайдер - Яндекс.Диск
BACKUP_CLOUD_PROVIDER=yandex

# Локальный путь для бэкапов
BACKUP_LOCAL_PATH=./backups

# Расписание в формате cron (0 2 * * * = каждый день в 2:00)
BACKUP_SCHEDULE=0 2 * * *

# Хранить бэкапы 7 дней
BACKUP_RETENTION_DAYS=7

# Время жизни черновиков в часах (168 = 7 дней)
BACKUP_DRAFT_TTL_HOURS=168

# OAuth токен Яндекс.Диска
BACKUP_YANDEX_TOKEN=your_oauth_token_here
```

## Настройка периодичности бэкапов

Периодичность бэкапов настраивается через параметр `schedule` в формате cron.

### Формат cron

```
минута час день_месяца месяц день_недели
*      *   *           *      *
```

- **минута**: 0-59
- **час**: 0-23
- **день_месяца**: 1-31
- **месяц**: 1-12
- **день_недели**: 0-7 (0 или 7 = воскресенье)

### Примеры расписаний

```json
"schedule": "0 2 * * *"           // Каждый день в 2:00
"schedule": "0 */6 * * *"         // Каждые 6 часов
"schedule": "0 0 * * 0"           // Каждое воскресенье в полночь
"schedule": "0 0 1 * *"           // 1-го числа каждого месяца в полночь
"schedule": "*/30 * * * *"        // Каждые 30 минут
"schedule": "0 12,18 * * *"       // Каждый день в 12:00 и 18:00
```

### Настройка через конфигурационный файл

```json
{
  "backup": {
    "schedule": "0 2 * * *",
    "retention_days": 7
  }
}
```

### Настройка через переменные окружения

```bash
BACKUP_SCHEDULE=0 2 * * *
BACKUP_RETENTION_DAYS=7
```

### Параметры хранения

- **retention_days**: Количество дней хранения бэкапов (по умолчанию 7)
- **draft_ttl_hours**: Время жизни черновиков в часах (по умолчанию 168 = 7 дней)

### Проверка целостности бэкапов

Система автоматически проверяет целостность загруженных бэкапов с помощью SHA256 хэш-сумм:

1. Вычисляется хэш локального файла перед загрузкой
2. После загрузки файл скачивается во временную директорию
3. Вычисляется хэш загруженного файла
4. Хэши сравниваются для подтверждения целостности

Если хэши не совпадают, загрузка считается неудачной и возвращается ошибка.

## Получение OAuth токена Яндекс.Диска

### 1. Создание приложения

1. Перейдите на [Yandex OAuth](https://oauth.yandex.ru/client/new)
2. Заполните форму:
   - **Название приложения**: Knowledge Graph Backup
   - **Описание**: Резервное копирование базы данных
   - **Ссылки на сайт**: `http://localhost`
   - **Redirect URI**: `https://oauth.yandex.ru/verification_code`
   - **Разрешения (DLS)**: выберите:
     - `cloud_api:disk.write` - Запись в любом месте на Диске
     - `cloud_api:disk.read` - Чтение всего Диска
3. Получите **Client ID**

### 2. Получение токена

**Способ А: Через браузер (проще)**

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

**Способ Б: Через curl**

```bash
curl -X POST "https://oauth.yandex.ru/token" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=YOUR_REFRESH_TOKEN" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

## Использование

### Ручной запуск бэкапа (Windows)

```powershell
# Установите переменные окружения
$env:BACKUP_CLOUD_ENABLED = "true"
$env:BACKUP_CLOUD_PROVIDER = "yandex"
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token_here"

# Запустите скрипт бэкапа
.\scripts\backup-personal.ps1
```

### Ручной запуск бэкапа (Linux/Mac)

```bash
# Установите переменные окружения
export BACKUP_CLOUD_ENABLED=true
export BACKUP_CLOUD_PROVIDER=yandex
export BACKUP_YANDEX_TOKEN=your_oauth_token_here

# Запустите скрипт бэкапа
./scripts/backup-personal.sh
```

## Автоматизация через cron

### Linux/Mac

Добавьте в crontab (`crontab -e`):

```bash
# Ежедневный бэкап в 2:00 ночи
0 2 * * * cd /path/to/knowledge-graph && BACKUP_CLOUD_ENABLED=true BACKUP_CLOUD_PROVIDER=yandex BACKUP_YANDEX_TOKEN=your_token ./scripts/backup-personal.sh
```

### Windows (Task Scheduler)

1. Откройте Task Scheduler
2. Create Task → Triggers → Daily at 2:00 AM
3. Action: Start a program
   - Program: `powershell.exe`
   - Arguments: `-ExecutionPolicy Bypass -File "D:\knowledge-graph\scripts\backup-personal.ps1"`
   - Add environment variables:
     - `BACKUP_CLOUD_ENABLED=true`
     - `BACKUP_CLOUD_PROVIDER=yandex`
     - `BACKUP_YANDEX_TOKEN=your_token`

## Проверка бэкапов

### Просмотр бэкапов на Яндекс.Диске

1. Откройте [Яндекс.Диск](https://disk.yandex.ru)
2. Перейдите в папку `knowledge-graph-backups`
3. Бэкапы будут называться: `backup-personal-YYYY-MM-DD.sql.gz`

### Восстановление из бэкапа

```bash
# Скачайте бэкап с Яндекс.Диска (через веб-интерфейс или WebDAV)
# Распакуйте
gunzip backup-personal-2024-05-17.sql.gz

# Восстановите в PostgreSQL
psql -h localhost -p 5433 -U personal -d knowledge_personal < backup-personal-2024-05-17.sql
```

## Устранение проблем

### Ошибка "BACKUP_YANDEX_TOKEN not set"

Убедитесь, что переменная окружения `BACKUP_YANDEX_TOKEN` установлена:
```bash
echo $BACKUP_YANDEX_TOKEN  # Linux/Mac
echo $env:BACKUP_YANDEX_TOKEN  # PowerShell
```

### Ошибка "Yandex.Disk upload failed"

- Проверьте, что токен правильный и не истек
- Убедитесь, что у вас есть достаточно места на Яндекс.Диске
- Проверьте подключение к интернету

### Токен истек

OAuth токены Яндекс.Диска могут истекать. Получите новый токен по инструкции выше.

### Ошибка 401 Unauthorized

- Токен неверный или истек
- Проверьте, что приложение имеет доступ к Яндекс.Диску

## Безопасность

- **Никогда не коммитьте** `.env` файл с реальными токенами
- Храните токен в безопасном месте (менеджер паролей)
- Регулярно обновляйте токен
- Ограничьте права приложения только к Яндекс.Диску
- Используйте отдельные токены для разных окружений

## Дополнительные возможности

### Изменение папки для бэкапов

Папка по умолчанию для бэкапов: `/KnowledgeGraphBackups`

Чтобы изменить, отредактируйте конфигурацию в `knowledge-graph.config.json` или используйте переменную окружения `BACKUP_YANDEX_FOLDER`.

### Изменение периода хранения

Измените `BACKUP_RETENTION_DAYS` в конфигурации (по умолчанию 7 дней).

### Шифрование бэкапов

Для дополнительной безопасности можно зашифровать бэкапы перед загрузкой:

```bash
# Шифрование перед загрузкой
gpg --symmetric --cipher-algo AES256 backup-personal-2024-05-17.sql.gz

# Расшифровка после скачивания
gpg --decrypt backup-personal-2024-05-17.sql.gz.gpg > backup-personal-2024-05-17.sql.gz
```

## ?? ������� ������ ������� �������

### �������� �� 2026-06-15

#### ��������� ����� ?
- **������**: ��������
- **������������**: ./backups/backup-personal-YYYY-MM-DD.sql.gz
- **�������������**: ������ 24 ���� (������������� ����� Docker)
- **�������**: ������������� ������� ������ ������ 7 ����
- **������������ ������**:
  - ackup-personal-2026-06-01.sql
  - ackup-personal-2026-06-02.sql
  - ackup-personal-2026-06-07.sql
  - ackup-personal-2026-06-08.sql.gz
  - ackup-personal-2026-06-15.sql.gz (����������)

#### ������.���� ����� ?
- **������**: ������� ���������
- **��������**: BACKUP_YANDEX_TOKEN �� �����
- **������������ ����������**:
  `ash
  BACKUP_CLOUD_ENABLED=true              # ? �������
  BACKUP_YANDEX_TOKEN=                   # ? ������
  BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups  # ? ����� ������
  `
- **���� ����������**:
  `
  Backup completed successfully: /backups/backup-personal-2026-06-15.sql.gz
  Uploading to Yandex.Disk...
  Warning: BACKUP_YANDEX_TOKEN not set
  Cleaning up old backups (keeping last 7 days)...
  Backup process completed.
  `

#### Docker ������ backup_scheduler ?
- **������**: ��������
- **���������**: kg-backup-scheduler
- **�������������**: ������ 24 ����
- **������**: /scripts/backup-personal.sh
- **������ ��������**: ��������� (CRLF > LF) ��� ���������� ������ � Linux �����������

### ����������� �������� ��� ��������� ������.���� ������

1. **�������� OAuth �����**:
   `
   https://oauth.yandex.ru/authorize?response_type=token&client_id=c0ebe342af7d48fbbbfcf2d2eedb8f9e
   `

2. **���������� �����** (���� �� ��������):
   `powershell
   # PowerShell:
    =  ���_�����
   docker-compose -f docker-compose.personal.yml restart backup_scheduler
   `

   ��� � .env �����:
   `ash
   BACKUP_YANDEX_TOKEN=���_�����
   `

3. **������������� ���������**:
   `ash
   docker-compose -f docker-compose.personal.yml restart backup_scheduler
   `

### ����������

**��������� ���� ������:**
`ash
docker logs kg-backup-scheduler --tail 20
`

**��������� ���������� ��������� � ����������:**
`ash
docker inspect kg-backup-scheduler --format='{{range .Config.Env}}{{println .}}{{end}}' | findstr BACKUP
`

**��������� ������������ ��������� ������:**
`ash
dir backups
`
