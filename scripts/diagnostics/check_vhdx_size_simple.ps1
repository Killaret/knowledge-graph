$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'
if (Test-Path $vhdxFile) {
    $size = (Get-Item $vhdxFile).Length / 1GB
    Write-Output "Current VHDX size: $($size.ToString('F2')) GB"
} else {
    Write-Output "File not found: $vhdxFile"
}