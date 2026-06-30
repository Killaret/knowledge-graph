# Source Text Handler

Java microservice for processing and importing documents into the Knowledge Graph system.

## Purpose

The service accepts document import tasks from the Redis queue, processes them (parses, chunks), creates notes in the main backend via HTTP API, and publishes results back to the queue.

## Architecture

### Technology Stack
- **Java 17** with Maven
- **Apache Tika 2.9.2** - universal document parser (PDF, DOCX, TXT, etc.)
- **Redis (Lettuce)** - asynq queues and state storage
- **Resilience4j** - circuit breaker and retry for HTTP requests
- **Jackson** - JSON serialization

### Architectural Patterns
- **Clean Architecture** - separation into domain, application, infrastructure layers
- **Hexagonal Architecture** - ports (InboundQueuePort, OutboundQueuePort, NoteCreatorPort)
- **Strategy Pattern** - ChunkingStrategy, ParserFactory
- **Circuit Breaker** - protection against backend API failures

## Main Components

### Domain Layer
- `DocumentParser` - document parser interface
- `ChunkingStrategy` - text chunking strategy
- `LinkDetector` - link detection between chunks
- `ImportTask` - import task (FILE, URL, TEXT)
- `DocumentChunk` - text fragment with index
- `ImportResult` - processing result

### Application Layer
- `ImportDocumentHandler` - orchestrator of the entire import process
- Ports: `InboundQueuePort`, `OutboundQueuePort`, `NoteCreatorPort`

### Infrastructure Layer
- `HybridChunker` - hybrid chunking algorithm (paragraphs + sliding window)
- `ParserFactory` - parser factory (Tika for files/URLs, Text for plain text)
- `TikaDocumentParser` - parser based on Apache Tika
- `TextDocumentParser` - simple text parser
- `AsynqInboundAdapter` - inbound queue adapter
- `AsynqOutboundAdapter` - outbound queue adapter
- `NoteCreatorHttpClient` - HTTP client for backend API
- `ResilientNoteCreator` - wrapper with circuit breaker and retry
- `RedisImportStateRepository` - state storage in Redis

## Algorithm

1. **Task retrieval** - reading from queue `asynq:import:document:pending`
2. **Parsing** - depending on task type:
   - FILE: base64 decode + Tika parsing
   - URL: download and Tika parsing
   - TEXT: use as-is
3. **Chunking** - splitting text into fragments using `HybridChunker`:
   - Short paragraphs → separate chunks
   - Long paragraphs → sliding window by sentences
   - Chunk overlap for context preservation
4. **Note creation** - for each chunk:
   - HTTP POST request to backend API
   - Retry on errors (3 attempts)
   - Circuit breaker on mass backend failure
5. **Link detection** - optional, if enabled in settings
6. **Result publication** - to queue `asynq:import:responses:pending`

## Configuration

### Environment Variables
- `REDIS_URI` - Redis connection URL (default: `redis://localhost:6379/0`)
- `BACKEND_URL` - Backend API URL for note creation (hardcoded: `http://backend:8080`)

### Chunking Settings (in ImportOptions)
- `chunkSize` - maximum chunk size in words
- `overlap` - overlap size between chunks in words
- `minChunkLength` - minimum chunk length in characters

## Running

### Via Docker Compose
```bash
# Main stack
docker-compose up -d source-text-handler

# Personal instance
docker-compose -f docker-compose.personal.yml up -d source-text-handler_personal
```

### Locally with Maven
```bash
# Build
mvn clean package

# Run
java -jar target/SourceTextHandler-1.0-SNAPSHOT.jar
```

## System Integration

### Incoming Tasks (Redis)
Queue: `asynq:import:document:pending`

Task format (JSON):
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

### Outgoing Results (Redis)
Queue: `asynq:import:responses:pending`

Result format:
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
The service calls the following endpoints:

**Note creation:**
```
POST /api/notes
Content-Type: application/json

{
  "title": "chunk title",
  "content": "chunk content",
  "metadata": {}
}
```

**Link creation:**
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

The service provides a health check on port 8081:
```
GET http://localhost:8081/health
```

Returns Redis connection status.

## Testing

### Test Task
```bash
# View test task
cat test-task.json

# Send to queue (requires redis-cli)
redis-cli -h localhost -p 6379 LPUSH "asynq:import:document:pending" "$(cat test-task.json)"
```

## Development

### Project Structure
```
source-text-handler/
├── src/main/java/com/alximac/knowledgegraph/texthandler/
│   ├── application/        # Business logic and ports
│   ├── domain/            # Domain model and services
│   ├── infrastructure/    # Infrastructure implementation
│   └── config/            # Application configuration
├── pom.xml                # Maven configuration
├── Dockerfile             # Docker image
└── README.md              # Documentation
```

### Adding a New Parser
1. Implement `DocumentParser` interface
2. Add to `ParserFactory`
3. Update `TaskType` if needed

### Changing Chunking Strategy
1. Implement `ChunkingStrategy` interface
2. Replace `HybridChunker` in `AppConfig`

## Troubleshooting

### Redis Connection Issues
- Check `REDIS_URI` variable
- Ensure Redis is available: `redis-cli ping`

### Note Creation Errors
- Ensure backend is accessible at the specified URL
- Check logs for HTTP request error details
- Circuit breaker may temporarily block requests during mass failures

### File Parsing Issues
- Ensure the file is supported by Apache Tika
- Ensure content is correctly base64-encoded for FILE type
- For URLs, ensure the resource is accessible

## License

The service is part of the Knowledge Graph project.
