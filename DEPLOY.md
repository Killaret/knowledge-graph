# Knowledge Graph — Deployment Quick Start

> This is the entry point for local deployment.  
> The full Russian-language local deployment guide is in [`DEPLOY.ru.md`](DEPLOY.ru.md).  
> For production, Kubernetes, and CI/CD details see [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md).

## TL;DR

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
cp .env.example .env

# Edit .env: set JWT_SECRET and DB passwords

# Download the NLP model once
docker compose -f docker-compose.personal.yml run --rm -e HF_HUB_OFFLINE=0 nlp-personal python - <<'PY'
from sentence_transformers import SentenceTransformer
SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')
PY

# Start the Personal stack (your private data)
docker compose -f docker-compose.personal.yml up -d --build
```

## Where to find detailed instructions

| Language | Scope | File |
|---|---|---|
| Russian | Local dev/personal/test stacks, `.env`, model cache, ports, troubleshooting | [`DEPLOY.ru.md`](DEPLOY.ru.md) |
| English | Production, K8s, CI/CD, monitoring | [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md) |
| Any | Testing and regression | [`docs/TESTING.md`](docs/TESTING.md) |
| Any | Backups | [`docs/BACKUP.md`](docs/BACKUP.md) |
| Any | Environment variables and runtime config | [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md) |

## Important safety rules

- **Personal volumes** (`pgdata_personal`, `redisdata_personal`, `mongodbdata_personal`) contain live data. Never delete them without a fresh backup.
- **`.env` is not committed** and holds all secrets.
- **Run E2E/BDD only on the isolated test stack**, never on Personal or Dev.
- On Windows, use `http://127.0.0.1` URLs for Playwright, not `localhost`.
