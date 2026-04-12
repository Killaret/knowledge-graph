# PowerShell script to seed test data for Knowledge Graph
$ErrorActionPreference = "Stop"

$baseUrl = "http://localhost:8080"

# Test data
$notes = @(
    @{ title = "Introduction to Machine Learning"; content = "Machine learning is a subset of artificial intelligence..."; type = "star" },
    @{ title = "Deep Learning Basics"; content = "Deep learning uses neural networks with multiple layers..."; type = "planet" },
    @{ title = "Neural Networks History"; content = "The first neural network was proposed in 1943..."; type = "comet" },
    @{ title = "Python for Data Science"; content = "Python is the most popular language for data science..."; type = "galaxy" },
    @{ title = "React vs Svelte"; content = "Both are modern frontend frameworks..."; type = "star" },
    @{ title = "Docker Containers"; content = "Docker revolutionized application deployment..."; type = "planet" },
    @{ title = "Kubernetes Guide"; content = "Kubernetes is an orchestration platform..."; type = "star" },
    @{ title = "Graph Databases"; content = "Graph databases excel at relationship queries..."; type = "galaxy" },
    @{ title = "REST API Design"; content = "RESTful APIs follow specific design principles..."; type = "comet" },
    @{ title = "Microservices Architecture"; content = "Microservices break monoliths into smaller services..."; type = "planet" }
)

Write-Host "Creating test notes..." -ForegroundColor Green
$createdNotes = @()

foreach ($note in $notes) {
    try {
        $json = $note | ConvertTo-Json -Compress
        $response = Invoke-RestMethod -Uri "$baseUrl/notes" -Method POST -ContentType "application/json" -Body $json
        Write-Host "  Created: $($note.title) (ID: $($response.id))" -ForegroundColor Gray
        $createdNotes += $response
    } catch {
        Write-Host "  Error creating $($note.title): $_" -ForegroundColor Red
    }
}

Write-Host "`nCreating links between notes..." -ForegroundColor Green

# Create some links
$links = @(
    @{ source = 0; target = 1; type = "related"; weight = 0.9 },
    @{ source = 0; target = 2; type = "reference"; weight = 0.7 },
    @{ source = 1; target = 2; type = "related"; weight = 0.8 },
    @{ source = 3; target = 0; type = "uses"; weight = 0.6 },
    @{ source = 4; target = 3; type = "compares"; weight = 0.5 },
    @{ source = 5; target = 6; type = "related"; weight = 0.9 },
    @{ source = 6; target = 7; type = "uses"; weight = 0.7 },
    @{ source = 8; target = 9; type = "related"; weight = 0.6 },
    @{ source = 0; target = 7; type = "reference"; weight = 0.5 }
)

foreach ($link in $links) {
    try {
        $sourceId = $createdNotes[$link.source].id
        $targetId = $createdNotes[$link.target].id
        
        $linkData = @{
            source_note_id = $sourceId
            target_note_id = $targetId
            link_type = $link.type
            weight = $link.weight
        } | ConvertTo-Json -Compress
        
        $response = Invoke-RestMethod -Uri "$baseUrl/links" -Method POST -ContentType "application/json" -Body $linkData
        Write-Host "  Link created: $($createdNotes[$link.source].title) -> $($createdNotes[$link.target].title)" -ForegroundColor Gray
    } catch {
        Write-Host "  Error creating link: $_" -ForegroundColor Red
    }
}

Write-Host "`nTest data creation complete!" -ForegroundColor Green
Write-Host "Created $($createdNotes.Count) notes with $($links.Count) links" -ForegroundColor Cyan
