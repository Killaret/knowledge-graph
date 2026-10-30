# BACKUP-1. Настроить облачный бэкап Personal-стека

## Состояние

2026-09-09: облачная загрузка на Yandex.Disk отключена (`BACKUP_CLOUD_ENABLED=false`), локальный бэкап работает. Причина — текущий `BACKUP_YANDEX_TOKEN` не прошёл `401 Unauthorized`, и владелец решил вручную переносить бэкапы, пока облако не будет настроено отдельно.

## Что нужно сделать

1. Получить рабочий OAuth-токен Yandex.Disk:
   - https://yandex.ru/dev/disk/poligon/
   - убедиться, что у токена есть права на запись в Диск.
2. Обновить `.env`:
   - `BACKUP_YANDEX_TOKEN=<new_token>`
   - `BACKUP_YANDEX_OAUTH_TOKEN=<new_token>`
3. Включить облако:
   - `BACKUP_CLOUD_ENABLED=true`
4. Пересоздать `backup_scheduler`, чтобы подхватилось новое окружение:
   ```powershell
   docker compose -f docker-compose.personal.yml up -d --force-recreate backup_scheduler
   ```
5. Прогнать ручной бэкап и убедиться, что `Successfully uploaded to Yandex.Disk` появляется без `401`.
6. Проверить, что папка `/KnowledgeGraphBackups` (или настроенная `BACKUP_YANDEX_FOLDER`) создалась на Диске.

## Заметки

- Локальный бэкап (в `./backups`) остаётся и работает в любом случае.
- `YANDEX_CLIENT_ID` и `YANDEX_CLIENT_SECRET` не участвуют в загрузке бэкапов — они для OAuth-входа пользователей.
- `backup-personal.ps1` на хосте требует локальный `pg_dump`/`gzip`; для проверки удобнее `docker exec kg-backup-scheduler /bin/sh /scripts/backup-personal.sh daily`.
