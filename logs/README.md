# Логирование в проекте Knowledge Graph

## Структура директорий

```
logs/
├── frontend/        # Логи фронтенда (Svelte/Vite)
├── backend/         # Логи бэкенда (Go)
├── nlp-service/     # Логи NLP-сервиса (Python)
└── README.md        # Этот файл
```

## Настройка логирования

### Frontend

Используйте утилиту `logger` из `$lib/utils/logger`:

```typescript
import { createLogger } from '$lib/utils/logger';

const logger = createLogger('MyComponent');

logger.debug('Debug message', { data: 'value' });
logger.info('Info message');
logger.warn('Warning message');
logger.error('Error message', error);
```

Логи выводятся в консоль с префиксом `[timestamp] [LEVEL] [context]`.

### Backend

```go
import "knowledge-graph/backend/internal/infrastructure/logger"

// Инициализация при старте приложения
logger.Initialize(logger.Config{
    Level:      logger.INFO,
    JSONFormat: true,
    LogFile:    "logs/backend/app.log",
})

// Использование
log := logger.WithContext("handler")
log.Info("Request processed", map[string]interface{}{"id": id})
log.Error("Failed to process", err)
```

## Правила

1. **Не коммитьте файлы логов** - они добавлены в `.gitignore`
2. **Структура логов** должна быть в git (`.gitkeep` файлы)
3. **Уровни логирования**:
   - `DEBUG` - детальная информация для разработки
   - `INFO` - общая информация о работе
   - `WARN` - предупреждения, не критичные
   - `ERROR` - ошибки, требующие внимания

## Очистка логов

Логи старше 30 дней автоматически удаляются при ротации (необходимо настроить).
