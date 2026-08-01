# Seed Test Data - Windows PowerShell
# Creates a known set of test notes and links for manual/automated testing.
# Requires the test stack to be running at localhost:18083.

param(
    [int]$NoteCount = 100,
    [int]$LinkCount = 60,
    [int]$NlpWaitSeconds = 600,
    [int]$Seed = 0,
    [string]$ReportPath = "$PSScriptRoot\seed-report.json",
    # Percentage of created notes to publish (make public) so the anonymous/public
    # graph view has data to render in real-auth mode. Set to 0 to keep all notes
    # private (previous behavior).
    [int]$PublicPercent = 20
)

if ($Seed -gt 0) {
    Get-Random -SetSeed $Seed | Out-Null
}

$apiUrl = "http://127.0.0.1:18083/api/v1"
$postgresContainer = "kg-test-postgres"

$testUser = @{
    login = "testuser"
    email = "testuser@example.com"
    password = "TestPassword123!"
}

# All valid note types in the backend
$noteTypes = @(
    'star', 'planet', 'comet', 'galaxy', 'asteroid',
    'satellite', 'debris', 'nebula', 'dust', 'unknown', 'blackhole'
)

$linkTypes = @('reference', 'dependency', 'related', 'custom')

$report = @{
    startedAt = (Get-Date -Format o)
    noteCount = 0
    linkCount = 0
    publicNoteCount = 0
    embeddingCount = 0
    keywordNoteCount = 0
    noteIds = @()
    linkIds = @()
    typeDistribution = @{
    }
    errors = @()
    durations = @{}
}

function Add-ReportError {
    param([string]$Context, [string]$Message)
    $report.errors += @{ context = $Context; message = $Message }
    Write-Host "[$Context] ERROR: $Message" -ForegroundColor Red
}

# ---------------------------------------------------------------------------
# 1. Clean existing test data (only the kg-test-postgres container)
# ---------------------------------------------------------------------------
$cleanStart = Get-Date
Write-Host "Cleaning existing test data..." -ForegroundColor Cyan

try {
    $cleanOutput = docker exec $postgresContainer psql `
        -U kb_user -d knowledge_test -t -A -c `
        "TRUNCATE TABLE notes, note_embeddings, note_keywords, links CASCADE;" 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw $cleanOutput
    }
    Write-Host "Existing test data truncated." -ForegroundColor Green
}
catch {
    Add-ReportError -Context "clean" -Message "Failed to truncate test data: $_"
    exit 1
}

$report.durations.cleanSeconds = [math]::Round(((Get-Date) - $cleanStart).TotalSeconds, 2)

# ---------------------------------------------------------------------------
# 2. Register / login test user
# ---------------------------------------------------------------------------
$authStart = Get-Date
Write-Host "Registering test user..." -ForegroundColor Yellow

try {
    Invoke-RestMethod -Uri "$apiUrl/auth/register" -Method Post `
        -Body ($testUser | ConvertTo-Json) -ContentType "application/json" | Out-Null
    Write-Host "Test user registered." -ForegroundColor Green
}
catch {
    Write-Host "Registration skipped or user already exists: $_" -ForegroundColor Yellow
}

Write-Host "Logging in test user..." -ForegroundColor Yellow

try {
    $loginResponse = Invoke-RestMethod -Uri "$apiUrl/auth/login" -Method Post `
        -Body (@{login = $testUser.login; password = $testUser.password} | ConvertTo-Json) `
        -ContentType "application/json"
    $token = $loginResponse.access_token
    if (-not $token) {
        throw "No access_token in login response. Response keys: $((($loginResponse | Get-Member -MemberType NoteProperty).Name -join ', '))"
    }
    Write-Host "Login successful." -ForegroundColor Green
}
catch {
    Add-ReportError -Context "login" -Message "Login failed: $_"
    exit 1
}

$headers = @{
    Authorization = "Bearer $token"
}

$report.durations.authSeconds = [math]::Round(((Get-Date) - $authStart).TotalSeconds, 2)

# ---------------------------------------------------------------------------
# 3. Create notes
# ---------------------------------------------------------------------------
$notesStart = Get-Date
Write-Host "Creating $NoteCount test notes..." -ForegroundColor Yellow

$noteIds = [System.Collections.ArrayList]::new()
$createdNotes = [System.Collections.ArrayList]::new()

function Get-NoteContent {
    param([string]$Type, [int]$Index)
    return "This is a test $Type note number $Index. It contains enough text for NLP processing to extract keywords and compute embeddings. " +
           "Knowledge graph helps explore the cosmos of ideas."
}

for ($i = 0; $i -lt $NoteCount; $i++) {
    $type = $noteTypes[$i % $noteTypes.Count]
    $title = "Seed $type $(($i + 1).ToString('000'))"
    $content = Get-NoteContent -Type $type -Index $i

    $note = @{
        title = $title
        content = $content
        type = $type
    }

    try {
        $noteResponse = Invoke-RestMethod -Uri "$apiUrl/notes" -Method Post `
            -Body ($note | ConvertTo-Json) -ContentType "application/json" -Headers $headers
        [void]$noteIds.Add($noteResponse.data.id)
        [void]$createdNotes.Add($noteResponse.data)

        if (-not $report.typeDistribution.ContainsKey($type)) {
            $report.typeDistribution[$type] = 0
        }
        $report.typeDistribution[$type] += 1

        if (($i + 1) % 10 -eq 0) {
            Write-Host "  Created $(($i + 1)) notes..." -ForegroundColor DarkGray
        }
    }
    catch {
        Add-ReportError -Context "create-note" -Message "Failed to create note $i ($type): $_"
    }
}

$report.noteCount = $noteIds.Count
$report.noteIds = $noteIds
$report.durations.createNotesSeconds = [math]::Round(((Get-Date) - $notesStart).TotalSeconds, 2)

Write-Host "Created $($noteIds.Count) notes." -ForegroundColor Green

# ---------------------------------------------------------------------------
# 3.5. Publish a subset of notes so the anonymous/public graph view has data
#      (previously all seeded notes stayed private, leaving the public graph
#      empty in real-auth mode — see docs/MANUAL_TEST_ISSUES.md).
# ---------------------------------------------------------------------------
$publishStart = Get-Date
$publicCount = [int][math]::Ceiling($noteIds.Count * ($PublicPercent / 100))

if ($publicCount -gt 0) {
    Write-Host "Publishing $publicCount of $($noteIds.Count) notes ($PublicPercent%) for the public graph..." -ForegroundColor Yellow
    $publishedCount = 0

    for ($i = 0; $i -lt $publicCount; $i++) {
        try {
            Invoke-RestMethod -Uri "$apiUrl/notes/$($noteIds[$i])/publish" -Method Post `
                -Headers $headers | Out-Null
            $publishedCount++
        }
        catch {
            Add-ReportError -Context "publish-note" -Message "Failed to publish note $($noteIds[$i]): $_"
        }
    }

    $report.publicNoteCount = $publishedCount
    Write-Host "Published $publishedCount notes." -ForegroundColor Green
}
else {
    Write-Host "PublicPercent is 0 — all notes remain private." -ForegroundColor DarkGray
}

$report.durations.publishSeconds = [math]::Round(((Get-Date) - $publishStart).TotalSeconds, 2)

# ---------------------------------------------------------------------------
# 4. Wait for NLP processing (embeddings + keywords)
# ---------------------------------------------------------------------------
$nlpStart = Get-Date
Write-Host "Waiting for NLP service to process notes (embeddings + keywords)..." -ForegroundColor Yellow

$embeddingCount = 0
$keywordNoteCount = 0
$elapsed = 0

while ($elapsed -lt $NlpWaitSeconds) {
    try {
        $embeddingCount = [int]((docker exec $postgresContainer psql -U kb_user -d knowledge_test -t -A `
            -c "SELECT COUNT(*) FROM note_embeddings;" 2>&1) -join " ").Trim()
        $keywordNoteCount = [int]((docker exec $postgresContainer psql -U kb_user -d knowledge_test -t -A `
            -c "SELECT COUNT(DISTINCT note_id) FROM note_keywords;" 2>&1) -join " ").Trim()
    }
    catch {
        Add-ReportError -Context "nlp-poll" -Message "Failed to poll NLP progress: $_"
    }

    Write-Host "  NLP progress: $embeddingCount / $($noteIds.Count) embeddings, $keywordNoteCount / $($noteIds.Count) keyword notes..." -ForegroundColor DarkGray

    if ($embeddingCount -ge $noteIds.Count -and $keywordNoteCount -ge $noteIds.Count) {
        break
    }

    Start-Sleep -Seconds 5
    $elapsed += 5
}

$report.embeddingCount = $embeddingCount
$report.keywordNoteCount = $keywordNoteCount
$report.durations.nlpWaitSeconds = [math]::Round(((Get-Date) - $nlpStart).TotalSeconds, 2)

if ($embeddingCount -lt $noteIds.Count -or $keywordNoteCount -lt $noteIds.Count) {
    Add-ReportError -Context "nlp-wait" -Message "NLP processing did not complete within $NlpWaitSeconds seconds. Embeddings: $embeddingCount/$($noteIds.Count), Keywords: $keywordNoteCount/$($noteIds.Count)"
}
else {
    Write-Host "NLP processing complete." -ForegroundColor Green
}

# ---------------------------------------------------------------------------
# 5. Create links
# ---------------------------------------------------------------------------
$linksStart = Get-Date
Write-Host "Creating $LinkCount test links..." -ForegroundColor Yellow

$linkIds = [System.Collections.ArrayList]::new()
$createdPairs = [System.Collections.Generic.HashSet[string]]::new()
$attempts = 0
$maxAttempts = $LinkCount * 5

while ($linkIds.Count -lt $LinkCount -and $attempts -lt $maxAttempts) {
    $attempts++

    if ($noteIds.Count -lt 2) {
        Add-ReportError -Context "create-link" -Message "Not enough notes to create links."
        break
    }

    $sourceIndex = Get-Random -Minimum 0 -Maximum $noteIds.Count
    $targetIndex = Get-Random -Minimum 0 -Maximum $noteIds.Count

    if ($sourceIndex -eq $targetIndex) {
        continue
    }

    $sourceId = $noteIds[$sourceIndex]
    $targetId = $noteIds[$targetIndex]
    $pairKey = "$sourceId-$targetId"

    if ($createdPairs.Contains($pairKey)) {
        continue
    }

    $linkType = $linkTypes[(Get-Random -Minimum 0 -Maximum $linkTypes.Count)]
    $weight = [math]::Round((Get-Random -Minimum 1 -Maximum 11) / 10, 2)

    $link = @{
        source_note_id = $sourceId
        target_note_id = $targetId
        link_type = $linkType
        weight = $weight
    }

    try {
        $linkResponse = Invoke-RestMethod -Uri "$apiUrl/links" -Method Post `
            -Body ($link | ConvertTo-Json) -ContentType "application/json" -Headers $headers
        [void]$linkIds.Add($linkResponse.data.id)
        [void]$createdPairs.Add($pairKey)

        if ($linkIds.Count % 10 -eq 0) {
            Write-Host "  Created $($linkIds.Count) links..." -ForegroundColor DarkGray
        }
    }
    catch {
        Add-ReportError -Context "create-link" -Message "Failed to create link $sourceId -> $targetId ($linkType): $_"
    }
}

$report.linkCount = $linkIds.Count
$report.linkIds = $linkIds
$report.durations.createLinksSeconds = [math]::Round(((Get-Date) - $linksStart).TotalSeconds, 2)

Write-Host "Created $($linkIds.Count) links." -ForegroundColor Green

# ---------------------------------------------------------------------------
# 6. Verify graph service
# ---------------------------------------------------------------------------
$graphStart = Get-Date
Write-Host "Verifying graph service..." -ForegroundColor Yellow

try {
    $graphResponse = Invoke-RestMethod -Uri "http://127.0.0.1:19091/api/v1/graph/full" -Method Get -Headers $headers -TimeoutSec 30
    $report.graphNodes = $graphResponse.meta.total_nodes
    $report.graphLinks = $graphResponse.meta.total_links
    Write-Host "Graph service: $($graphResponse.meta.total_nodes) nodes, $($graphResponse.meta.total_links) links." -ForegroundColor Green
}
catch {
    Add-ReportError -Context "graph-service" -Message "Graph service verification failed: $_"
}

$report.durations.graphCheckSeconds = [math]::Round(((Get-Date) - $graphStart).TotalSeconds, 2)

# ---------------------------------------------------------------------------
# 7. Save report
# ---------------------------------------------------------------------------
$report.finishedAt = (Get-Date -Format o)
$report.durations.totalSeconds = [math]::Round(
    ($report.durations.cleanSeconds + $report.durations.authSeconds + $report.durations.createNotesSeconds + $report.durations.publishSeconds + $report.durations.nlpWaitSeconds + $report.durations.createLinksSeconds + $report.durations.graphCheckSeconds),
    2
)

$report | ConvertTo-Json -Depth 10 | Out-File -FilePath $ReportPath -Encoding UTF8

Write-Host ""
Write-Host "Seed report saved to: $ReportPath" -ForegroundColor Cyan
Write-Host "Notes: $($report.noteCount) (public: $($report.publicNoteCount)) | Links: $($report.linkCount) | Embeddings: $($report.embeddingCount) | Keyword notes: $($report.keywordNoteCount) | Graph nodes: $($report.graphNodes) | Graph links: $($report.graphLinks)" -ForegroundColor Cyan
Write-Host "Total duration: $($report.durations.totalSeconds) seconds" -ForegroundColor Cyan

if ($report.errors.Count -gt 0) {
    Write-Host "There were $($report.errors.Count) errors. See report for details." -ForegroundColor Red
    exit 1
}

Write-Host "Test data seeded successfully!" -ForegroundColor Green
