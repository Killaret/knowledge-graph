# Source Text Handler

Java микросервис для обработки и импорта документов в систему Knowledge Graph.

## Назначение

Сервис принимает задачи на импорт документов из очереди Redis, обрабатывает их (парсит, разбивает на чанки), создает заметки в основном backend через HTTP API и публикует результаты обратно в очередь.

## Архитектура

### Технологический стек
- **Java 17** с Maven
- **Apache Tika 2.9.2** - универсальный парсер документов (PDF, DOCX, TXT и др.)
- **Redis (Lettuce)** - очереди asynq и хранилище состояния
- **Resilience4j** - circuit breaker и retry для HTTP запросов
- **Jackson** - JSON сериализация

### Архитектурные паттерны
- **Clean Architecture** - разделение на domain, application, infrastructure слои
- **Hexagonal Architecture** - порты (InboundQueuePort, OutboundQueuePort, NoteCreatorPort)
- **Strategy Pattern** - ChunkingStrategy, ParserFactory
- **Circuit Breaker** - защита от сбоев backend API

## Основные компоненты

### Domain слой
- `DocumentParser` - интерфейс парсера документов
- `ChunkingStrategy` - стратегия разбиения текста на чанки
- `LinkDetector` - обнаружение связей между чанками
- `ImportTask` - задача на импорт (FILE, URL, TEXT)
- `DocumentChunk` - фрагмент текста с индексом
- `ImportResult` - результат обработки

### Application слой
- `ImportDocumentHandler` - оркестратор всего процесса импорта
- Порты: `InboundQueuePort`, `OutboundQueuePort`, `NoteCreatorPort`

### Infrastructure слой
- `HybridChunker` - гибридный алгоритм чанкинга (параграфы + sliding window)
- `ParserFactory` - фабрика парсеров (Tika для файлов/URL, Text для plain text)
- `TikaDocumentParser` - парсер на базе Apache Tika
- `TextDocumentParser` - простой текстовый парсер
- `AsynqInboundAdapter` - адаптер входящей очереди
- `AsynqOutboundAdapter` - адаптер исходящей очереди
- `NoteCreatorHttpClient` - HTTP клиент для backend API
- `ResilientNoteCreator` - обертка с circuit breaker и retry
- `RedisImportStateRepository` - хранилище состояния в Redis

## Алгоритм работы

1. **Получение задачи** - чтение из очереди `asynq:import:document:pending`
2. **Парсинг** - в зависимости от типа задачи:
   - FILE: декодирование base64 + парсинг через Tika
   - URL: загрузка и парсинг через Tika
   - TEXT: использование как есть
3. **Чанкинг** - разбиение текста на фрагменты с помощью `HybridChunker`:
   - Короткие параграфы → отдельные чанки
   - Длинные параграфы → sliding window по предложениям
   - Перекрытие чанков для сохранения контекста
4. **Создание заметок** - для каждого чанка:
   - HTTP POST запрос к backend API
   - Retry при ошибках (3 попытки)
   - Circuit breaker при массовом отказе backend
5. **Обнаружение ссылок** - опционально, если включено в настройках
6. **Публикация результата** - в очередь `asynq:import:responses:pending`

## Конфигурация

### Переменные окружения
- `REDIS_URI` - URL для подключения к Redis (по умолчанию: `redis://localhost:6379/0`)
- `BACKEND_URL` - URL backend API для создания заметок (захардкожено: `http://backend:8080`)

### Настройки чанкинга (в ImportOptions)
- `chunkSize` - максимальный размер чанка в словах
- `overlap` - размер перекрытия между чанками в словах
- `minChunkLength` - минимальная длина чанка в символах

## Запуск

### Через Docker Compose
```bash
# Основной стек
docker-compose up -d source-text-handler

# Личный инстанс
docker-compose -f docker-compose.personal.yml up -d source-text-handler_personal
```

### Локально с Maven
```bash
# Сборка
mvn clean package

# Запуск
java -jar target/SourceTextHandler-1.0-SNAPSHOT.jar
```

## Интеграция с системой

### Входящие задачи (Redis)
Очередь: `asynq:import:document:pending`

Формат задачи (JSON):
```json
{
  "eventId": "unique-event-id",
  "correlationId": "correlation-id",
  "type": "FILE|URL|TEXT",
  "content": "base64-encoded-content|url|plain-text",
  "contentType": "application/pdf",
  "importOptions": {
    "chunkSize": 500,
    "overlap": 50,
    "minChunkLength": 100,
    "createLinks": true
  },
  "metadata": {
    "filename": "document.pdf",
    "author": "John Doe"
  }
}
```

### Исходящие результаты (Redis)
Очередь: `asynq:import:responses:pending`

Формат результата:
```json
{
  "correlationId": "correlation-id",
  "status": "COMPLETED|PARTIAL|FAILED",
  "noteIds": ["note-id-1", "note-id-2"],
  "links": [
    {
      "sourceNoteId": "note-id-1",
      "targetNoteId": "note-id-2",
      "strength": 0.8
    }
  ],
  "errors": ["error message"],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### HTTP API Backend
Сервис вызывает следующие endpoints:

**Создание заметки:**
```
POST /api/notes
Content-Type: application/json

{
  "title": "chunk title",
  "content": "chunk content",
  "metadata": {}
}
```

**Создание связи:**
```
POST /api/links
Content-Type: application/json

{
  "source_note_id": "note-id-1",
  "target_note_id": "note-id-2",
  "strength": 0.8
}
```

## Health Check

Сервис предоставляет health check на порту 8081:
```
GET http://localhost:8081/health
```

Возвращает статус подключения к Redis.

## Тестирование

### Тестовая задача
```bash
# Просмотр тестовой задачи
cat test-task.json

# Отправка в очередь (требует redis-cli)
redis-cli -h localhost -p 6379 LPUSH "asynq:import:document:pending" "$(cat test-task.json)"
```

## Разработка

### Структура проекта
```
source-text-handler/
├── src/main/java/com/alximac/knowledgegraph/texthandler/
│   ├── application/        # Бизнес-логика и порты
│   ├── domain/            # Доменная модель и сервисы
│   ├── infrastructure/    # Реализация инфраструктуры
│   └── config/            # Конфигурация приложения
├── pom.xml                # Maven конфигурация
├── Dockerfile             # Docker образ
└── README.md              # Документация
```

### Добавление нового парсера
1. Реализовать интерфейс `DocumentParser`
2. Добавить в `ParserFactory`
3. Обновить `TaskType` если нужно

### Изменение стратегии чанкинга
1. Реализовать интерфейс `ChunkingStrategy`
2. Заменить `HybridChunker` в `AppConfig`

## Troubleshooting

### Проблемы с подключением к Redis
- Проверьте переменную `REDIS_URI`
- Убедитесь что Redis доступен: `redis-cli ping`

### Ошибки при создании заметок
- Проверьте что backend доступен по указанному URL
- Посмотрите логи для деталей ошибок HTTP запросов
- Circuit breaker может временно блокировать запросы при массовых сбоях

### Проблемы с парсингом файлов
- Убедитесь что файл поддерживается Apache Tika
- Проверьте что content корректно закодирован в base64 для FILE типа
- Для URL проверьте что ресурс доступен

## License

Сервис является частью проекта Knowledge Graph.
