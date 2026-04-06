# inspect_docs.ps1
$docsPath = Join-Path (Get-Location) "docs"
if (-not (Test-Path $docsPath)) {
    Write-Host "Папка docs не найдена." -ForegroundColor Red
    exit 1
}

Write-Host "=== Анализ документации в папке $docsPath ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "=== Структура папок (до глубины 3) ===" -ForegroundColor Green
Get-ChildItem $docsPath -Directory -Recurse | ForEach-Object {
    $relative = $_.FullName.Substring($docsPath.Length + 1)
    $depth = ($relative -split '\\').Count
    if ($depth -le 3) {
        $indent = "  " * ($depth - 1)
        Write-Host "$indent📁 $relative"
    }
}
Write-Host ""

Write-Host "=== Все файлы в docs (с размерами и датами) ===" -ForegroundColor Green
$files = Get-ChildItem $docsPath -File -Recurse | Sort-Object FullName
foreach ($file in $files) {
    $relative = $file.FullName.Substring($docsPath.Length + 1)
    $size = "{0:N2} KB" -f ($file.Length / 1KB)
    $modified = $file.LastWriteTime.ToString("yyyy-MM-dd HH:mm")
    Write-Host "$relative | $size | $modified"
}
Write-Host ""

Write-Host "=== Статистика по .md файлам (строки) ===" -ForegroundColor Green
Get-ChildItem $docsPath -Recurse -Filter *.md | ForEach-Object {
    $lines = (Get-Content $_.FullName | Measure-Object -Line).Lines
    $relative = $_.FullName.Substring($docsPath.Length + 1)
    Write-Host "$relative : $lines строк"
}
Write-Host ""

Write-Host "=== Проверка битых ссылок в .md файлах ===" -ForegroundColor Green
$broken = $false
Get-ChildItem $docsPath -Recurse -Filter *.md | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $regex = '\[.*?\]\((.*?)\)'
    $matches = [regex]::Matches($content, $regex)
    foreach ($match in $matches) {
        $link = $match.Groups[1].Value
        if ($link -match '^https?://|^#|^mailto:|^$/') { continue }
        $target = Join-Path (Split-Path $_.FullName -Parent) $link
        if (-not (Test-Path $target)) {
            Write-Host "Битая ссылка в $($_.FullName): $link" -ForegroundColor Red
            $broken = $true
        }
    }
}
if (-not $broken) { Write-Host "Битых ссылок не найдено." -ForegroundColor Green }
Write-Host ""

Write-Host "=== Наличие ключевых файлов ===" -ForegroundColor Green
$required = @(
    "architecture/README.md",
    "architecture/glossary.md",
    "architecture/adr.md",
    "architecture/atam.md",
    "architecture/c4/context.puml",
    "architecture/c4/container.puml",
    "architecture/c4/component.puml",
    "architecture/uml/deployment-local.puml",
    "architecture/uml/deployment-k8s.puml",
    "architecture/uml/class-domain.puml",
    "architecture/uml/er-diagram.puml",
    "architecture/uml/sequence-create-note.puml",
    "architecture/uml/sequence-suggestions.puml",
    "CONFIGURATION.md"
)
foreach ($file in $required) {
    $full = Join-Path $docsPath $file
    if (Test-Path $full) {
        Write-Host "✅ $file" -ForegroundColor Green
    } else {
        Write-Host "❌ $file" -ForegroundColor Red
    }
}
Write-Host ""
Write-Host "=== Анализ завершён ===" -ForegroundColor Cyan