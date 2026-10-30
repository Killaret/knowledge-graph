# Инвентарь лицензий зависимостей

Проект — MIT (`LICENSE` в корне), образы публикуются на Docker Hub, поэтому
копилефтные зависимости недопустимы: распространение образа является триггером.
Инцидент: `yake` прожил в `nlp-service` с первого коммита с GPL-3.0-or-later
при метаданных «LGPLv3» — см. `docs/tasks/NLP-2-yake-replace-keybert-lemmatization.md`
и `DEPENDABOT-25-yake-license.md`.

**Проверка:** `node scripts/testing/check-licenses.mjs` — фаза `check-all` и CI
(`licenses` в `core-checks.tsv`). Краснеет, если:

- прямая зависимость отсутствует в таблицах ниже;
- в её лицензии встречается `GPL`, `LGPL`, `AGPL`, `SSPL`, `EUPL`.

Скрипт проверяет только прямые зависимости и полноту инвентаря — это процессный
барьер, а не аудит транзитивных пакетов. При добавлении зависимости сначала
проверить лицензию по файлу `LICENSE` внутри пакета (метаданные могут врать),
затем занести строку сюда.

## nlp-service (Python, `nlp-service/requirements.txt`)

| Пакет | Экосистема | Версия | Лицензия | Источник |
|---|---|---|---|---|
| fastapi | python | 0.115.5 | MIT | PyPI metadata |
| uvicorn | python | 0.34.0 | BSD-3-Clause | PyPI metadata |
| sentence-transformers | python | 2.7.0 | Apache-2.0 | PyPI metadata |
| pydantic | python | 2.13.5 | MIT | PyPI metadata |
| python-dotenv | python | 1.2.3 | BSD-3-Clause | PyPI metadata |
| keybert | python | 0.9.0 | MIT | PyPI metadata |
| pymorphy3 | python | 2.0.6 | MIT | PyPI metadata |
| pymorphy3-dicts-ru | python | 2.4.417150.4580142 | MIT | PyPI metadata; словарные данные — OpenCorpora (см. ниже) |
| dawg2-python | python | 0.9.0 | MIT | PyPI metadata |
| nltk | python | 3.10.3 | Apache-2.0 | PyPI metadata |
| huggingface-hub | python | 0.23.0 | Apache-2.0 | PyPI metadata |
| httpx | python | 0.28.1 | BSD-3-Clause | PyPI metadata |

### Данные (не код)

| Данные | Лицензия | Примечание |
|---|---|---|
| Словари `pymorphy3-dicts-ru` (OpenCorpora) | CC BY-SA 3.0 | Данные словарей, а не код: share-alike касается самих данных, пакет — MIT |
| NLTK `wordnet`, `punkt`, `stopwords` | WordNet: Princeton licence (BSD-like); punkt/stopwords — Apache-2.0 | Скачиваются при сборке образа |
| Модель `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2` | Apache-2.0 | HuggingFace model card |

## backend (Go, `backend/go.mod`, прямые зависимости)

| Пакет | Экосистема | Версия | Лицензия | Источник |
|---|---|---|---|---|
| github.com/DATA-DOG/go-sqlmock | go | 1.5.2 | BSD-3-Clause | LICENSE в репозитории |
| github.com/alicebob/miniredis/v2 | go | 2.39.0 | MIT | LICENSE в репозитории |
| github.com/cenkalti/backoff/v4 | go | 4.3.0 | MIT | LICENSE в репозитории |
| github.com/gin-contrib/gzip | go | 1.1.0 | MIT | LICENSE в репозитории |
| github.com/gin-gonic/gin | go | 1.12.0 | MIT | LICENSE в репозитории |
| github.com/golang-jwt/jwt/v5 | go | 5.3.1 | MIT | LICENSE в репозитории |
| github.com/google/uuid | go | 1.6.0 | BSD-3-Clause | LICENSE в репозитории |
| github.com/hibiken/asynq | go | 0.26.0 | MIT | LICENSE в репозитории |
| github.com/jackc/pgx/v5 | go | 5.10.0 | MIT | LICENSE в репозитории |
| github.com/joho/godotenv | go | 1.5.1 | MIT | LICENSE в репозитории |
| github.com/lib/pq | go | 1.12.3 | MIT | LICENSE в репозитории |
| github.com/pgvector/pgvector-go | go | 0.4.1 | MIT | LICENSE в репозитории |
| github.com/redis/go-redis/v9 | go | 9.22.0 | BSD-2-Clause | LICENSE в репозитории |
| github.com/stretchr/testify | go | 1.12.1 | MIT | LICENSE в репозитории |
| github.com/swaggo/files | go | 1.0.1 | MIT | LICENSE в репозитории |
| github.com/swaggo/gin-swagger | go | 1.6.1 | MIT | LICENSE в репозитории |
| github.com/testcontainers/testcontainers-go | go | 0.44.0 | MIT | LICENSE в репозитории |
| github.com/testcontainers/testcontainers-go/modules/postgres | go | 0.44.0 | MIT | LICENSE в репозитории |
| go.mongodb.org/mongo-driver | go | 1.17.9 | Apache-2.0 | LICENSE в репозитории |
| golang.org/x/crypto | go | 0.55.0 | BSD-3-Clause | LICENSE в репозитории |
| golang.org/x/net | go | 0.58.0 | BSD-3-Clause | LICENSE в репозитории |
| golang.org/x/text | go | 0.41.0 | BSD-3-Clause | LICENSE в репозитории |
| gopkg.in/yaml.v3 | go | 3.0.1 | MIT/Apache-2.0 | LICENSE в репозитории |
| gorm.io/datatypes | go | 1.2.7 | MIT | LICENSE в репозитории |
| gorm.io/driver/postgres | go | 1.6.3 | MIT | LICENSE в репозитории |
| gorm.io/gorm | go | 1.31.2 | MIT | LICENSE в репозитории |

## graph-service (Go, `services/graph-service/go.mod`, прямые зависимости)

| Пакет | Экосистема | Версия | Лицензия | Источник |
|---|---|---|---|---|
| google.golang.org/grpc | go | 1.83.2 | Apache-2.0 | LICENSE в репозитории |
| google.golang.org/protobuf | go | 1.36.12 | BSD-3-Clause | LICENSE в репозитории |

(остальные прямые зависимости graph-service — те же пакеты, что у backend:
miniredis, golang-jwt, uuid, pgx, go-redis, testify, testcontainers — см.
таблицу backend выше.)

## frontend (npm, `frontend/package.json`)

| Пакет | Экосистема | Лицензия | Источник |
|---|---|---|---|
| @argos-ci/playwright | npm | MIT | package.json/license |
| @cucumber/cucumber | npm | MIT | package.json/license |
| @eslint/js | npm | MIT | package.json/license |
| @playwright/test | npm | Apache-2.0 | package.json/license |
| @sveltejs/adapter-node | npm | MIT | package.json/license |
| @sveltejs/kit | npm | MIT | package.json/license |
| @sveltejs/vite-plugin-svelte | npm | MIT | package.json/license |
| @testing-library/jest-dom | npm | MIT | package.json/license |
| @testing-library/svelte | npm | MIT | package.json/license |
| @testing-library/user-event | npm | MIT | package.json/license |
| @types/d3-force | npm | MIT | package.json/license |
| @types/node | npm | MIT | package.json/license |
| @types/three | npm | MIT | package.json/license |
| @vitest/coverage-v8 | npm | MIT | package.json/license |
| cross-env | npm | MIT | package.json/license |
| d3-force | npm | ISC | package.json/license |
| d3-force-3d | npm | MIT | package.json/license |
| eslint | npm | MIT | package.json/license |
| eslint-plugin-jsx-a11y | npm | MIT | package.json/license |
| globals | npm | MIT | package.json/license |
| happy-dom | npm | ISC | package.json/license |
| jsdom | npm | MIT | package.json/license |
| ky | npm | MIT | package.json/license |
| madge | npm | MIT | package.json/license |
| msw | npm | MIT | package.json/license |
| prettier | npm | MIT | package.json/license |
| prettier-plugin-svelte | npm | MIT | package.json/license |
| svelte | npm | MIT | package.json/license |
| svelte-check | npm | MIT | package.json/license |
| svelte-eslint-parser | npm | MIT | package.json/license |
| three | npm | MIT | package.json/license |
| tippy.js | npm | MIT | package.json/license |
| tsx | npm | MIT | package.json/license |
| typescript | npm | Apache-2.0 | package.json/license |
| typescript-eslint | npm | MIT | package.json/license |
| vite | npm | MIT | package.json/license |
| vite-tsconfig-paths | npm | MIT | package.json/license |
| vitest | npm | MIT | package.json/license |
| web-streams-polyfill | npm | MIT | package.json/license |
