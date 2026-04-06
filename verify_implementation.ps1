# verify_implementation.ps1 - English only
$backendPath = Join-Path (Get-Location) "backend"

Write-Host "=== CHECK CODE VS DOCUMENTATION ===" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $backendPath)) {
    Write-Host "ERROR: backend folder not found" -ForegroundColor Red
    exit 1
}

Write-Host "1. KEY FILES" -ForegroundColor Green
$files = @(
    "cmd/server/main.go",
    "internal/config/config.go",
    "internal/domain/graph/traversal.go",
    "internal/application/graph/neighbor_loader.go",
    "internal/application/graph/embedding_loader.go",
    "internal/application/graph/composite_loader.go",
    "internal/application/queries/graph/get_suggestions.go",
    "internal/infrastructure/db/postgres/embedding_repo.go"
)
foreach ($f in $files) {
    $full = Join-Path $backendPath $f
    if (Test-Path $full) {
        Write-Host "  OK $f" -ForegroundColor Green
    } else {
        Write-Host "  MISS $f" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "2. CONFIG DEFAULTS (config.go)" -ForegroundColor Green
$configFile = Join-Path $backendPath "internal/config/config.go"
if (Test-Path $configFile) {
    $content = Get-Content $configFile -Raw
    $alpha = if ($content -match 'getFloatEnv\("RECOMMENDATION_ALPHA",\s*([\d.]+)\)') { $matches[1] } else { "not found" }
    $beta  = if ($content -match 'getFloatEnv\("RECOMMENDATION_BETA",\s*([\d.]+)\)') { $matches[1] } else { "not found" }
    $depth = if ($content -match 'getIntEnv\("RECOMMENDATION_DEPTH",\s*(\d+)\)') { $matches[1] } else { "not found" }
    $decay = if ($content -match 'getFloatEnv\("RECOMMENDATION_DECAY",\s*([\d.]+)\)') { $matches[1] } else { "not found" }
    $limit = if ($content -match 'getIntEnv\("EMBEDDING_SIMILARITY_LIMIT",\s*(\d+)\)') { $matches[1] } else { "not found" }
    Write-Host "  ALPHA = $alpha" -ForegroundColor Yellow
    Write-Host "  BETA = $beta" -ForegroundColor Yellow
    Write-Host "  DEPTH = $depth" -ForegroundColor Yellow
    Write-Host "  DECAY = $decay" -ForegroundColor Yellow
    Write-Host "  EMBEDDING_LIMIT = $limit" -ForegroundColor Yellow
} else {
    Write-Host "  MISS config.go" -ForegroundColor Red
}
Write-Host ""

Write-Host "3. NEIGHBOR LOADERS" -ForegroundColor Green
$traversalFile = Join-Path $backendPath "internal/domain/graph/traversal.go"
if (Test-Path $traversalFile) {
    if (Select-String -Path $traversalFile -Pattern "type NeighborLoader interface" -Quiet) {
        Write-Host "  OK NeighborLoader interface" -ForegroundColor Green
    } else {
        Write-Host "  MISS NeighborLoader interface" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS traversal.go" -ForegroundColor Red
}

$neighborFile = Join-Path $backendPath "internal/application/graph/neighbor_loader.go"
if (Test-Path $neighborFile) {
    if (Select-String -Path $neighborFile -Pattern "func.*GetNeighbors" -Quiet) {
        Write-Host "  OK neighborLoader.GetNeighbors" -ForegroundColor Green
    } else {
        Write-Host "  MISS neighborLoader.GetNeighbors" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS neighbor_loader.go" -ForegroundColor Red
}

$embeddingFile = Join-Path $backendPath "internal/application/graph/embedding_loader.go"
if (Test-Path $embeddingFile) {
    if (Select-String -Path $embeddingFile -Pattern "func.*GetNeighbors" -Quiet) {
        Write-Host "  OK embeddingLoader.GetNeighbors" -ForegroundColor Green
    } else {
        Write-Host "  MISS embeddingLoader.GetNeighbors" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS embedding_loader.go" -ForegroundColor Red
}

$compositeFile = Join-Path $backendPath "internal/application/graph/composite_loader.go"
if (Test-Path $compositeFile) {
    if (Select-String -Path $compositeFile -Pattern "NewCompositeNeighborLoaderWithWeights" -Quiet) {
        Write-Host "  OK composite loader with weights" -ForegroundColor Green
    } else {
        Write-Host "  MISS composite loader" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS composite_loader.go" -ForegroundColor Red
}
Write-Host ""

Write-Host "4. PGVECTOR SIMILARITY SEARCH" -ForegroundColor Green
$embedRepoFile = Join-Path $backendPath "internal/infrastructure/db/postgres/embedding_repo.go"
if (Test-Path $embedRepoFile) {
    if (Select-String -Path $embedRepoFile -Pattern "func.*FindSimilarNotes" -Quiet) {
        Write-Host "  OK FindSimilarNotes implemented" -ForegroundColor Green
        if (Select-String -Path $embedRepoFile -Pattern "<=>" -Quiet) {
            Write-Host "  OK uses cosine distance operator (<=>)" -ForegroundColor Green
        } else {
            Write-Host "  WARN operator <=> not found" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  MISS FindSimilarNotes" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS embedding_repo.go" -ForegroundColor Red
}
Write-Host ""

Write-Host "5. DECAY BEHAVIOR (second level only)" -ForegroundColor Green
if (Test-Path $traversalFile) {
    $content = Get-Content $traversalFile -Raw
    if ($content -match 'if current\.depth > 0 \{\s*newWeight \*= decay') {
        Write-Host "  OK decay applied only at depth > 0 (second level+)" -ForegroundColor Green
    } elseif ($content -match 'newWeight := current\.weight \* edge\.Weight \* decay') {
        Write-Host "  WARN decay applied at ALL levels (first level also). Not matching ADR." -ForegroundColor Yellow
    } else {
        Write-Host "  MISS decay logic not found" -ForegroundColor Red
    }
} else {
    Write-Host "  MISS traversal.go" -ForegroundColor Red
}
Write-Host ""

Write-Host "6. CONFIG USAGE IN main.go" -ForegroundColor Green
$mainFile = Join-Path $backendPath "cmd/server/main.go"
if (Test-Path $mainFile) {
    if (Select-String -Path $mainFile -Pattern "config\.Load\(\)" -Quiet) {
        Write-Host "  OK main.go uses config.Load()" -ForegroundColor Green
    } else {
        Write-Host "  MISS config.Load() not used" -ForegroundColor Red
    }
    if (Select-String -Path $mainFile -Pattern "NewCompositeNeighborLoaderWithWeights" -Quiet) {
        Write-Host "  OK uses composite loader with weights" -ForegroundColor Green
    } else {
        Write-Host "  WARN composite loader not used (maybe old code)" -ForegroundColor Yellow
    }
} else {
    Write-Host "  MISS main.go" -ForegroundColor Red
}
Write-Host ""

Write-Host "7. ENV VARS IN docker-compose.yml (backend)" -ForegroundColor Green
$composeFile = Join-Path (Get-Location) "docker-compose.yml"
if (Test-Path $composeFile) {
    $content = Get-Content $composeFile -Raw
    $vars = @("RECOMMENDATION_ALPHA", "RECOMMENDATION_BETA", "RECOMMENDATION_DEPTH", "RECOMMENDATION_DECAY", "EMBEDDING_SIMILARITY_LIMIT")
    foreach ($var in $vars) {
        if ($content -match "${var}:") {
            Write-Host "  OK ${var} defined" -ForegroundColor Green
        } else {
            Write-Host "  WARN ${var} NOT defined (using defaults from config.go)" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "  MISS docker-compose.yml" -ForegroundColor Red
}
Write-Host ""

Write-Host "=== CHECK COMPLETE ===" -ForegroundColor Cyan