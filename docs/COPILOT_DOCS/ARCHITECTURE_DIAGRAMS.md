# Визуальное представление процесса

## 1. Процесс создания и обработки заметки

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENT SENDS REQUEST                         │
│                   POST /notes (with content)                     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
        ┌────────────────────────────────────────┐
        │          BACKEND SERVICE               │
        │     (kg-backend on port 8080)          │
        │                                        │
        │  1. Parse JSON request                 │
        │  2. Validate data                      │
        │  3. Create Note object                 │
        └────────────────────┬───────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
        ┌──────────────────┐    ┌──────────────────────────┐
        │  PostgreSQL      │    │   Redis via Asynq        │
        │  (kg-postgres)   │    │   (kg-redis)             │
        │                  │    │                          │
        │  Save note to DB │    │  Enqueue 2 tasks:        │
        │                  │    │  ├─ extract:keywords     │
        │ Database: notes  │    │  └─ compute:embedding    │
        └──────────────────┘    └──────────────┬───────────┘
                                               │
                                    ┌──────────▼──────────┐
                                    │  REDIS QUEUE        │
                                    │  (asynq:queues)     │
                                    │                     │
                                    │  Task 1: Keywords   │
                                    │  Task 2: Embedding  │
                                    └──────────┬──────────┘
                                               │
                                    ┌──────────▼──────────┐
                                    │  WORKER SERVICE      │
                                    │  (kg-worker)         │
                                    │                      │
                                    │ Consumes tasks from  │
                                    │ Redis queue          │
                                    └──────────┬──────────┘
                       ┌────────────────────────┴────────────────────┐
                       │                                             │
                       ▼                                             ▼
          ┌──────────────────────────┐          ┌─────────────────────────┐
          │  EXTRACT KEYWORDS        │          │  COMPUTE EMBEDDING      │
          │                          │          │                         │
          │ 1. Get note from DB      │          │ 1. Get note from DB    │
          │ 2. Call NLP service      │          │ 2. Call NLP service    │
          │ 3. Receive keywords      │          │ 3. Receive embedding   │
          │ 4. Save to DB            │          │ 4. Save to DB          │
          │    (note_keywords)       │          │    (embeddings)        │
          └──────────────────────────┘          └─────────────────────────┘
                       │                                             │
                       └────────────────────┬──────────────────────┘
                                            │
                                 ┌──────────▼──────────┐
                                 │  TASK COMPLETED     │
                                 │  Both results saved │
                                 │  to PostgreSQL      │
                                 └─────────────────────┘
```

## 2. Архитектура Docker контейнеров

```
┌─────────────────────────────────────────────────────────────────┐
│                    DOCKER COMPOSE NETWORK                        │
└─────────────────────────────────────────────────────────────────┘

    ┌─────────────────┐
    │   kg-postgres   │     Database
    │   (PG 16)       │     ├─ notes
    │                 │     ├─ note_keywords
    │   Port: 5432    │     └─ embeddings
    │   (internal)    │
    └────────┬────────┘
             │
    ┌────────┴────────────────────────────────────────────┐
    │                                                     │
    ▼                                                     ▼
┌──────────────────┐                        ┌──────────────────────┐
│   kg-redis       │                        │   kg-backend         │
│   (Redis 7)      │                        │   (Go API Server)    │
│                  │                        │                      │
│ Port: 6379 ◄────┼─ connects ─────────┐   │ Port: 8080 ◄─ HTTP   │
│                  │ (Read/Write)       │   │                      │
└──────────────────┘                    │   │ - REST API           │
                                        │   │ - Send tasks to      │
                                        │   │   Redis via Asynq    │
                                        │   └──────────┬──────────┘
                                        │              │
                                        │   ┌──────────▼──────────┐
                                        │   │   kg-worker        │
                                        │   │   (Go Worker)      │
                                        └──┼─  Consume tasks    │
                                           │   from Redis       │
                                           │                    │
                                           │ - Extract keywords│
                                           │ - Compute embed  │
                                           │ - Save to DB     │
                                           │                    │
                                           └────────┬───────────┘
                                                    │
                                           ┌────────▼──────────┐
                                           │   kg-nlp          │
                                           │   (Python)        │
                                           │                   │
                                           │ Port: 5000        │
                                           │ - Extract keywords│
                                           │ - Compute embeddings
                                           │   (HuggingFace     │
                                           │    all-MiniLM-L6)  │
                                           └───────────────────┘
```

## 3. Жизненный цикл async task

```
┌──────────────┐
│  Task Created │  ← Enqueued from Backend
└──────┬───────┘
       │
       ▼
┌──────────────┐     ┌─────────────────────────────────────┐
│  In Queue    │────→│ Waiting for Worker to pick it up    │
└──────┬───────┘     │ (typically < 1 second)              │
       │             └─────────────────────────────────────┘
       │
       ▼
┌──────────────┐     ┌─────────────────────────────────────┐
│  Processing  │────→│ Worker is actively processing       │
└──────┬───────┘     │ (may call external services)        │
       │             └─────────────────────────────────────┘
       │
       ├─────────────┐
       │             │
       ▼             ▼
    ┌─────┐     ┌──────────┐
    │ OK  │     │ FAILURE  │
    └──┬──┘     └──┬───────┘
       │           │
       ▼           ▼
    DONE        RETRY or DROP
       │           │
       └─────┬─────┘
             │
             ▼
        ┌─────────────┐
        │ Task ended  │
        │ (success or │
        │  final fail)│
        └─────────────┘
```

## 4. Параллельная обработка задач

```
Time ──────────────────────────────────────────────────────────────>

Client sends: POST /notes
│
Backend processes:
├─ Save to DB: 1ms
├─ Enqueue Task 1: extract:keywords  ─┐
└─ Enqueue Task 2: compute:embedding  ─┼─→ Both go to Redis
                                         │
                                    ┌────▼──────────┐
                                    │  Redis Queue  │
                                    │               │
                                    │ [Task 1]      │
                                    │ [Task 2]      │
                                    └────┬──────────┘
                                         │
                        ┌────────────────┴────────────────┐
                        │                                 │
                        ▼ (Worker with Concurrency: 10)  ▼
                    Worker Thread 1             Worker Thread 2
                    Processing Task 1:          Processing Task 2:
                    extract:keywords            compute:embedding
                    (calls NLP)                 (calls NLP)
                    Takes ~500ms                Takes ~800ms
                        │                           │
                        └─── Both save results ─────┘
                                    │
                                    ▼
                            PostgreSQL updates
                            ├─ note_keywords
                            └─ embeddings

PARALLELISM: Tasks are processed concurrently! ✓
CONCURRENCY: 10 workers handling multiple tasks at once ✓
```

## 5. Состояния обработки в логах

```
BACKEND LOGS:
──────────────
[1] taskQueue is nil? false
[2] Enqueuing tasks for note <uuid>
[3] EnqueueExtractKeywords called for note <uuid>
[4] Marshal error: <none>
[5] Enqueue error: <none>
[6] Task enqueued: TaskID:..., Queue:default, State:enqueued
[7] EnqueueComputeEmbedding called for note <uuid>
[8] Task enqueued: TaskID:..., Queue:default, State:enqueued


REDIS:
──────
asynq:queues:default         ← Task queue name
asynq:t:<task-id>            ← Task metadata
asynq:processed:<task-id>    ← Processed task ID


WORKER LOGS:
────────────
[1] HandleExtractKeywords: received task {"note_id":"<uuid>","top_n":10}
[2] HandleExtractKeywords: extracted 5 keywords for note <uuid>
[3] HandleExtractKeywords: successfully processed note <uuid> with 5 keywords

[4] HandleComputeEmbedding: received task {"note_id":"<uuid>"}
[5] HandleComputeEmbedding: found note <uuid>, processing...
[6] HandleComputeEmbedding: computed embedding for note <uuid> (size=384)
[7] HandleComputeEmbedding: successfully processed note <uuid>


POSTGRESQL:
───────────
INSERT INTO note_keywords (note_id, keyword, weight)
VALUES ('<uuid>', 'keyword1', 0.95), ('keyword2', 0.87), ...

INSERT INTO embeddings (note_id, embedding)
VALUES ('<uuid>', '[0.123, 0.456, ...]'::vector)
```

---

## Файлы для справки:

- 📖 README_ASYNC_SETUP.md - Полное руководство
- ⚡ QUICK_START.md - Быстрый старт
- ✅ REBUILD_CHECKLIST.md - Диагностика
- 📋 SUMMARY.md - Итоговый отчет
