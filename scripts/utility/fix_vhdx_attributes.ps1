$vhdxFile = 'C:\Users\89209\AppData\Local\Docker\wsl\disk\docker_data.vhdx'

Write-Output "Removing problematic attributes from VHDX file..."
Write-Output "File: $vhdxFile"
Write-Output ""

# Remove sparse attribute
Write-Output "Removing SparseFile attribute..."
fsutil sparse setflag $vhdxFile 0

# Remove compression
Write-Output "Removing Compressed attribute..."
compact /U $vhdxFile

Write-Output ""
Write-Output "Verifying attributes after cleanup..."
$file = Get-Item $vhdxFile
Write-Output "Current attributes: $($file.Attributes)"

# Check sparse status again
Write-Output ""
Write-Output "Sparse status check:"
fsutil sparse queryflag $vhdxFile

Write-Output ""
Write-Output "Attributes removed successfully!"
Write-Output "Now you can try diskpart commands again."