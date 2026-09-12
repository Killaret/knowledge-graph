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

## Сравнение версий: `yake 0.4.8` (текущая) vs `0.7.3` (предлагаемая)

| Аспект | `0.4.8` | `0.7.3` | Влияние на проект |
|---|---|---|---|
| Дата релиза | 26 апр 2021 | 9 фев 2026 | — |
| Требуемый Python | py2/py3 (фактически ≥3.6) | ≥3.10 | Проект на `python:3.11-slim`, подходит |
| Базовый API | `yake.KeywordExtractor(lan, n, dedupLim, dedupFunc, windowsSize, top)` | Тот же набор плюс `lemmatize`, `lemmatizer`, `lemma_aggregation` | Обратная совместимость сохраняется |
| Лемматизация | Нет | Да (с 0.6.0+) | Потенциально снижает дублирование: «дерево/деревья», «run/running» |
| Производительность | Базовая | «Performance Optimizations» (0.7.3 release notes) | Может быть быстрее, конкретных замеров нет |
| Зависимости | `tabulate`, `click>=6.0`, `numpy`, `segtok`, `networkx`, `jellyfish` | `tabulate`, `click>=6.0`, `numpy>=1.24.0`, `segtok`, `networkx`, `jellyfish` | `numpy` подняли до 1.24.0 |
| Лицензия PyPI | `LGPLv3` | `LGPLv3` | Метаданные PyPI выглядят одинаково |
| Лицензия `LICENSE` файла (sdist/GitHub) | GPL-3.0-only (по Licensie scan) | AGPL-3.0-or-later | Репозиторий INESCTEC перешёл на AGPL с 0.6.0+ |

## Почему Dependency Review видит `AGPL-3.0-only AND AGPL-3.0-or-later AND LGPL-3.0-or-later`

GitHub `dependency-review-action` собирает лицензии из двух источников:

1. **PyPI-метаданные** (`pyproject.toml`): поле `license = {text = "LGPLv3"}` + classifier `License :: OSI Approved :: GNU General Public License v3 (GPLv3)`.
2. **Файл `LICENSE`** в sdist/GitHub: текст GNU Affero General Public License v3 or later.

Поэтому сканер видит одновременно LGPL (метаданные), GPL (classifier) и AGPL (LICENSE-файл) и выдаёт комбинированную строку. CI падает, потому что в `allow-licenses` нет ни AGPL, ни LGPL.

## Что значит «риск лицензионной политики»

`yake 0.7.3` распространяется, судя по `LICENSE`, под **AGPL-3.0-or-later**. Это сильный copyleft:

- Если сервис использует AGPL-код, то при предоставлении доступа к сервису через сеть пользователям у вас может возникнуть обязанность раскрывать исходный код **всего сервиса** (эффект Affero).
- `nlp-service` — отдельный Python/FastAPI микросервис, но он является частью всего Knowledge Graph. Решение вопроса «является ли использование `yake` внутри `nlp-service` производным произведением» — юридическое, не инженерное.
- Проект сейчас, по-видимому, **не open-source под AGPL**. Поэтому обновление `yake` до AGPL-версии без юридической проверки создаёт риск для интеллектуальной собственности.
- Альтернатива: в `yake` README и `LICENSE` указано, что для промышленных проектов, не желающих использовать AGPL, доступна **коммерческая лицензия**.

## Использование в нашем коде

Файл: <ref_file file="D:\knowledge-graph\nlp-service\app\nlp_utils.py" />

```python
kw_extractor = yake.KeywordExtractor(lan="ru", top=20, stopwords=stop_words)
...
keywords = kw_extractor.extract_keywords(text)
```

Используется только базовый API. Новая функциональность (лемматизация) по умолчанию не включена. То есть **функциональная выгода от обновления сейчас минимальна**.

## Варианты решения

1. **Закрыть PR, оставить `yake 0.4.8`.**
   - Самый безопасный путь. Текущая версия работает, риска лицензии нет.
2. **Разрешить AGPL/LGPL в `allow-licenses`.**
   - Нужно понимать правовые последствия: `yake` — copyleft, распространяется в том же окружении, что и NLP-сервис.
   - Для `nlp-service` (отдельный Python-сервис, общается по HTTP) это может быть приемлемо, но требует решения владельца/юриста.
3. **Заменить `yake` на альтернативу.**
   - Возможные замены: `keybert`, `rake-nltk`, `TextRank` (own implementation), `scikit-learn` TF-IDF.
   - Требует проверки качества ключевых слов и пересчёта рекомендаций.

## Рекомендация

**Функциональная выгода от `0.7.3` для нашего кода минимальна:** мы не используем лемматизацию, базовый `KeywordExtractor.extract_keywords()` работает одинаково в обеих версиях.

**Юридический риск реален:** `yake 0.7.3` фактически поставляется под AGPL-3.0-or-later. Если Knowledge Graph не планирует стать open-source под AGPL, обновление без юридической проверки опасно. В репозитории `yake` README прямо говорит: *«A commercial license is also available for use in industrial projects and collaborations that do not wish to use the AGPL 3 license»*.

**Предпочтительный путь:**

1. **Оставить `yake 0.4.8`** и **закрыть PR #25**. Это безопасно и не требует лицензионной чистки.
2. Если лемматизация действительно нужна — **купить коммерческую лицензию** у INESC TEC или **перейти на альтернативу** (`keybert`, `rake-nltk`, собственный TF-IDF/TextRank).
3. Добавлять `AGPL-3.0-only`, `AGPL-3.0-or-later`, `LGPL-3.0-or-later` в `allow-licenses` рекомендуется **только после консультации с юристом** и при явном согласии владельца на AGPL-риск.

## Решение владельца (2026-09-12)

- PR #25 **не мержится**.
- `yake` **остаётся `0.4.8`** до реализации замены.
- Выбран путь: **заменить `yake` на `keybert` (MIT) с обязательной лемматизацией**.
- Детальный план и открытые вопросы: [`YAKE-REPLACE-KEYBERT-LEMMATIZATION.md`](YAKE-REPLACE-KEYBERT-LEMMATIZATION.md).

## Следующее действие

Claude Code готовит полную постановку по замене `yake` на `keybert` + лемматизация (рус/англ). Реализация — только после принятой постановки.

## Ссылки

- PR #25: <https://github.com/Killaret/knowledge-graph/pull/25>
- Заблокированный run Dependency Review: <https://github.com/Killaret/knowledge-graph/actions/runs/34708560386/job/103593243789>
- `allow-licenses` в `security.yml`: <ref_file file="D:\knowledge-graph\.github\workflows\security.yml" />
