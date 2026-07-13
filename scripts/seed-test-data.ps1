# Seed Test Data - Windows PowerShell
# This script registers a test user and creates test notes and links

$apiUrl = "http://localhost:8083/api/v1"

Write-Host "Seeding test data..." -ForegroundColor Cyan

# Test user credentials
$testUser = @{
    login = "testuser"
    email = "testuser@example.com"
    password = "TestPassword123!"
}

# Register test user
Write-Host "Registering test user..." -ForegroundColor Yellow
try {
    $registerResponse = Invoke-RestMethod -Uri "$apiUrl/auth/register" -Method Post -Body ($testUser | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Test user registered successfully" -ForegroundColor Green
} catch {
    Write-Host "User might already exist or registration failed: $_" -ForegroundColor Yellow
}

# Login to get token
Write-Host "Logging in test user..." -ForegroundColor Yellow
try {
    $loginResponse = Invoke-RestMethod -Uri "$apiUrl/auth/login" -Method Post -Body (@{login = $testUser.login; password = $testUser.password} | ConvertTo-Json) -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "Login successful" -ForegroundColor Green
} catch {
    Write-Host "Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    Authorization = "Bearer $token"
}

# Create test notes
Write-Host "Creating test notes..." -ForegroundColor Yellow
$noteTypes = @("star", "planet", "comet", "galaxy", "asteroid")
$noteIds = @()

foreach ($type in $noteTypes) {
    $note = @{
        title = "Test $type Note"
        content = "This is a test note of type $type"
        type = $type
    }
    
    try {
        $noteResponse = Invoke-RestMethod -Uri "$apiUrl/notes" -Method Post -Body ($note | ConvertTo-Json) -ContentType "application/json" -Headers $headers
        $noteIds += $noteResponse.data.id
        Write-Host "Created note: $($note.title) (ID: $($noteResponse.data.id))" -ForegroundColor Green
    } catch {
        Write-Host "Failed to create note: $_" -ForegroundColor Red
    }
}

# Create test links
Write-Host "Creating test links..." -ForegroundColor Yellow
if ($noteIds.Count -ge 2) {
    $link = @{
        source_note_id = $noteIds[0]
        target_note_id = $noteIds[1]
        link_type = "related"
        weight = 1.0
    }
    
    try {
        $linkResponse = Invoke-RestMethod -Uri "$apiUrl/links" -Method Post -Body ($link | ConvertTo-Json) -ContentType "application/json" -Headers $headers
        Write-Host "Created link between $($noteIds[0]) and $($noteIds[1])" -ForegroundColor Green
    } catch {
        Write-Host "Failed to create link: $_" -ForegroundColor Red
    }
    
    # Create second link
    if ($noteIds.Count -ge 3) {
        $link2 = @{
            source_note_id = $noteIds[1]
            target_note_id = $noteIds[2]
            link_type = "dependency"
            weight = 0.5
        }
        
        try {
            $linkResponse2 = Invoke-RestMethod -Uri "$apiUrl/links" -Method Post -Body ($link2 | ConvertTo-Json) -ContentType "application/json" -Headers $headers
            Write-Host "Created link between $($noteIds[1]) and $($noteIds[2])" -ForegroundColor Green
        } catch {
            Write-Host "Failed to create second link: $_" -ForegroundColor Red
        }
    }
}

Write-Host "`nTest data seeded successfully!" -ForegroundColor Green
Write-Host "Created $($noteIds.Count) notes and 2 links" -ForegroundColor Green
