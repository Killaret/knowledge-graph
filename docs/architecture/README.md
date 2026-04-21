# Архитектура Knowledge Graph

**Версия:** 1.0  
**Статус:** Проектирование завершено  
**Дата:** Март 2026 (актуально на момент написания)

---

## 📚 Оглавление

1. [Введение](#1-введение)
2. [Глоссарий](glossary.md)
3. [C4 Model](#c4-model)
   - [Контекст (Level 1)](c4/context.puml)
   - [Контейнеры (Level 2)](c4/container.puml)
   - [Компоненты (Level 3)](c4/component.puml)
4. [UML Диаграммы](#uml-диаграммы)
   - [Диаграмма развёртывания (локально)](uml/deployment-local.puml)
   - [Диаграмма развёртывания (Kubernetes)](uml/deployment-k8s.puml)
   - [Диаграмма классов (Domain)](uml/class-domain.puml)
   - [ER-диаграмма (модель данных)](uml/er-diagram.puml)
   - [Sequence: создание заметки](uml/sequence-create-note.puml)
   - [Sequence: рекомендации](uml/sequence-suggestions.puml)
5. [Архитектурные решения (ADR)](#5-архитектурные-решения-adr)
6. [ATAM анализ](atam.md)
7. [Конфигурация системы](../CONFIGURATION_EN.md)
8. [Иерархическая кластеризация (спецификация)](clustering.md)

---

## 1. Введение

Документация описывает архитектуру сервиса **Knowledge Graph** — базы знаний с графовой структурой, перекрёстными ссылками и интеллектуальными рекомендациями.

**Ключевые характеристики:**
- DDD с чёткими слоями
- CQRS (команды и запросы)
- Асинхронная обработка через очереди (asynq + Redis)
- Эмбеддинги через pgvector
- Рекомендации на основе BFS графа (глубина 3, затухание 0.5) **и семантического сходства эмбеддингов** с настраиваемыми весами.
- **Прогрессивный рендеринг**: Fog of War эффект, инкрементальная загрузка узлов, анимированная камера (Three.js)

---

## 2. Backend Архитектура

### 2.1 Структура проекта (DDD Layers)

```
backend/
├── cmd/
│   ├── server/           # HTTP API (Gin)
│   └── worker/           # Async task processor (asynq)
├── internal/
│   ├── domain/           # Entities, Value Objects
│   ├── application/      # Use cases, Command/Query handlers
│   ├── infrastructure/   # PostgreSQL, Redis, Asynq, NLP client
│   └── interfaces/       # HTTP handlers, DTOs
└── migrations/           # SQL migrations
```

### 2.2 Поток запроса (Request Flow)

```
HTTP Request → Router → Handler → Application → Domain → Infrastructure
                (Gin)    (DTO)     (Use Case)   (Entity)  (DB/Cache/Queue)
```

### 2.3 Async Tasks Flow

```
Handler → Enqueue task → Redis → Worker → Process → Save to DB
```

### 2.4 Recommendation Algorithm

```
Cache Check → BFS Links (α=0.5) + Semantic Search (β=0.5) → Combine → Rank
```

---

## 3. Технологический стек

| Компонент | Технология |
|-----------|-----------|
| Бэкенд | Go + Gin + GORM |
| Фронтенд | Svelte 5 + TypeScript |
| База данных | PostgreSQL 16 + pgvector |
| Кэш / очереди | Redis 7 + asynq |
| NLP | Python + FastAPI + sentence-transformers |
| Инфраструктура | Docker Compose (локально) → Kubernetes (продакшен) |

---

## 3. C4 Model

Диаграммы C4 описаны в формате PlantUML (`.puml`). Для просмотра установите расширение PlantUML в VSCode.

| Уровень | Описание | Файл |
|---------|----------|------|
| **Level 1: Context** | Система и внешние пользователи | [`c4/context.puml`](c4/context.puml) |
| **Level 2: Containers** | Приложения: backend, frontend, БД, Redis, NLP | [`c4/container.puml`](c4/container.puml) |
| **Level 3: Components** | DDD слои внутри backend | [`c4/component.puml`](c4/component.puml) |

---

## 4. UML Диаграммы

| Диаграмма | Описание | Файл |
|-----------|----------|------|
| **Deployment (local)** | Локальное развёртывание (Docker Compose) | [`uml/deployment-local.puml`](uml/deployment-local.puml) |
| **Deployment (k8s)** | Развёртывание в Kubernetes | [`uml/deployment-k8s.puml`](uml/deployment-k8s.puml) |
| **Class Diagram** | Domain модель (Note, Link, Value Objects) | [`uml/class-domain.puml`](uml/class-domain.puml) |
| **ER Diagram** | Таблицы PostgreSQL и связи | [`uml/er-diagram.puml`](uml/er-diagram.puml) |
| **Sequence (create)** | Создание заметки: синхронно + асинхронно | [`uml/sequence-create-note.puml`](uml/sequence-create-note.puml) |
| **Sequence (suggestions)** | Получение рекомендаций (BFS + кэш) | [`uml/sequence-suggestions.puml`](uml/sequence-suggestions.puml) |
---

## 5. Архитектурные решения (ADR)

Список принятых решений с обоснованием — в папке [`decisions/`](decisions/).

**Краткий список:**
1. DDD с чёткими слоями
2. CQRS (команды/запросы)
3. Link — отдельная сущность
4. Синхронное удаление связей (MVP)
5. Валидация: Value Objects + Application
6. Eventual consistency для read model
7. Value Objects (Title, Content, LinkType)
8. Эмбеддинги в Infrastructure
9. Domain Events + Event Bus
10. Миграции (папка `migrations/`)
11. Specifications в Domain слое
12. In-memory Bus (без middleware)
13. Эмбеддинги хранятся отдельно
14. Комбинирование явных связей и семантического сходства в рекомендациях (α, β)

---

## 6. ATAM анализ

Анализ ключевых сценариев (создание заметки, рекомендации, удаление заметки) — в файле [`atam.md`](atam.md).

---

## 7. Конфигурация системы

Все настраиваемые параметры описаны в отдельном документе:  
👉 [**CONFIGURATION_EN.md**](../CONFIGURATION_EN.md)

---

## 8. Как работать с документацией

1. Установите расширение **PlantUML** в VSCode
2. Откройте любой `.puml` файл → нажмите `Alt + D` для предпросмотра
3. Для экспорта в PNG: правой кнопкой → `PlantUML: Export Current Diagram`
4. Все изменения в документации коммитьте в Git вместе с кодом

---

## 9. Ссылки

- [PlantUML Language Guide](https://plantuml.com/guide)
- [C4 Model](https://c4model.com/)
- [pgvector documentation](https://github.com/pgvector/pgvector)
- [Asynq](https://github.com/hibiken/asynq)
- [Локальное развёртывание (Docker Compose)](uml/deployment-local.puml)
- [Развёртывание в Kubernetes](uml/deployment-k8s.puml)
- [Иерархическая кластеризация (спецификация)](clustering.md)