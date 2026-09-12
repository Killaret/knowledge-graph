# SEC-2. Ротация Yandex backup token

## Постановка

При развёртывании Personal-стека 2026-09-09 Devin прочитал `.env` для проверки конфигурации. Значение `BACKUP_YANDEX_TOKEN` / `BACKUP_YANDEX_OAUTH_TOKEN` попало в лог инструмента. Считать токен скомпрометированным и отозвать.

## Кто выполняет

**Человек** — отзыв и перевыпуск токена возможны только владельцем аккаунта Yandex.

## Порядок действий

1. Отозвать старый токен в настройках Yandex:
   - https://id.yandex.ru/ → «Безопасность» → «Приложения и пароли» → найти приложение/скрипт резервного копирования и нажать «Отозвать доступ».
   - Или через страницу тестирования Disk API: https://yandex.ru/dev/disk/poligon/ — там есть ссылка на отзыв.
2. Выпустить новый OAuth-токен для Yandex Disk:
   - https://yandex.ru/dev/disk/poligon/ → «Получить OAuth-токен».
3. Обновить `.env` в корне репозитория:
   - `BACKUP_YANDEX_TOKEN=<new_token>`
   - `BACKUP_YANDEX_OAUTH_TOKEN=<new_token>`
4. Перезапустить `backup_scheduler`, чтобы новые значения подтянулись:
   ```powershell
   docker compose -f docker-compose.personal.yml restart backup_scheduler
   ```
5. Проверить загрузку:
   ```powershell
   docker compose -f docker-compose.personal.yml run --rm --no-deps backup_scheduler /bin/sh /scripts/backup-personal.sh daily
   ```
   В консоли должно появиться `Successfully uploaded to Yandex.Disk` (или warning, если что-то не так).

## Проверки

- Старый токен больше не принимается Yandex.
- Новый токен проходит тестовую загрузку.
- `.env` остаётся в `.gitignore` и не попадает в коммиты.

## Связанные файлы

- `.env` — секреты, не коммитится.
- `docker-compose.personal.yml` — сервис `backup_scheduler`.
- `scripts/devops/backup-personal.sh` — скрипт, который использует токен.
