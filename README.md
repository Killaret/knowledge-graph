# 🌌 Knowledge Graph

<div align="center">

![Knowledge Graph](https://img.shields.io/badge/Knowledge-Graph-blue?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=for-the-badge&logo=svelte)
![TypeScript](https://img.shields.io/badge/TypeScript-5-blue?style=for-the-badge&logo=typescript)
![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=for-the-badge&logo=python)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**База знаний с графовой структурой и интеллектуальными рекомендациями**

[Features](#-features) • [Quick Start](#-quick-start) • [Architecture](#-architecture) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## ✨ Features

- 🌟 **3D Visualization** — заметки как небесные тела в интерактивном космосе
- 🔗 **Graph Structure** — перекрёстные ссылки между заметками
- 🧠 **Smart Recommendations** — интеллектуальные рекомендации на основе NLP
- 🎨 **Celestial Types** — звёзды, планеты, кометы, галактики для разных типов контента
- 🔍 **Semantic Search** — полнотекстовый поиск с pgvector
- 📝 **Draft System** — автосохранение черновиков в MongoDB
- 🔒 **Authentication** — JWT, OAuth2, API keys
- 🎯 **Achievements** — геймификация с системой достижений
- 🌐 **Multi-language** — поддержка русского и английского
- 💾 **Cloud Backup** — резервное копирование на Яндекс.Диск

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Go 1.23+, Node.js 20+, Python 3.11+, Java 17+

### One-Command Start

```bash
# Full stack with Docker
docker-compose up -d

# Personal instance on port 3001
docker-compose -f docker-compose.personal.yml up -d
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Development Mode

```bash
# Backend (Go)
cd backend && go run ./cmd/server

# Frontend (Svelte)
cd frontend && npm run dev

# NLP Service (Python)
cd nlp-service && uvicorn app.main:app --reload

# Source Text Handler (Java)
cd source-text-handler && mvn spring-boot:run
```

---

## 📖 Table of Contents

- [Architecture](#-architecture)
- [Technology Stack](#-technology-stack)
- [Project Structure](#-project-structure)
- [Documentation](#-documentation)
- [Security](#-security)
- [AI Agents](#-ai-agents)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🏗️ Architecture

### Architecture Patterns

- **Clean Architecture** — разделение на Domain, Application, Infrastructure, Interfaces
- **Domain-Driven Design (DDD)** — богатая доменная модель с Value Objects
- **CQRS-Lite** — оптимизация чтения/записи
- **Event-Driven** — инкрементальные обновления графа через события

### High-Level Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │    │   Backend   │    │  NLP Service│
│  (Svelte 5) │◄──►│  (Go 1.23)  │◄──►│  (Python)    │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Browser   │    │ PostgreSQL  │    │   MongoDB   │
│  (3D Three.js)│   │ + pgvector  │    │   (Drafts)  │
└─────────────┘    └─────────────┘    └─────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │    Redis    │
                   │ (Cache/Queue)│
                   └─────────────┘
```

---

## 💻 Technology Stack

### Backend
- **Language:** Go 1.23+
- **Framework:** Gin + GORM
- **Database:** PostgreSQL 16 + pgvector
- **Cache/Queue:** Redis 7 + asynq
- **Auth:** JWT, OAuth2 (Yandex), API Keys
- **Graph Service:** gRPC (microservice)

### Frontend
- **Framework:** SvelteKit (Svelte 5)
- **Language:** TypeScript
- **3D Graphics:** Three.js
- **State:** Svelte stores
- **Testing:** Vitest + Playwright

### NLP Service
- **Language:** Python 3.11+
- **Framework:** FastAPI
- **ML:** sentence-transformers, YAKE, NLTK

### Source Text Handler
- **Language:** Java 17+
- **Framework:** Spring Boot
- **Build:** Maven
- **Features:** Document parsing, chunking, retry + circuit breaker

### Infrastructure
- **Containerization:** Docker Compose
- **Reverse Proxy:** Nginx
- **Monitoring:** Prometheus (planned)
- **Backup:** Яндекс.Диск WebDAV

---

## 📁 Project Structure

```
knowledge-graph/
├── backend/                 # Go backend (REST API, workers)
│   ├── cmd/                # Server & Worker entry points
│   ├── internal/           # DDD layers
│   │   ├── domain/         # Business logic
│   │   ├── application/    # Use cases
│   │   ├── infrastructure/ # DB, Redis, config
│   │   └── interfaces/     # HTTP handlers
│   └── migrations/         # SQL migrations
├── frontend/               # SvelteKit frontend
│   ├── src/
│   │   ├── lib/           # Business logic, API clients
│   │   ├── components/    # UI components
│   │   └── routes/        # SvelteKit pages
│   └── tests/             # Playwright E2E tests
├── nlp-service/           # Python NLP service (embeddings)
├── source-text-handler/   # Java document processing
├── services/
│   └── graph-service/     # gRPC graph layout service
├── docs/                  # Architecture, ADR, UML diagrams
├── scripts/               # Utility scripts (cleanup, diagnostics)
└── tests/                 # BDD tests (Cucumber + Playwright)
```

---

## 📚 Documentation

### Core Documentation
- [🎯 Roadmap](docs/ROADMAP.md) — план развития продукта
- [📐 Architecture](docs/architecture/README.md) — C4 модель, UML, ADR
- [🚀 Deployment](docs/DEPLOYMENT_EN.md) — руководство по развертыванию
- [⚙️ Configuration](docs/CONFIGURATION.md) — настройка системы

### Feature Documentation
- [🔐 Authentication](backend/internal/auth/README.md) — система авторизации
- [🎯 Achievements](backend/internal/application/achievement/) — геймификация
- [💾 Backup](docs/BACKUP.md) — резервное копирование
- [🔍 NLP Integration](docs/RECOMMENDATION_ARCHITECTURE.md) — рекомендации

### Developer Guides
- [🤖 AI Agents](docs/AGENTS.md) — использование AI агентов
- [📝 Commands](COMMANDS.md) — справочник команд
- [🧪 Testing](TEST_STATUS.md) — статус и покрытие тестов
- [🎨 Frontend Patterns](frontend/FRONTEND_PATTERNS.md) — паттерны фронтенда

### Service Documentation
- [📄 Source Text Handler](source-text-handler/README.md) — обработка документов
- [🔄 API Errors](docs/API_ERRORS.md) — формат ошибок API
- [🗄️ Database Schema](docs/SaaS_DATABASE_SCHEMA.md) — схема БД

---

## 🔒 Security

### Dependency Protection
- ✅ **npm ci only** in CI (never npm install)
- ✅ **minimumReleaseAge=7** in `.npmrc`
- ✅ **npm audit** with high severity threshold
- ✅ **package-lock.json** controlled by CODEOWNERS
- ✅ **Lifecycle scripts whitelist**

### Automated Security
- ✅ **Dependabot** — weekly dependency updates
- ✅ **Dependency Review Action** — PR vulnerability scanning
- ✅ **Daily security scans** — automated audits

### GitHub & CI/CD Security
- ✅ **CODEOWNERS** — mandatory review for dependency changes
- ✅ **Minimal permissions** — `contents: read` for GitHub Actions
- ✅ **Branch protection** — approval required for main/release
- ✅ **Secret scanning** & **push protection** enabled

---

## 🤖 AI Agents

This project uses a comprehensive AI agent ecosystem:

### Available Agents (9)
| Agent | Focus | Tools |
|-------|-------|-------|
| **Orchestrator** | Coordination | All tools |
| **Backend Go** | Go API, DB | `backend-go-tools.md` |
| **Frontend Svelte** | Svelte 5, UI | `frontend-tools.md` |
| **Python NLP** | Python, ML | `python-nlp-tools.md` |
| **Integration** | API contracts | `integration-tools.md` |
| **Infrastructure** | Docker, K8s | `infrastructure-tools.md` |
| **DevOps** | CI/CD, backup | `devops-tools.md` |
| **Performance** | Optimization | `performance-tools.md` |
| **Security** | Security, auth | `security-tools.md` |

### Primary AI Assistant
**[Devin](https://cli.devin.ai/docs)** by Cognition — main AI development assistant with full tool access for autonomous coding, debugging, and CI/CD integration.

---

## 🛠️ Development

### Setup

```bash
# Clone repository
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph

# Install dependencies
cd backend && go mod download
cd ../frontend && npm install
cd ../nlp-service && pip install -r requirements.txt
cd ../source-text-handler && mvn clean install
```

### Local Development

```bash
# Start services with Docker
docker-compose up -d

# Or run individually:
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm run dev

# NLP Service
cd nlp-service && uvicorn app.main:app --reload

# Source Text Handler
cd source-text-handler && mvn spring-boot:run
```

### Code Quality

```bash
# Backend linting
cd backend && golangci-lint run

# Frontend linting
cd frontend && npm run lint

# Type checking
cd frontend && npm run check
```

---

## 🧪 Testing

### Unit Tests

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm run test:unit

# NLP Service
cd nlp-service && pytest
```

### Integration Tests

```bash
# Backend integration
cd backend && go test -tags=integration ./...

# Frontend E2E
cd frontend && npm run test

# BDD tests
cd tests && npm run test:bdd
```

### Test Coverage
- **Backend:** >85% coverage required
- **Frontend:** >60% coverage required
- **Overall:** All tests must pass before merge

---

## 🚀 Deployment

### Docker Deployment

```bash
# Production stack
docker-compose -f docker-compose.prod.yml up -d

# Personal instance
docker-compose -f docker-compose.personal.yml up -d

# CI/CD testing
docker-compose -f docker-compose.test.yml up -d
```

### Environment Variables

See `.env.example` and `knowledge-graph.config.json` for configuration options.

### Backup & Restore

```bash
# Manual backup
./scripts/utility/backup-personal.sh

# Automatic backup (configured in docker-compose)
# Back up to Яндекс.Дisk via WebDAV
```

---

## 🤝 Contributing

### Development Workflow

1. **Fork** the repository
2. **Create branch** (`git checkout -b feature/amazing-feature`)
3. **Commit** changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to branch (`git push origin feature/amazing-feature`)
5. **Open Pull Request**

### Commit Messages

Follow conventional commits:
- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation
- `refactor:` code refactoring
- `test:` testing
- `chore:` maintenance

### Code Review

- All PRs require approval from maintainers
- Security changes require additional review
- Dependency changes require CODEOWNERS approval

### Project Rules

- **Backend:** Clean Architecture, DDD, no globals, dependency injection
- **Frontend:** Atomic design, Svelte 5, TypeScript strict mode
- **Infrastructure:** Docker multi-stage builds, health checks
- **Testing:** >60% coverage required
- **Security:** Never commit secrets, use environment variables

See [`AGENTS.md`](docs/AGENTS.md) for detailed development guidelines.

---

## 📊 Project Status

### Recent Updates
- ✅ Fixed Playwright and svelte-check errors
- ✅ Updated GraphDelta types for incremental updates
- ✅ Added anomaly configuration types
- ✅ Enhanced CI/CD pipeline security
- ✅ Improved docker-compose configurations
- ✅ Added comprehensive AI agent documentation

### Current Focus
- 🚀 Performance optimization for graph rendering
- 🔐 Enhanced authentication and authorization
- 🎯 Achievement system implementation
- 📄 Document processing pipeline (Java)
- 🌐 Multi-language support improvements

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **Three.js** for 3D graphics
- **Svelte** for reactive UI
- **Gin** for Go web framework
- **PostgreSQL + pgvector** for vector similarity search
- **sentence-transformers** for NLP embeddings
- **Devin AI** for AI-assisted development

---

## 📞 Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/Killaret/knowledge-graph/issues)
- 💬 [Discussions](https://github.com/Killaret/knowledge-graph/discussions)
- 📧 Email: (see repository contact)

---

<div align="center">

**Built with ❤️ using Devin AI**

![Star](https://img.shields.io/github/stars/Killaret/knowledge-graph?style=social)
![Fork](https://img.shields.io/github/forks/Killaret/knowledge-graph?style=social)
![Watch](https://img.shields.io/github/watchers/Killaret/knowledge-graph?style=social)

</div>