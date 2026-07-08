# Seed Test Data for Knowledge Graph Test Stack
# This script populates the test database with sample data

$ErrorActionPreference = "Stop"

$baseUrl = "http://localhost:8083/api/v1"

Write-Host "Seeding test data into Knowledge Graph Test Stack..." -ForegroundColor Green

# Check if backend is available
Write-Host "Checking backend availability..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "Backend is healthy" -ForegroundColor Green
} catch {
    Write-Host "Backend is not available. Please start the test stack first." -ForegroundColor Red
    exit 1
}

# Create test user
Write-Host "Creating test user..." -ForegroundColor Yellow
try {
    $userResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -ContentType "application/json" -Body @{
        login = "testuser"
        password = "Test123!"
        name = "Test User"
    } | ConvertTo-Json
    Write-Host "Test user created successfully" -ForegroundColor Green
    $token = $userResponse.token
} catch {
    Write-Host "User might already exist or error occurred. Trying to login..." -ForegroundColor Yellow
    try {
        $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -ContentType "application/json" -Body @{
            login = "testuser"
            password = "Test123!"
        } | ConvertTo-Json
        $token = $loginResponse.token
        Write-Host "Logged in successfully" -ForegroundColor Green
    } catch {
        Write-Host "Failed to create or login test user" -ForegroundColor Red
        exit 1
    }
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# Create test notes
Write-Host "Creating test notes..." -ForegroundColor Yellow
$testNotes = @(
    @{
        title = "Test Star Note"
        content = "This is a test star note for testing purposes"
        type = "star"
    },
    @{
        title = "Test Planet Note"
        content = "This is a test planet note for testing purposes"
        type = "planet"
    },
    @{
        title = "Test Galaxy Note"
        content = "This is a test galaxy note for testing purposes"
        type = "galaxy"
    },
    @{
        title = "Test Nebula Note"
        content = "This is a test nebula note for testing purposes"
        type = "nebula"
    },
    @{
        title = "Test Black Hole Note"
        content = "This is a test black hole note for testing purposes"
        type = "blackhole"
    }
)

foreach ($note in $testNotes) {
    try {
        $noteResponse = Invoke-RestMethod -Uri "$baseUrl/notes" -Method Post -ContentType "application/json" -Headers $headers -Body ($note | ConvertTo-Json)
        Write-Host "Created note: $($note.title)" -ForegroundColor Green
    } catch {
        Write-Host "Failed to create note: $($note.title)" -ForegroundColor Red
    }
}

# Create test tags
Write-Host "Creating test tags..." -ForegroundColor Yellow
$testTags = @("test", "automation", "qa", "e2e", "sample")

foreach ($tag in $testTags) {
    try {
        $tagResponse = Invoke-RestMethod -Uri "$baseUrl/tags" -Method Post -ContentType "application/json" -Headers $headers -Body @{
            name = $tag
            color = "#FF5733"
        } | ConvertTo-Json
        Write-Host "Created tag: $tag" -ForegroundColor Green
    } catch {
        Write-Host "Failed to create tag: $tag" -ForegroundColor Yellow
    }
}

# Create test connections between notes
Write-Host "Creating test connections..." -ForegroundColor Yellow
try {
    $notes = Invoke-RestMethod -Uri "$baseUrl/notes" -Method Get -Headers $headers
    if ($notes.Count -ge 2) {
        $connection = @{
            from_note_id = $notes[0].id
            to_note_id = $notes[1].id
            relationship_type = "related"
        }
        $connResponse = Invoke-RestMethod -Uri "$baseUrl/connections" -Method Post -ContentType "application/json" -Headers $headers -Body ($connection | ConvertTo-Json)
        Write-Host "Created connection between notes" -ForegroundColor Green
    }
} catch {
    Write-Host "Failed to create connection" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Test data seeding completed!" -ForegroundColor Green
Write-Host "Created 5 test notes, 5 test tags, and sample connections" -ForegroundColor Cyan
