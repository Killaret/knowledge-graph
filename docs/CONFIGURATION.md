# Конфигурация Knowledge Graph

Все параметры системы задаются через переменные окружения. Для локальной разработки их можно определить в файле `.env` в корне проекта (рядом с `docker-compose.yml`). В продакшене переменные передаются через `environment` в `docker-compose.yml` или через `ConfigMap`/`Secrets` в Kubernetes.

## Обязательные переменные

| Переменная | Описание | Пример |
|------------|----------|--------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable` |

## Опциональные переменные

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `SERVER_PORT` | Порт HTTP-сервера | `8080` |
| `REDIS_URL` | Адрес Redis (очередь asynq + кэш) | `localhost:6379` |
| `NLP_SERVICE_URL` | Адрес Python-сервиса для эмбеддингов и ключевых слов | `http://localhost:5000` |

## Параметры рекомендаций (граф + эмбеддинги)

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `RECOMMENDATION_ALPHA` | Вес явных связей в итоговой оценке | `0.7` |
| `RECOMMENDATION_BETA` | Вес семантического сходства (эмбеддинги) | `0.3` |
| `RECOMMENDATION_DEPTH` | Максимальная глубина BFS-обхода графа | `3` |
| `RECOMMENDATION_DECAY` | Коэффициент затухания для косвенных связей (применяется начиная со второго уровня) | `0.5` |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | Время жизни кэша рекомендаций в Redis (секунды) | `300` |
| `EMBEDDING_SIMILARITY_LIMIT` | Количество кандидатов при поиске похожих заметок через pgvector | `30` |

## Параметры визуализации графа

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `GRAPH_LOAD_DEPTH` | Максимальная глубина загрузки графа для визуализации | `2` |

## Полный пример `.env` файла

```env
# Обязательные
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable

# Опциональные
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000

# Рекомендации
RECOMMENDATION_ALPHA=0.7
RECOMMENDATION_BETA=0.3
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_CACHE_TTL_SECONDS=300
EMBEDDING_SIMILARITY_LIMIT=30

# Визуализация графа
GRAPH_LOAD_DEPTH=2

## Примеры различных конфигураций

### 1. Только явные связи (без эмбеддингов)
```env
RECOMMENDATION_ALPHA=1.0
RECOMMENDATION_BETA=0.0
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
```

### 2. Только семантическое сходство (игнорировать явные связи)
```env
RECOMMENDATION_ALPHA=0.0
RECOMMENDATION_BETA=1.0
```

### 3. Сбалансированный режим (50/50)
```env
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
```

### 4. Глубокий обход графа (осторожно, может быть медленно)
```env
RECOMMENDATION_DEPTH=5
RECOMMENDATION_DECAY=0.3
```

### 5. Увеличенный кэш рекомендаций (10 минут)
```env
RECOMMENDATION_CACHE_TTL_SECONDS=600
```

### 6. Больше кандидатов от эмбеддингов
```env
EMBEDDING_SIMILARITY_LIMIT=50
```

### 7. Высокая чувствительность к эмбеддингам (α=0.3, β=0.7)
```env
RECOMMENDATION_ALPHA=0.3
RECOMMENDATION_BETA=0.7
```

### 8. Глубокая визуализация графа (больше уровней соседей)
```env
GRAPH_LOAD_DEPTH=4
```

### 9. Минимальная визуализация графа (только прямые связи)
```env
GRAPH_LOAD_DEPTH=1
```

Как применить изменения
После редактирования .env или переменных в docker-compose.yml перезапустите бэкенд:

bash
docker-compose restart backend
Проверьте логи – должна появиться строка с загруженной конфигурацией:

bash
docker logs kg-backend --tail 30 | grep "Config loaded"
Пример вывода:

text
Config loaded: alpha=0.70, beta=0.30, depth=3, decay=0.50, cacheTTL=5m0s, embeddingLimit=30, graphLoadDepth=2

Формула итогового веса рекомендации
Для заметки-источника A и кандидата C:

text
score = α * explicit_weight(A, C) + β * content_sim(A, C)
explicit_weight(A, C) — сумма весов всех явных связей от A к C (с затуханием для косвенных путей, начиная со второго уровня).

content_sim(A, C) — косинусное сходство эмбеддингов, нормализованное в диапазон [0,1].

α + β не обязательно равно 1, но для сохранения масштаба рекомендуется сумма = 1.

Примечания
Эмбеддинги вычисляются асинхронно воркером. Убедитесь, что воркер (kg-worker) запущен и обработал задачи для заметок, иначе content_sim будет равен 0.

При beta > 0 требуется, чтобы в таблице note_embeddings были векторы для всех заметок. Если их нет, рекомендации будут полагаться только на явные связи.

Изменение RECOMMENDATION_DEPTH больше 3 может существенно увеличить нагрузку на БД и время ответа.

Кэш рекомендаций инвалидируется по TTL. Для немедленной очистки можно перезапустить бэкенд или удалить ключи вручную:
docker exec kg-redis redis-cli KEYS "suggestions:*" | xargs docker exec kg-redis redis-cli DEL