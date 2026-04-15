# 📚 Knowledge Graph — Центральная навигация

Этот репозиторий содержит документацию и исходный код для системы **Knowledge Graph** — интеллектуальной базы знаний с графовой структурой и семантическими рекомендациями.

🌐 **Языки**: [Русский](README.md) | [English](README_EN.md)

---

## 📁 Основные разделы

### 🏗️ Архитектура
- [Архитектура системы (C4, UML, компоненты)](docs/architecture/README.md)
- [Архитектурные решения (ADR)](docs/architecture/adr.md)
- [ATAM анализ](docs/architecture/atam.md)
- [Глоссарий](docs/architecture/glossary.md)

### 📚 Документация
- [Конфигурация системы](docs/CONFIGURATION.md) (Русский) | [English](docs/CONFIGURATION_EN.md)
- [Руководство по развёртыванию](docs/DEPLOYMENT.md) (Русский) | [English](docs/DEPLOYMENT_EN.md)
- [UX Guidelines](docs/UX_GUIDELINES.md)
- [Идеи для развития](docs/IDEAS.md)
- [Основной README по архитектуре](docs/architecture/README.md)

### 🛠️ Настройка и запуск
- [Руководство по настройке](docs/CONFIGURATION.md)
- [Запуск через Docker Compose](docker-compose.yml)

### 📂 Структура проекта
```
├── backend/          # Go-бэкенд (Gin, GORM, DDD)
├── frontend/         # Svelte 5 фронтенд
├── nlp-service/      # Python-сервис для эмбеддингов (FastAPI)
├── docs/             # Полная документация
│   ├── CONFIGURATION_EN.md    # English version
│   ├── CONFIGURATION.md       # Russian version
│   ├── DEPLOYMENT_EN.md       # English version
│   └── DEPLOYMENT.md          # Russian version
├── init-db/          # SQL-скрипты инициализации базы
├── scripts/          # Вспомогательные скрипты
├── docker-compose.yml # Оркестрация сервисов
├── README.md         # Russian README
└── README_EN.md      # English README
```

## 🔗 Полезные ссылки
- [PlantUML Language Guide](https://plantuml.com/guide)
- [C4 Model](https://c4model.com/)
- [pgvector documentation](https://github.com/pgvector/pgvector)
- [Asynq](https://github.com/hibiken/asynq)

---
📄 Последнее обновление: 2026-04-15

## 🚀 Быстрый старт

### Требования
- Docker 20.10+ и Docker Compose 2.0+
- 4GB RAM минимум (8GB рекомендуется)
- 10GB свободного места

### Запуск
```bash
# Клонирование репозитория
git clone https://github.com/your-org/knowledge-graph.git
cd knowledge-graph

# Запуск всех сервисов
docker-compose up -d

# Проверка логов
docker-compose logs -f backend
```

### Проверка установки
```bash
# Проверка backend
curl http://localhost:8080/health

# Проверка NLP сервиса
curl http://localhost:5000/health

# Доступ к приложению
open http://localhost:8080
```
