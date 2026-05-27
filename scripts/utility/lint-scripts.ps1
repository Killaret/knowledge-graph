# Simple lint for PowerShell scripts using PSScriptAnalyzer if installed
if (Get-Command Invoke-ScriptAnalyzer -ErrorAction SilentlyContinue) {
    Write-Host "Running PSScriptAnalyzer on scripts/*.ps1"
    Invoke-ScriptAnalyzer -Path "$PSScriptRoot\*.ps1" -Recurse -Severity Error
} else {
    Write-Host "PSScriptAnalyzer not found; skipping PowerShell lint"
}
