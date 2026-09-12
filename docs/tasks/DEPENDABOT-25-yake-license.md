# DEPENDABOT-25: `yake` 0.4.8 → 0.7.3 — блокер по лицензии

## PR

- <https://github.com/Killaret/knowledge-graph/pull/25>
- Ветка: `dependabot/pip/nlp-service/yake-0.7.3`
- Файл: `nlp-service/requirements.txt`
- Изменение: `yake==0.4.8` → `yake==0.7.3`

## Что такое `yake`

`yake` — unsupervised keyword extraction. В `nlp-service` используется для извлечения ключевых слов из текста заметок.

## Почему заблокирован

`Dependency Review` падает с лицензионной ошибкой:

```text
yake 0.7.3 — AGPL-3.0-only AND AGPL-3.0-or-later AND LGPL-3.0-or-later
```

Текущий `allow-licenses` в `.github/workflows/security.yml`:

```yaml
allow-licenses: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, LicenseRef-scancode-google-patent-license-golang
```

AGPL/LGPL не входят в разрешённый список.

## CI-статус

| Чек | Результат |
|---|---|
| `Analyze (actions/go/javascript-typescript/python)` | ✅ |
| `CodeQL` | ✅ |
| `Frontend Security Audit` | ✅ |
| `Dependency Review` | ❌ license |

## Варианты решения

1. **Закрыть PR, оставить `yake 0.4.8`.**
   - Самый безопасный путь. Текущая версия работает, риска лицензии нет.
2. **Разрешить AGPL/LGPL в `allow-licenses`.**
   - Нужно понимать правовые последствия: `yake` — copyleft, распространяется в том же окружении, что и NLP-сервис.
   - Для `nlp-service` (отдельный Python-сервис, общается по HTTP) это может быть приемлемо, но требует решения владельца/юриста.
3. **Заменить `yake` на альтернативу.**
   - Возможные замены: `keybert`, `rake-nltk`, `TextRank` (own implementation), `scikit-learn` TF-IDF.
   - Требует проверки качества ключевых слов и пересчёта рекомендаций.

## Следующее действие

Решение владельца: закрыть, разрешить лицензию или заменить библиотеку.

## Ссылки

- PR #25: <https://github.com/Killaret/knowledge-graph/pull/25>
- Заблокированный run Dependency Review: <https://github.com/Killaret/knowledge-graph/actions/runs/34708560386/job/103593243789>
- `allow-licenses` в `security.yml`: <ref_file file="D:\knowledge-graph\.github\workflows\security.yml" />
