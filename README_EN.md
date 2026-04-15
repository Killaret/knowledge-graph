# 📚 Knowledge Graph — Central Navigation

This repository contains documentation and source code for the **Knowledge Graph** system — an intelligent knowledge base with graph structure and semantic recommendations.

🌐 **Languages**: [English](README_EN.md) | [Русский](README.md)

---

## 📁 Main Sections

### 🏗️ Architecture
- [System Architecture (C4, UML, Components)](docs/architecture/README.md)
- [Architectural Decision Records (ADR)](docs/architecture/adr.md)
- [ATAM Analysis](docs/architecture/atam.md)
- [Glossary](docs/architecture/glossary.md)

### 📚 Documentation
- [System Configuration](docs/CONFIGURATION_EN.md) (English) | [Русский](docs/CONFIGURATION.md)
- [UX Guidelines](docs/UX_GUIDELINES.md)
- [Ideas for Development](docs/IDEAS.md)

### 🛠️ Setup and Deployment
- [Configuration Guide](docs/CONFIGURATION_EN.md) (English) | [Русский](docs/CONFIGURATION.md)
- [Deployment Guide](docs/DEPLOYMENT_EN.md) (English) | [Русский](docs/DEPLOYMENT.md)
- [Docker Compose Launch](docker-compose.yml)

### 📂 Project Structure
```
├── backend/          # Go backend (Gin, GORM, DDD)
├── frontend/         # Svelte 5 frontend
├── nlp-service/      # Python embeddings service (FastAPI)
├── docs/             # Full documentation
│   ├── CONFIGURATION_EN.md    # English version
│   ├── CONFIGURATION.md       # Russian version
│   ├── DEPLOYMENT_EN.md       # English version
│   └── DEPLOYMENT.md          # Russian version
├── init-db/          # Database initialization SQL scripts
├── scripts/          # Helper scripts
└── docker-compose.yml # Service orchestration
```

---

## 🚀 Quick Start

### Prerequisites
- Docker 20.10+ and Docker Compose 2.0+
- 4GB RAM minimum (8GB recommended)
- 10GB free disk space

### Launch
```bash
# Clone repository
git clone https://github.com/your-org/knowledge-graph.git
cd knowledge-graph

# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f backend
```

### Verify Installation
```bash
# Backend health
curl http://localhost:8080/health

# NLP service health
curl http://localhost:5000/health

# Access application
open http://localhost:8080
```

---

## 🌟 Key Features

- **3D Progressive Rendering**: Fog of War effect with incremental node loading
- **Semantic Recommendations**: Combined graph BFS (α=0.5) + embeddings (β=0.5)
- **Async Processing**: asynq queue for NLP tasks (keywords, embeddings)
- **Full Text Search**: PostgreSQL + pgvector for semantic similarity
- **Modern Stack**: Go (Gin), Svelte 5, Three.js, Python (FastAPI)

---

## 🔗 Useful Links

- [PlantUML Language Guide](https://plantuml.com/guide)
- [C4 Model](https://c4model.com/)
- [pgvector documentation](https://github.com/pgvector/pgvector)
- [Asynq](https://github.com/hibiken/asynq)

---

📄 Last updated: 2026-04-15
