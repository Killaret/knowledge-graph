# Configuration Guide

This document describes the configuration system for Knowledge Graph project.

## Overview

The project uses a hierarchical configuration system with the following priority:

```
Environment Variables > knowledge-graph.config.json > Hard-coded Defaults
```

## Configuration Files

### 1. `knowledge-graph.config.json`

Central JSON configuration file located in the project root. Used by both backend and frontend.

**Structure:**
```json
{
  "backend": {
    "server": { "rate_limit": {...} },
    "database": {...},
    "search": {...},
    "recommendation": {...},
    "pagination": {...},
    "graph": {...},
    "embedding": {...},
    "asynq": {...}
  },
  "frontend": {
    "test": {...},
    "graph": { "2d": {...}, "3d": {...} },
    "api": {...}
  },
  "ci_cd": {...},
  "nlp": {...}
}
```

### 2. `.env` / `.env.example`

Environment variables for sensitive data and deployment-specific settings.

**Required Variables:**
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `NLP_SERVICE_URL` - NLP service endpoint

**Optional Variables (override JSON config):**
- `SERVER_PORT` - Backend port (default: 8080)
- `SERVER_RATE_LIMIT_ENABLED` - Enable rate limiting
- `SERVER_RATE_LIMIT_REQUESTS` - Rate limit requests
- `RECOMMENDATION_DEPTH` - BFS depth for recommendations
- `RECOMMENDATION_ALPHA` - Weight for link component
- `RECOMMENDATION_BETA` - Weight for semantic component
- And many more...

See `.env.example` for complete list.

## Backend Configuration (Go)

### Loading Order

1. Load `knowledge-graph.config.json` (if exists)
2. Check environment variables (override JSON values)
3. Use hard-coded defaults (if neither set)

### Usage in Code

```go
import "knowledge-graph/internal/config"

cfg := config.Load()

// Access configuration
depth := cfg.RecommendationDepth
ttl := cfg.RecommendationCacheTTL
```

### Adding New Configuration Parameter

1. Add field to `JSONConfig` struct in `backend/internal/config/config.go`:
```go
type JSONConfig struct {
    Backend struct {
        // ... existing fields ...
        NewSection struct {
            NewParam int `json:"new_param"`
        } `json:"new_section"`
    }
}
```

2. Add field to `Config` struct:
```go
type Config struct {
    // ... existing fields ...
    NewParam int
}
```

3. Add loading logic in `Load()`:
```go
cfg.NewParam = getIntEnv("NEW_PARAM", getJSONIntOrDefault(jsonCfg, 
    func(j *JSONConfig) int { return j.Backend.NewSection.NewParam }, 
    42)) // default value
```

4. Add to `knowledge-graph.config.json`:
```json
{
  "backend": {
    "new_section": {
      "new_param": 42
    }
  }
}
```

5. Document in `.env.example`:
```bash
# New Section
NEW_PARAM=42
```

## Frontend Configuration (TypeScript)

### Loading

Frontend imports JSON directly at build time:

```typescript
import config from '../../../knowledge-graph.config.json';

export const apiConfig = config.frontend.api;
export const graphConfig = config.frontend.graph;
```

### Limitations

- Frontend only uses a subset of configuration
- Changes require rebuild
- No runtime environment variable support (use build args for different environments)

## Docker Compose Configuration

### Development Stack (docker-compose.yml)

- Port: 3000
- Full stack with all services
- Shared volumes for development

### Personal Instance (docker-compose.personal.yml)

- Port: 3001
- Isolated database and Redis
- Persistent volumes

## Validation & CI/CD

### Local Validation

```bash
# Backend
cd backend
make check-config        # Validate JSON config
cd backend
make check-migrations    # Check migration drift

# Manual checks
# Validate JSON syntax
cat knowledge-graph.config.json | jq . > /dev/null

# Check environment variables
source .env && go run backend/cmd/server/main.go
```

### CI/CD Checks

The following checks run automatically:

1. **Config Validation** (`.github/workflows/ci-config-validation.yml`)
   - JSON syntax validation
   - Config loads successfully
   - All env vars documented in `.env.example`

2. **Migration Drift Check**
   - Applies migrations to test database
   - Verifies GORM AutoMigrate compatibility

3. **Docker Compose Validation**
   - Syntax validation
   - Environment variable consistency

## Troubleshooting

### Config Not Loading

1. Check file exists in project root: `knowledge-graph.config.json`
2. Validate JSON syntax: `cat knowledge-graph.config.json | jq .`
3. Check backend logs for "[Config] Loading JSON config from..."

### Environment Variables Not Applied

1. Verify variable name matches exactly (case-sensitive)
2. Check `.env.example` for correct naming
3. Ensure variable is exported: `export VAR_NAME=value`

### Migration Errors

1. Check database connection: `pg_isready -h localhost -p 5432`
2. Verify migrations are applied: `migrate -path backend/migrations -database "$DATABASE_URL" version`
3. Check for drift: `make check-migrations`

## Security Considerations

- Never commit `.env` files with real credentials
- Use `.env.example` as template with dummy values
- Store sensitive data in environment variables only
- Rotate secrets regularly in production
