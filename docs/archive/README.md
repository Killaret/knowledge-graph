# 📦 Архив документации

**Статус:** Устаревшие файлы, сохранены для истории

**Дата архивации:** 2026-04-15

---

## Что здесь находится

Эта папка содержит исторические версии документации, которые больше не актуальны, но могут быть полезны для понимания эволюции проекта.

---

## Архивированные файлы

### Отчёты тестирования (апрель 2026)

| Файл | Дата | Почему архивирован |
|------|------|-------------------|
| `API_VERIFICATION_REPORT_2026-04-13.md` | 13 апреля | Проблемы из отчёта исправлены в коде |
| `FINAL_TEST_REPORT_2026-04-13.md` | 13 апреля | Устарел (9/12 → 48 тестов) |
| `TEST_FIXES_REPORT_2026-04-13.md` | 13 апреля | Исправления уже в main branch |
| `IMPLEMENTATION_SUMMARY_2026-04-13.md` | 13 апреля | Дублирует CHANGELOG.md |

### Файлы из корня репозитория (июль 2026)

| Файл | Дата | Почему архивирован |
|------|------|-------------------|
| `FINAL_TEST_REPORT.md` | июль | Устаревший финальный отчёт; заменён актуальным прогоном |
| `REGRESSION_TEST_REPORT.md` | июль | Устаревший отчёт регрессии |
| `FRONTEND_REFACTORING_LOG.md` | июль | Исторический лог рефакторинга |
| `MANUAL_TEST_REPORT.md` | июль | Устаревший ручной отчёт |
| `MANUAL_TESTING_CHECKLIST_COMPLETE.md` | июль | Заменён `../MANUAL_TEST_CHECKLISTS.md` |
| `MANUAL_TESTING_CHECKLIST_COMPLETE_EN.md` | июль | Заменён `../MANUAL_TEST_CHECKLISTS.md` |
| `PERSONAL_STACK_DEEP_VERIFICATION.md` | июль | Устаревшая верификация personal stack |
| `PERSONAL_STACK_VERIFICATION_REPORT.md` | июль | Устаревший отчёт personal stack |
| `TEST_INSPECTION_REPORT.md` | июль | Устаревший инспекционный отчёт |

### Промежуточные отчёты из docs/ (июль 2026)

| Файл | Дата | Почему архивирован |
|------|------|-------------------|
| `DOCUMENTATION_AUDIT_REPORT.md` | июль | Аудит документации выполнен |
| `PRIORITY_FIXES_REPORT.md` | июль | Исправления внесены |
| `ROUTE_AUTH_REPORT.md` | июль | Проблемы auth исправлены |
| `TEST_EXECUTION_REPORT.md` | июль | Данные перенесены в финальный отчёт |

**Актуальная документация по тестированию:**
- `../TESTING.md` — Полное руководство по тестированию
- `../MANUAL_TEST_CHECKLISTS.md` — Чек-лист ручного тестирования (EN)
- `../MANUAL_TEST_CHECKLISTS_RU.md` — Чек-лист ручного тестирования (RU)
- `../REGRESSION_TEST_PLAN.md` — План регрессионного тестирования
- `../CHANGELOG.md` — История изменений с результатами тестов

---

## Где искать актуальную документацию

```
📁 docs/
├── DEPLOYMENT.md          ← Развёртывание (Docker, K8s)
├── TESTING.md             ← Тестирование (Go, Playwright)
├── CONFIGURATION.md       ← Env переменные
├── CHANGELOG.md           ← История v1.0.0
├── API_ERRORS.md          ← Ошибки API
├── FRONTEND_ARCHITECTURE.md ← Three.js, Progressive Rendering
├── architecture/          ← C4 Model, ADR, UML
├── COPILOT_DOCS/         ← AI-ассистенты
└── archive/             ← 📦 Вы здесь
```

---

## Контекст

Версия 1.0.0 выпущена 2026-04-15:
- ✅ 3D Progressive Rendering (Fog of War)
- ✅ Полный стек тестирования (48 E2E тестов)
- ✅ Production-ready документация
- ✅ Backend: 25+ unit тестов

Отчёты из этого архива относятся к промежуточным этапам разработки MVP.

---

**Не используйте файлы из этой папки для текущей работы!**

Используйте актуальную документацию из корня `docs/`.
