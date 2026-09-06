# Full Test Cycle - Windows PowerShell (Isolated Testing Model)
# This script orchestrates the complete testing cycle with full stack isolation.
# Dev and personal stacks are stopped during testing to prevent resource conflicts.
# All temporary snapshots are saved to scripts/testing/temp/snapshots/.

param(
    [switch]$SkipManual
)

$frontendPort = if ($env:FRONTEND_PORT) { $env:FRONTEND_PORT } else { "3002" }
$frontendUrl = "http://127.0.0.1:$frontendPort"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Knowledge Graph Full Test Cycle" -ForegroundColor Cyan
Write-Host "  (Isolated Testing Model)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "For comprehensive regression testing, see docs/REGRESSION_TEST_PLAN.md" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️  WARNING: Dev and personal stacks will be stopped during testing" -ForegroundColor Yellow
Write-Host ""

$scriptDir = $PSScriptRoot
$repoDir = Split-Path -Parent (Split-Path -Parent $scriptDir)
Set-Location $repoDir
# Detect pre-existing stacks by exact container names. The test stack shares the
# compose project name, so `docker compose ps` would otherwise see kg-test-*
# containers and falsely report dev/personal as running.
$runningContainers = @(docker ps --format "{{.Names}}" 2>$null)
$devWasRunning = $runningContainers -contains "kg-backend"
$personalWasRunning = $runningContainers -contains "kg-backend-personal"

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$snapshotBase = Join-Path (Join-Path $scriptDir "temp") "snapshots"
$snapshotDir = Join-Path $snapshotBase $timestamp
New-Item -ItemType Directory -Path $snapshotDir -Force | Out-Null

# Phase tracking: every step registers its name and exit code.
# Shared implementation lives in lib/phase-tracking.ps1 so the regression
# test exercises the same code path as this script.
. "$scriptDir\lib\phase-tracking.ps1"
$script:SnapshotDir = $snapshotDir

function Stop-TestStack {
    Write-Host "`nStopping test stack (if still running)..." -ForegroundColor Yellow
    $restoreDir = if ($PWD.Path) { $PWD.Path } else { $scriptDir }
    try {
        Set-Location $repoDir
        & $scriptDir\stop-test.ps1 | Out-Null
        Write-Host "  ✓ Test stack stopped" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Could not stop test stack automatically" -ForegroundColor Yellow
    } finally {
        Set-Location $restoreDir
    }
}

function Restore-Stacks {
    Write-Host "`nRestoring previously running dev and personal stacks..." -ForegroundColor Yellow
    $restoreDir = if ($PWD.Path) { $PWD.Path } else { $scriptDir }
    try {
        Set-Location $repoDir
        # Build dev/personal images only if they are missing (e.g., after full cleanup)
        if ($devWasRunning) {
            $devBackendImage = docker images --format '{{.Repository}}:{{.Tag}}' | Select-String -Pattern '^knowledge-graph-backend:latest$' -Quiet
            if ($devBackendImage) {
                docker compose up -d --wait | Out-Null
            } else {
                docker compose up -d --build --wait | Out-Null
            }
        }
        if ($personalWasRunning) {
            $personalBackendImage = docker images --format '{{.Repository}}:{{.Tag}}' | Select-String -Pattern '^knowledge-graph-backend_personal:latest$' -Quiet
            if ($personalBackendImage) {
                docker compose -f docker-compose.personal.yml up -d --wait | Out-Null
            } else {
                docker compose -f docker-compose.personal.yml up -d --build --wait | Out-Null
            }
        }
        Write-Host "  ✓ Previously running stacks restored" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Could not restore all stacks automatically" -ForegroundColor Yellow
    } finally {
        Set-Location $restoreDir
    }
}

try {
    # Step 0: Capture dev stack state snapshot
    Write-Host "[Step 0/28] Capturing dev stack state snapshot..." -ForegroundColor Yellow
    docker ps --format "{{.Names}}" | Where-Object { $_ -like "kg-*" -and $_ -notlike "kg-test-*" } | Sort-Object > "$snapshotDir\pre-test-ps.txt"
    Write-Host "  ✓ Container names snapshot saved to $snapshotDir\pre-test-ps.txt" -ForegroundColor Green

    try {
        Invoke-RestMethod -Uri "http://127.0.0.1:18080/health" -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File "$snapshotDir\pre-test-health.json"
        Write-Host "  ✓ Health snapshot saved to $snapshotDir\pre-test-health.json" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Dev health endpoint not available (stack may be stopped)" -ForegroundColor Yellow
    }

    try {
        Invoke-RestMethod -Uri "http://127.0.0.1:18080/api/v1/notes?limit=1" -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File "$snapshotDir\pre-test-notes.json"
        Write-Host "  ✓ Notes snapshot saved to $snapshotDir\pre-test-notes.json" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Dev API not available (stack may be stopped)" -ForegroundColor Yellow
    }

    # Step 1: Stop dev stack
    Write-Host "`n[Step 1/28] Stopping dev stack..." -ForegroundColor Yellow
    docker compose down
    if ($LASTEXITCODE -ne 0) { Write-Host "  ⚠ docker compose down returned $LASTEXITCODE" -ForegroundColor Yellow }
    Write-Host "  ✓ Dev stack stopped" -ForegroundColor Green

    # Step 2: Stop personal stack
    Write-Host "`n[Step 2/28] Stopping personal stack..." -ForegroundColor Yellow
    docker compose -f docker-compose.personal.yml down
    if ($LASTEXITCODE -ne 0) { Write-Host "  ⚠ docker compose down returned $LASTEXITCODE" -ForegroundColor Yellow }
    Write-Host "  ✓ Personal stack stopped" -ForegroundColor Green

    # Step 3: Check stacks identity
    Write-Host "`n[Step 3/28] Checking stacks identity..." -ForegroundColor Yellow
    & $scriptDir\..\ci\check-stacks-identity.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Stacks have differences" -ForegroundColor Red
        Write-Host "Please fix the differences before running tests" -ForegroundColor Red
        Write-Host "⚠ Continuing with isolated testing despite identity differences" -ForegroundColor Yellow
    } else {
        Write-Host "✓ Stacks are identical" -ForegroundColor Green
    }

    # Step 4: Start test stack
    Write-Host "`n[Step 4/28] Starting test stack..." -ForegroundColor Yellow
    # SKIP_AUTH is needed only while docker compose builds/starts the stack;
    # leaving it set globally leaks into `go test` and breaks config tests.
    $previousSkipAuth = $env:SKIP_AUTH
    $env:SKIP_AUTH = "true"
    try {
        & $scriptDir\start-test.ps1
        $startTestStackExit = $LASTEXITCODE
    } finally {
        if ($null -ne $previousSkipAuth) {
            $env:SKIP_AUTH = $previousSkipAuth
        } else {
            Remove-Item Env:SKIP_AUTH -ErrorAction SilentlyContinue
        }
    }
    Register-Phase -Name "Start test stack" -ExitCode $startTestStackExit
    if ($startTestStackExit -ne 0) {
        throw "Failed to start test stack"
    }
    Write-Host "✓ Test stack started" -ForegroundColor Green

    # Step 5: Seed test data
    Write-Host "`n[Step 5/28] Seeding test data..." -ForegroundColor Yellow
    & $scriptDir\seed-test-data.ps1
    $seedTestDataExit = $LASTEXITCODE
    Register-Phase -Name "Seed test data" -ExitCode $seedTestDataExit
    if ($seedTestDataExit -ne 0) {
        Write-Host "  WARNING: Continuing anyway (data might already exist)" -ForegroundColor Yellow
    }

    # Step 6: Docker Build Verification
    Write-Host "`n[Step 6/28] Docker Build Verification..." -ForegroundColor Yellow
    Write-Host "  Checking Docker images..." -ForegroundColor Yellow
    docker images --format "{{.Repository}}:{{.Tag}}" | Select-String -Pattern "^knowledge-graph" | Out-Host
    Write-Host "  ✓ Docker images checked" -ForegroundColor Green

    # Step 7: NLP Service Tests
    Write-Host "`n[Step 7/28] NLP Service Tests..." -ForegroundColor Yellow
    try {
        $null = Invoke-RestMethod -Uri "http://127.0.0.1:15002/health" -Method Get -TimeoutSec 5
        Write-Host "  ✓ NLP health endpoint: OK" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ NLP health endpoint: FAILED" -ForegroundColor Yellow
    }

    # Step 8: Backend Unit Tests
    Write-Host "`n[Step 8/28] Backend Unit Tests..." -ForegroundColor Yellow
    Write-Host "  Running backend unit tests..." -ForegroundColor Yellow
    Set-Location $repoDir\backend
    # Run packages sequentially to reduce concurrent testcontainers load on Docker Desktop
    go test -p=1 -count=1 ./...
    $backendTestExit = $LASTEXITCODE
    Set-Location $repoDir
    Register-Phase -Name "Backend unit tests" -ExitCode $backendTestExit
    if ($backendTestExit -ne 0) {
        throw "Backend unit tests failed"
    }
    Write-Host "  ✓ Backend unit tests completed" -ForegroundColor Green

    # Step 9: Backend Integration Tests
    Write-Host "`n[Step 9/28] Backend Integration Tests..." -ForegroundColor Yellow
    Write-Host "  Running backend integration tests (requires Linux/WSL Docker)..." -ForegroundColor Yellow
    Set-Location $repoDir\backend
    go test -tags=integration -p=1 -count=1 ./...
    $backendIntegrationExit = $LASTEXITCODE
    Set-Location $repoDir
    Register-Phase -Name "Backend integration tests" -ExitCode $backendIntegrationExit
    if ($backendIntegrationExit -ne 0) {
        Write-Host "  WARNING: Backend integration tests failed (exit code $backendIntegrationExit); continuing" -ForegroundColor Yellow
    }


    # Step 10: Backend API Verification
    Write-Host "`n[Step 10/28] Backend API Verification..." -ForegroundColor Yellow
    try {
        $null = Invoke-RestMethod -Uri "http://127.0.0.1:18083/health" -Method Get -TimeoutSec 5
        Write-Host "  ✓ Test backend health: OK" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Test backend health: FAILED" -ForegroundColor Red
    }

    try {
        $null = Invoke-RestMethod -Uri "http://127.0.0.1:18083/api/v1/notes?limit=1" -Method Get -TimeoutSec 5
        Write-Host "  ✓ Test backend API: OK" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Test backend API: FAILED" -ForegroundColor Red
    }

    # Step 11: Asynchronous Tasks Verification
    Write-Host "`n[Step 11/28] Asynchronous Tasks Verification..." -ForegroundColor Yellow
    docker logs kg-test-worker --tail 10 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ⚠ Worker logs not available" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ Worker logs checked" -ForegroundColor Green
    }

    # Step 12: PGVECTOR Verification
    Write-Host "`n[Step 12/28] PGVECTOR Verification..." -ForegroundColor Yellow
    docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';" | Out-Null
    $pgvectorExit = $LASTEXITCODE
    Register-Phase -Name "PGVECTOR verification" -ExitCode $pgvectorExit
    if ($pgvectorExit -ne 0) {
        throw "PGVECTOR verification failed"
    }
    Write-Host "  ✓ PGVECTOR extension checked" -ForegroundColor Green

    # Step 13: Redis & MongoDB Verification
    Write-Host "`n[Step 13/28] Redis & MongoDB Verification..." -ForegroundColor Yellow
    docker exec kg-test-redis redis-cli PING | Out-Null
    $redisExit = $LASTEXITCODE
    docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')" | Out-Null
    $mongoExit = $LASTEXITCODE
    $redisMongoExit = if ($redisExit -ne 0 -or $mongoExit -ne 0) { 1 } else { 0 }
    Register-Phase -Name "Redis and MongoDB verification" -ExitCode $redisMongoExit
    if ($redisMongoExit -ne 0) {
        throw "Redis and/or MongoDB verification failed"
    }
    Write-Host "  ✓ Redis and MongoDB checked" -ForegroundColor Green

    # Step 14: Frontend Unit Tests + Coverage
    Write-Host "`n[Step 14/28] Frontend Unit Tests + Coverage..." -ForegroundColor Yellow
    Write-Host "  Running frontend unit tests with coverage (npm run test:coverage)..." -ForegroundColor Yellow
    Write-Host "  Tip: coverage report can be regenerated anytime with: cd frontend; npm run test:coverage" -ForegroundColor Cyan
    Set-Location $repoDir\frontend
    npm run test:coverage
    $frontendTestExit = $LASTEXITCODE
    Set-Location $repoDir
    Register-Phase -Name "Frontend unit tests + coverage" -ExitCode $frontendTestExit
    if ($frontendTestExit -ne 0) {
        throw "Frontend unit tests or coverage check failed"
    }
    Write-Host "  ✓ Frontend unit tests + coverage completed" -ForegroundColor Green

    # Step 15: E2E/BDD Phase 1 — SKIP_AUTH=true test stack
    Write-Host "`n[Step 15/28] E2E/BDD Phase 1 — SKIP_AUTH=true test stack..." -ForegroundColor Yellow
    $env:FRONTEND_URL = $frontendUrl
    $env:BACKEND_URL = "http://127.0.0.1:18083"
    $env:SKIP_AUTH = "true"
    Set-Location $repoDir\frontend
    npx playwright test --project=chromium-skip-auth
    $e2eSkipAuthExit = $LASTEXITCODE
    Register-Phase -Name "E2E SKIP_AUTH" -ExitCode $e2eSkipAuthExit
    if ($e2eSkipAuthExit -ne 0) {
        Write-Host "  WARNING: SKIP_AUTH E2E tests failed (exit $e2eSkipAuthExit); continuing" -ForegroundColor Yellow
    }

    node scripts/run-bdd.cjs
    $bddSkipAuthExit = $LASTEXITCODE
    Register-Phase -Name "BDD SKIP_AUTH" -ExitCode $bddSkipAuthExit
    if ($bddSkipAuthExit -ne 0) {
        Write-Host "  WARNING: SKIP_AUTH BDD tests failed (exit $bddSkipAuthExit); continuing" -ForegroundColor Yellow
    }
    Set-Location $repoDir

    # Step 16: E2E Phase 2 — real auth (SKIP_AUTH=false) test stack
    Write-Host "`n[Step 16/28] E2E Phase 2 — real auth (SKIP_AUTH=false) test stack..." -ForegroundColor Yellow
    Write-Host "  Stopping SKIP_AUTH test stack and restarting with SKIP_AUTH=false..." -ForegroundColor Yellow
    Set-Location $repoDir
    & $scriptDir\stop-test.ps1
    $env:SKIP_AUTH = "false"
    $env:VITE_SKIP_AUTH = "false"
    # Rebuild the frontend so VITE_SKIP_AUTH is baked into the bundle as "false".
    docker compose -f docker-compose.test.yml up -d --build --wait
    $realAuthStackExit = $LASTEXITCODE
    Register-Phase -Name "Real-auth test stack start" -ExitCode $realAuthStackExit
    if ($realAuthStackExit -ne 0) {
        throw "Real auth test stack failed to start"
    }
    Write-Host "  Test stack started with SKIP_AUTH=false" -ForegroundColor Green

    # Seed test data for the real-auth test user so @manual/@canvas tests have graph nodes
    Write-Host "  Seeding real-auth test data..." -ForegroundColor Yellow
    & $scriptDir\seed-test-data.ps1 -NoteCount 50 -LinkCount 20
    $seedRealAuthExit = $LASTEXITCODE
    Register-Phase -Name "Seed real-auth test data" -ExitCode $seedRealAuthExit
    if ($seedRealAuthExit -ne 0) {
        Write-Host "  WARNING: Failed to seed real-auth test data; continuing" -ForegroundColor Yellow
    }

    Set-Location $repoDir\frontend
    npx playwright test --project=chromium-real-auth
    $e2eRealAuthExit = $LASTEXITCODE
    Register-Phase -Name "E2E real auth" -ExitCode $e2eRealAuthExit
    if ($e2eRealAuthExit -ne 0) {
        Write-Host "  WARNING: Real auth E2E tests failed (exit $e2eRealAuthExit); continuing" -ForegroundColor Yellow
    }
    Set-Location $repoDir

    # Step 17: Visual Regression (Argos)
    Write-Host "`n[Step 17/28] Visual Regression (Argos)..." -ForegroundColor Yellow

    # Load token from gitignored argos.json if the env variable is not set.
    # Playwright config also loads this file, but the check above needs it here.
    if (-not $env:ARGOS_TOKEN) {
        $argosJsonPath = Join-Path (Join-Path $repoDir "frontend") "argos.json"
        if (Test-Path $argosJsonPath) {
            try {
                $argosJson = Get-Content $argosJsonPath -Raw | ConvertFrom-Json
                if ($argosJson.token) {
                    $env:ARGOS_TOKEN = $argosJson.token
                    Write-Host "  Argos token loaded from frontend/argos.json" -ForegroundColor Cyan
                }
            } catch {
                Write-Host "  Could not parse frontend/argos.json" -ForegroundColor Yellow
            }
        }
    }

    if ($env:ARGOS_TOKEN -or $env:ARGOS_UPLOAD_LOCAL) {
        # Visual regression expects a SKIP_AUTH test stack so the UI can render
        # graph/search state without logging in. Rebuild the frontend with
        # VITE_SKIP_AUTH=true and re-seed with a small, deterministic data set.
        Write-Host "  Preparing SKIP_AUTH test stack for visual regression..." -ForegroundColor Cyan
        Set-Location $repoDir
        & $scriptDir\stop-test.ps1

        $env:SKIP_AUTH = "true"
        $env:VITE_SKIP_AUTH = "true"
        docker compose -f docker-compose.test.yml up -d --build --wait
        $visualStackExit = $LASTEXITCODE
        Register-Phase -Name "Visual test stack start" -ExitCode $visualStackExit
        if ($visualStackExit -ne 0) {
            throw "Visual test stack failed to start"
        }
        Write-Host "  Visual test stack started with SKIP_AUTH=true" -ForegroundColor Green

        Write-Host "  Seeding deterministic visual test data..." -ForegroundColor Yellow
        & $scriptDir\seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42
        $seedVisualExit = $LASTEXITCODE
        Register-Phase -Name "Seed visual test data" -ExitCode $seedVisualExit
        if ($seedVisualExit -ne 0) {
            Write-Host "  WARNING: Failed to seed visual test data; continuing" -ForegroundColor Yellow
        }

        $env:FRONTEND_URL = $frontendUrl
        $env:BACKEND_URL = "http://127.0.0.1:18083"
        $env:SKIP_AUTH = "true"
        # Local runs do not set CI, so enable explicit upload when a token is available.
        if (-not $env:CI -and $env:ARGOS_TOKEN) {
            $env:ARGOS_UPLOAD_LOCAL = "true"
        }
        Set-Location $repoDir\frontend
        npx playwright test --project=visual
        $argosExit = $LASTEXITCODE
        Register-Phase -Name "Argos visual tests" -ExitCode $argosExit
        if ($argosExit -ne 0) {
            Write-Host "  WARNING: Argos visual tests failed (exit $argosExit); continuing" -ForegroundColor Yellow
        }
        Set-Location $repoDir
    } else {
        Write-Host "  ℹ Skipping Argos visual tests (ARGOS_TOKEN or ARGOS_UPLOAD_LOCAL not set)" -ForegroundColor Cyan
        Register-Phase -Name "Argos visual tests (skipped, no token)" -ExitCode 0
    }

    # Step 18: Manual testing instructions
    Write-Host "`n[Step 18/28] Test environment ready for manual testing" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  MANUAL TESTING INSTRUCTIONS" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Test stack URLs:" -ForegroundColor Yellow
    Write-Host "  Frontend: $frontendUrl" -ForegroundColor White
    Write-Host "  Backend API: http://127.0.0.1:18083" -ForegroundColor White
    Write-Host ""
    Write-Host "Follow the manual test checklist:" -ForegroundColor Yellow
    Write-Host "  docs/MANUAL_TEST_CHECKLISTS_RU.md" -ForegroundColor White
    Write-Host ""
    Write-Host "Test user credentials:" -ForegroundColor Yellow
    Write-Host "  Login: testuser" -ForegroundColor White
    Write-Host "  Password: TestPassword123!" -ForegroundColor White
    Write-Host ""
    $manualTestFlag = "$snapshotDir\continue-manual-test.flag"
    if ($SkipManual) {
        Write-Host "  SkipManual enabled: skipping interactive manual testing" -ForegroundColor Cyan
        New-Item -ItemType File -Path $manualTestFlag -Force | Out-Null
    } else {
        Write-Host "  Manual testing in progress..." -ForegroundColor Cyan
        Write-Host "  Press Enter in the interactive terminal, or create the flag file:" -ForegroundColor Cyan
        Write-Host "  $manualTestFlag" -ForegroundColor Cyan
        Write-Host "  The script will continue automatically when the flag is created." -ForegroundColor Cyan
    }
    while (-not (Test-Path $manualTestFlag)) {
        try {
            if ([Console]::KeyAvailable -and ([Console]::ReadKey($true).Key -eq 'Enter')) {
                break
            }
        } catch {
            # Non-interactive session (e.g. background shell), continue waiting for flag
        }
        Start-Sleep -Seconds 1
    }
    if (Test-Path $manualTestFlag) {
        Remove-Item $manualTestFlag
    }

    # Step 19: Public Graph Verification
    Write-Host "`n[Step 19/28] Public Graph Verification..." -ForegroundColor Yellow
    Write-Host "  ⏳ Manual verification required" -ForegroundColor Yellow
    Write-Host "  Please verify public graph functionality manually" -ForegroundColor Yellow

    # Step 20: CI/CD Verification
    Write-Host "`n[Step 20/28] CI/CD Verification..." -ForegroundColor Yellow
    Write-Host "  ⏳ Manual verification required" -ForegroundColor Yellow
    Write-Host "  Please verify CI/CD workflows manually" -ForegroundColor Yellow

    # Step 21: Documentation Verification
    Write-Host "`n[Step 21/28] Documentation Verification..." -ForegroundColor Yellow
    $docFiles = @("docs/AGENTS.md", ".windsurfrules")
    $docFiles | ForEach-Object {
        $p = Join-Path $repoDir $_
        if (Test-Path $p) {
            Write-Host "  ✓ $_ exists" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ $_ missing" -ForegroundColor Yellow
        }
    }
    try {
        $dirty = git diff --name-only 2>$null | Where-Object { $_ -match "^(docs/AGENTS\.md|\.windsurfrules|internal/(domain|infrastructure|application|interfaces))" }
        if ($dirty) {
            Write-Host "  ⚠ Architecture files changed; verify docs/AGENTS.md and .windsurfrules are updated:" -ForegroundColor Yellow
            $dirty | ForEach-Object { Write-Host "    - $_" -ForegroundColor Yellow }
        } else {
            Write-Host "  ✓ No architecture boundary changes detected" -ForegroundColor Green
        }
    } catch {
        Write-Host "  ⚠ Could not run git diff; verify docs manually" -ForegroundColor Yellow
    }


    # Step 22: Stop test stack
    Write-Host "`n[Step 22/28] Stopping test stack..." -ForegroundColor Yellow
    & $scriptDir\stop-test.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "WARNING: Failed to stop test stack" -ForegroundColor Yellow
    } else {
        Write-Host "✓ Test stack destroyed" -ForegroundColor Green
    }

    # Step 23: Prepare for dev stack restoration
    Write-Host "`n[Step 23/28] Preparing for dev stack restoration..." -ForegroundColor Yellow
    # Cleanup moved to after state checks / auto-commit
    # (temporary files will be cleaned after final verification)


    # Step 24: Start dev stack
    Write-Host "`n[Step 24/28] Restoring dev stack state..." -ForegroundColor Yellow
    if ($devWasRunning) {
        $devBackendImage = docker images --format '{{.Repository}}:{{.Tag}}' | Select-String -Pattern '^knowledge-graph-backend:latest$' -Quiet
        if ($devBackendImage) {
            docker compose up -d --wait
        } else {
            docker compose up -d --build --wait
        }
        $devRestoreExit = $LASTEXITCODE
        Register-Phase -Name "Restore dev stack" -ExitCode $devRestoreExit
        if ($devRestoreExit -ne 0) {
            throw "Failed to restore dev stack"
        }
        Write-Host "  ✓ Dev stack restored" -ForegroundColor Green
    } else {
        Write-Host "  ✓ Dev stack remains stopped" -ForegroundColor Green
        Register-Phase -Name "Restore dev stack" -Skipped
    }

    # Step 25: Start personal stack
    Write-Host "`n[Step 25/28] Restoring personal stack state..." -ForegroundColor Yellow
    if ($personalWasRunning) {
        $personalBackendImage = docker images --format '{{.Repository}}:{{.Tag}}' | Select-String -Pattern '^knowledge-graph-backend_personal:latest$' -Quiet
        if ($personalBackendImage) {
            docker compose -f docker-compose.personal.yml up -d --wait
        } else {
            docker compose -f docker-compose.personal.yml up -d --build --wait
        }
        $personalRestoreExit = $LASTEXITCODE
        Register-Phase -Name "Restore personal stack" -ExitCode $personalRestoreExit
        if ($personalRestoreExit -ne 0) {
            throw "Failed to restore personal stack"
        }
        Write-Host "  ✓ Personal stack restored" -ForegroundColor Green
    } else {
        Write-Host "  ✓ Personal stack remains stopped" -ForegroundColor Green
        Register-Phase -Name "Restore personal stack" -Skipped
    }

    # Step 26: State, identity and health checks
    Write-Host "`n[Step 26/28] State, identity and health checks" -ForegroundColor Yellow
    $devStateChanged = $false
    New-Item -ItemType Directory -Path $snapshotDir -Force | Out-Null
    docker ps --format "{{.Names}}" | Where-Object { $_ -like "kg-*" -and $_ -notlike "kg-test-*" } | Sort-Object > "$snapshotDir\post-test-ps.txt"
    Write-Host "  ✓ Post-test container names snapshot saved" -ForegroundColor Green

    if ($devWasRunning) {
        try {
            Invoke-RestMethod -Uri "http://127.0.0.1:18080/health" -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File "$snapshotDir\post-test-health.json"
            Write-Host "  ✓ Post-test health snapshot saved" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠ Dev health endpoint not available after restoration" -ForegroundColor Red
            Register-Phase -Name "Dev stack restoration (health endpoint)" -ExitCode 1
            throw "Dev stack restoration failed (health endpoint)"
        }

        try {
            Invoke-RestMethod -Uri "http://127.0.0.1:18080/api/v1/notes?limit=1" -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File "$snapshotDir\post-test-notes.json"
            Write-Host "  ✓ Post-test notes snapshot saved" -ForegroundColor Green
        } catch {
            try {
                Invoke-RestMethod -Uri "http://127.0.0.1:18080/api/v1/graph/all?limit=1" -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File "$snapshotDir\post-test-notes.json"
                Write-Host "  ✓ Post-test public graph snapshot saved (notes endpoint requires auth)" -ForegroundColor Green
            } catch {
                Write-Host "  ⚠ Dev API not available after restoration" -ForegroundColor Yellow
                $devStateChanged = $true
            }
        }
    }

    # Only compare the set of running containers pre- and post-test.
    # Health and API responses intentionally differ (timestamps, real data),
    # so we only verify that the endpoints are reachable.
    $prePs = "$snapshotDir\pre-test-ps.txt"
    $postPs = "$snapshotDir\post-test-ps.txt"
    if ((Test-Path $prePs) -and (Test-Path $postPs)) {
        # Wrap in @(...) so empty files produce empty arrays, not $null.
        if (Compare-Object (@(Get-Content $prePs)) (@(Get-Content $postPs))) {
            Write-Host "  ⚠ Dev container state changed during testing" -ForegroundColor Yellow
            $devStateChanged = $true
        } else {
            Write-Host "  ✓ Dev container state unchanged" -ForegroundColor Green
        }
    }

    $restoredStacks = @()
    if ($devWasRunning) { $restoredStacks += @{ Url = "http://127.0.0.1:18080"; Name = "dev" } }
    if ($personalWasRunning) { $restoredStacks += @{ Url = "http://127.0.0.1:18082"; Name = "personal" } }
    foreach ($stack in $restoredStacks) {
        $notesUrl = "$($stack.Url)/api/v1/notes?limit=1"
        $graphUrl = "$($stack.Url)/api/v1/graph/all?limit=1"
        $output = "$snapshotDir\$($stack.Name)-notes.json"
        $tries = 0
        $reachable = $false
        while ($tries -lt 3 -and -not $reachable) {
            try {
                $null = Invoke-RestMethod -Uri $notesUrl -Method Get -TimeoutSec 5
                Write-Host "  ✓ $($stack.Name) API is reachable" -ForegroundColor Green
                $reachable = $true
            } catch {
                try {
                    Start-Sleep -Seconds 2
                    $null = Invoke-RestMethod -Uri $graphUrl -Method Get -TimeoutSec 5
                    Write-Host "  ✓ $($stack.Name) public graph is reachable (notes endpoint requires auth)" -ForegroundColor Green
                    $reachable = $true
                } catch {
                    $tries++
                    if ($tries -ge 3) {
                        Write-Host "  ⚠ Could not reach $($stack.Name) API or public graph after retries" -ForegroundColor Yellow
                        $devStateChanged = $true
                    } else {
                        Start-Sleep -Seconds 3
                    }
                }
            }
        }
        if ($reachable) {
            Invoke-RestMethod -Uri $graphUrl -Method Get -TimeoutSec 5 | ConvertTo-Json | Out-File $output
        } else {
            "{}" | Out-File $output
        }
    }

    if ($devStateChanged) {
        Register-Phase -Name "Dev stack state verification" -ExitCode 1
        Write-Host "  WARNING: Dev stack state changed during testing" -ForegroundColor Yellow
        Write-Host "  This may indicate data leakage or side effects" -ForegroundColor Yellow
    } else {
        Register-Phase -Name "Dev stack state verification" -ExitCode 0
        Write-Host "  ✓ Dev stack state verified - no unexpected changes detected" -ForegroundColor Green
    }

    # Step 23: (continued)

    if ($devWasRunning) {
        & $scriptDir\..\ci\check-stacks-health.ps1 -Stack dev
        $devHealthExit = $LASTEXITCODE
        Register-Phase -Name "Dev stack health check" -ExitCode $devHealthExit
        if ($devHealthExit -ne 0) {
            throw "Dev stack is not healthy"
        }
    } else {
        Register-Phase -Name "Dev stack health check" -Skipped
    }

    if ($personalWasRunning) {
        & $scriptDir\..\ci\check-stacks-health.ps1 -Stack personal
        $personalHealthExit = $LASTEXITCODE
        Register-Phase -Name "Personal stack health check" -ExitCode $personalHealthExit
        if ($personalHealthExit -ne 0) {
            throw "Personal stack is not healthy"
        }
    } else {
        Register-Phase -Name "Personal stack health check" -Skipped
    }
    Write-Host "  ✓ Restored stacks health verified" -ForegroundColor Green

    # Step 27: Cleanup Temporary Files
    Write-Host "`n[Step 27/28] Cleanup Temporary Files..." -ForegroundColor Yellow
    python "$scriptDir\..\cleanup\cleanup-test-artifacts.py"
    $cleanupExit = $LASTEXITCODE
    Register-Phase -Name "Cleanup temporary files" -ExitCode $cleanupExit
    if ($cleanupExit -ne 0) {
        throw "Failed to clean temporary files"
    }

    $anyFailed = Test-AnyFailed
    Write-FinalSummary -Success:(-not $anyFailed)
    if ($anyFailed) {
        exit 1
    }
} catch {
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-FinalSummary -Success:$false
    exit 1
} finally {
    Stop-TestStack
    Restore-Stacks
}
