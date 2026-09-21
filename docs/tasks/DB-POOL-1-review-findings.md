# DB-POOL-1 — разбор ревью

Ревьюер: Claude Code. Дата: 2026-09-21. Реализация: Devin (09-18).

**Вердикт: отклонено** — параметризация сделана и подключена, но ничем не охраняется.

## Проверено

- `config.go`: секция `database.pool` + env `POSTGRES_MAX_OPEN_CONNS`, `POSTGRES_MAX_IDLE_CONNS`,
  `POSTGRES_CONN_MAX_LIFETIME_SECONDS`, `POSTGRES_CONN_MAX_IDLE_TIME_SECONDS`.
- `db.ConnectWithPool` вызывается из `cmd/server/main.go:387` и `cmd/worker/main.go:58` с
  `PoolConfig` из конфига; CLI-утилиты — через `db.Connect` с умолчаниями, что уместно.
- `GET /api/v1/metrics/database` анониму на стенде с `SKIP_AUTH=false` — **401**: гейт `RequireAdmin()` работает.
- Тесты `internal/infrastructure/db` и `cmd/server` зелёные.

## Блокер: значение из конфига можно игнорировать — и никто не заметит

Мутация: `sqlDB.SetMaxOpenConns(pool.MaxOpenConns)` → `sqlDB.SetMaxOpenConns(25)`:

```
ok  	knowledge-graph/internal/infrastructure/db	0.125s
```

Все тесты пакета зелёные. Они проверяют пустой и невалидный DSN, нормализацию `PoolConfig` и
статистику на невалидной базе — но ни один не открывает базу и не сверяет, что применилось
именно сконфигурированное значение. Задача называлась «параметризовать пул»; откати
параметризацию — и ничего не покраснеет.

Нужно: интеграционный тест (`-tags=integration`, testcontainers уже есть): `ConnectWithPool` с
`MaxOpenConns: 7` → `sqlDB.Stats().MaxOpenConnections == 7`; и тест конфига:
`POSTGRES_MAX_OPEN_CONNS=7` в env → `cfg.DatabasePoolMaxOpenConns == 7`. Мутация выше обязана
ронять первый.

## Принято без замечаний

- Умолчания не изменились — поведение стендов прежнее.
- Метрики за админом, а не за обычной авторизацией.
