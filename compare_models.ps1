# Сравнение папок с моделью
$folder1 = "D:\knowledge-graph\nlp-service\local_model_cache\all-MiniLM-L6-V2"
$folder2 = "D:\knowledge-graph\huggingface_cache"

Write-Host "=== Сравнение папок с моделью ===" -ForegroundColor Cyan
Write-Host "Папка 1: $folder1 (размер: $(Get-ChildItem $folder1 -Recurse -File | Measure-Object -Property Length -Sum | ForEach-Object { [math]::Round($_.Sum / 1MB, 2) }) MB)"
Write-Host "Папка 2: $folder2 (размер: $(Get-ChildItem $folder2 -Recurse -File | Measure-Object -Property Length -Sum | ForEach-Object { [math]::Round($_.Sum / 1MB, 2) }) MB)"
Write-Host ""

# Функция для получения относительных путей
function Get-RelativePaths {
    param($root, $files)
    $rootLen = $root.Length + 1
    $files | ForEach-Object { $_.FullName.Substring($rootLen) }
}

# Получаем списки файлов (рекурсивно)
$files1 = Get-ChildItem $folder1 -Recurse -File
$files2 = Get-ChildItem $folder2 -Recurse -File

$relPaths1 = Get-RelativePaths -root $folder1 -files $files1
$relPaths2 = Get-RelativePaths -root $folder2 -files $files2

# Файлы только в первой папке
$onlyIn1 = Compare-Object $relPaths1 $relPaths2 | Where-Object { $_.SideIndicator -eq '<=' } | Select-Object -ExpandProperty InputObject
if ($onlyIn1) {
    Write-Host "Файлы ТОЛЬКО в папке 1:" -ForegroundColor Yellow
    $onlyIn1 | ForEach-Object { Write-Host "  $_" }
} else {
    Write-Host "Нет файлов, уникальных для папки 1." -ForegroundColor Green
}

# Файлы только во второй папке
$onlyIn2 = Compare-Object $relPaths1 $relPaths2 | Where-Object { $_.SideIndicator -eq '=>' } | Select-Object -ExpandProperty InputObject
if ($onlyIn2) {
    Write-Host "`nФайлы ТОЛЬКО в папке 2:" -ForegroundColor Yellow
    $onlyIn2 | ForEach-Object { Write-Host "  $_" }
} else {
    Write-Host "`nНет файлов, уникальных для папки 2." -ForegroundColor Green
}

# Сравнение файлов с одинаковыми именами (размеры)
Write-Host "`nФайлы с одинаковыми именами, но разными размерами:" -ForegroundColor Cyan
$commonFiles = $relPaths1 | Where-Object { $relPaths2 -contains $_ }
$differences = @()
foreach ($rel in $commonFiles) {
    $file1 = $files1 | Where-Object { $_.FullName.EndsWith($rel) } | Select-Object -First 1
    $file2 = $files2 | Where-Object { $_.FullName.EndsWith($rel) } | Select-Object -First 1
    if ($file1.Length -ne $file2.Length) {
        $differences += [PSCustomObject]@{
            Path = $rel
            Size1 = $file1.Length
            Size2 = $file2.Length
        }
    }
}
if ($differences) {
    $differences | ForEach-Object { Write-Host "  $($_.Path) : $($_.Size1) B vs $($_.Size2) B" }
} else {
    Write-Host "  Все общие файлы имеют одинаковый размер." -ForegroundColor Green
}

# Дополнительно: хэши для ключевых файлов (model.safetensors)
Write-Host "`nХэши ключевых файлов (MD5):" -ForegroundColor Cyan
$keyFiles = @("model.safetensors", "pytorch_model.bin", "config.json")
foreach ($keyFile in $keyFiles) {
    $file1 = Get-ChildItem $folder1 -Recurse -Filter $keyFile | Select-Object -First 1
    $file2 = Get-ChildItem $folder2 -Recurse -Filter $keyFile | Select-Object -First 1
    if ($file1 -and $file2) {
        $hash1 = (Get-FileHash $file1.FullName -Algorithm MD5).Hash
        $hash2 = (Get-FileHash $file2.FullName -Algorithm MD5).Hash
        if ($hash1 -eq $hash2) {
            Write-Host "  $keyFile : одинаковые хэши" -ForegroundColor Green
        } else {
            Write-Host "  $keyFile : РАЗНЫЕ хэши" -ForegroundColor Red
            Write-Host "    Папка 1: $hash1"
            Write-Host "    Папка 2: $hash2"
        }
    } elseif ($file1) {
        Write-Host "  $keyFile : только в папке 1"
    } elseif ($file2) {
        Write-Host "  $keyFile : только в папке 2"
    } else {
        Write-Host "  $keyFile : не найден"
    }
}

Write-Host "`n=== Анализ завершён ===" -ForegroundColor Cyan