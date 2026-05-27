$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'

try {
    $stream = [System.IO.File]::Open($vhdxFile, 'Open', 'ReadWrite', 'None')
    $stream.Close()
    Write-Output 'File is now UNLOCKED - ready for diskpart'
} catch {
    Write-Output "File is still LOCKED: $($_.Exception.Message)"
}