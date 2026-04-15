# Create test notes with different types
$types = @(
    @{ title = '⭐ Star Note'; content = 'Star type note for testing'; type = 'star' },
    @{ title = '🪐 Planet Note'; content = 'Planet type note for testing'; type = 'planet' },
    @{ title = '☄️ Comet Note'; content = 'Comet type note for testing'; type = 'comet' },
    @{ title = '🌀 Galaxy Note'; content = 'Galaxy type note for testing'; type = 'galaxy' },
    @{ title = '📄 Default Note'; content = 'Default type note for testing'; type = 'default' },
    @{ title = '⭐ Another Star'; content = 'Second star note'; type = 'star' },
    @{ title = '🪐 Another Planet'; content = 'Second planet note'; type = 'planet' },
    @{ title = '☄️ Another Comet'; content = 'Second comet note'; type = 'comet' }
)

$created = @()
foreach ($n in $types) {
    $json = $n | ConvertTo-Json -Compress
    try {
        $r = Invoke-RestMethod -Uri 'http://localhost:8080/notes' -Method POST -ContentType 'application/json' -Body $json
        Write-Host "Created: $($r.title) - $($r.id)"
        $created += $r
    } catch {
        Write-Host "Error creating note: $_" -ForegroundColor Red
    }
}

# Create links between consecutive notes
Write-Host "`nCreating links..."
for ($i = 0; $i -lt $created.Count - 1; $i++) {
    $body = @{
        source_note_id = $created[$i].id
        target_note_id = $created[$i + 1].id
        link_type = 'related'
        weight = 0.8
    } | ConvertTo-Json -Compress
    
    try {
        Invoke-RestMethod -Uri 'http://localhost:8080/links' -Method POST -ContentType 'application/json' -Body $body
        Write-Host "Link: $($created[$i].title) -> $($created[$i + 1].title)"
    } catch {
        Write-Host "Link error: $_" -ForegroundColor Red
    }
}

# Create cross-links
Write-Host "`nCreating cross-links..."
$cross = @(
    @{ s = 0; t = 3 },
    @{ s = 1; t = 4 },
    @{ s = 2; t = 5 },
    @{ s = 3; t = 6 }
)

foreach ($l in $cross) {
    if ($l.t -lt $created.Count) {
        $body = @{
            source_note_id = $created[$l.s].id
            target_note_id = $created[$l.t].id
            link_type = 'reference'
            weight = 0.6
        } | ConvertTo-Json -Compress
        
        try {
            Invoke-RestMethod -Uri 'http://localhost:8080/links' -Method POST -ContentType 'application/json' -Body $body
            Write-Host "Cross-link: $($created[$l.s].title) -> $($created[$l.t].title)"
        } catch {}
    }
}

Write-Host "`nDone! Created $($created.Count) notes with various types." -ForegroundColor Green
