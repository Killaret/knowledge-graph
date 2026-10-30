# Logging in Knowledge Graph Project

## Directory Structure

```
logs/
├── frontend/        # Frontend logs (Svelte/Vite)
├── backend/         # Backend logs (Go)
├── nlp-service/     # NLP service logs (Python)
└── README.md        # This file
```

## Logging Setup

### Frontend

Use the `logger` utility from `$lib/utils/logger`:

```typescript
import { createLogger } from '$lib/utils/logger';

const logger = createLogger('MyComponent');

logger.debug('Debug message', { data: 'value' });
logger.info('Info message');
logger.warn('Warning message');
logger.error('Error message', error);
```

Logs are output to console with prefix `[timestamp] [LEVEL] [context]`.

### Backend

```go
import "knowledge-graph/backend/internal/infrastructure/logger"

// Initialize at application startup
logger.Initialize(logger.Config{
    Level:      logger.INFO,
    JSONFormat: true,
    LogFile:    "logs/backend/app.log",
})

// Usage
log := logger.WithContext("handler")
log.Info("Request processed", map[string]interface{}{"id": id})
log.Error("Failed to process", err)
```

## Rules

1. **Do not commit log files** - they are added to `.gitignore`
2. **Log structure** should be in git (`.gitkeep` files)
3. **Logging levels**:
   - `DEBUG` - detailed information for development
   - `INFO` - general operational information
   - `WARN` - non-critical warnings
   - `ERROR` - errors requiring attention

## Log Cleanup

Logs older than 30 days are automatically deleted during rotation (needs to be configured).
