Write-Output "=== Checking what is blocking the VHDX file ==="
Write-Output ""

$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'

Write-Output "Checking processes using the VHDX file..."
try {
    $processes = Get-Process | ForEach-Object {
        $process = $_
        try {
            $handles = (Get-Process -Id $process.Id -Module -ErrorAction SilentlyContinue).FileName
            if ($handles -like "*vhdx*" -or $handles -like "*docker*") {
                Write-Output "Process: $($process.ProcessName) (ID: $($process.Id))"
            }
        } catch {}
    }
} catch {
    Write-Output "Error checking processes: $_"
}

Write-Output ""
Write-Output "Checking file lock status..."
try {
    $fileInfo = New-Object System.IO.FileInfo($vhdxFile)
    $stream = $fileInfo.Open([System.IO.FileMode]::Open, [System.IO.FileAccess]::ReadWrite, [System.IO.FileShare]::None)
    $stream.Close()
    Write-Output "File is NOT locked"
} catch {
    Write-Output "File is LOCKED by another process"
    Write-Output "Error: $_"
}

Write-Output ""
Write-Output "Checking WSL status..."
try {
    wsl --status
} catch {
    Write-Output "WSL command failed: $_"
}

Write-Output ""
Write-Output "Checking running WSL distributions..."
try {
    wsl --list --running
} catch {
    Write-Output "WSL list command failed: $_"
}

Write-Output ""
Write-Output "Checking Docker processes..."
Get-Process | Where-Object { $_.ProcessName -like "*docker*" -or $_.ProcessName -like "*com.docker*" -or $_.ProcessName -like "*wsl*" } | Select-Object ProcessName, Id