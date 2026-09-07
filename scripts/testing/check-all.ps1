param(
    [switch]$Quick
)

$ErrorActionPreference = 'Continue'
$scriptDir = $PSScriptRoot
$repoDir = Split-Path -Parent (Split-Path -Parent $scriptDir)
$manifestPath = Join-Path $scriptDir 'core-checks.tsv'
$workflowPath = Join-Path $repoDir '.github\workflows\_core-checks.yml'
$syncScript = Join-Path $scriptDir 'check-core-workflow-sync.mjs'

. "$scriptDir\lib\phase-tracking.ps1"
$script:PhaseResults.Clear()
$script:SnapshotDir = $null
$checks = Import-Csv -Path $manifestPath -Delimiter "`t"
$coverCreated = $false

function Register-Skip {
    param([string]$Name, [string]$Reason)
    Register-Phase -Name $Name -Skipped -Reason $Reason
}

function Test-ToolAvailable {
    param([string]$Tool)

    foreach ($name in $Tool.Split('+')) {
        if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
            return "$name is not installed or not in PATH"
        }
    }
    if ($Tool -like 'docker*') {
        docker info *> $null
        if ($LASTEXITCODE -ne 0) {
            return 'Docker daemon is unavailable'
        }
    }
    return $null
}

function Invoke-Check {
    param($Check)

    $originalLocation = Get-Location
    $previousJwtSecret = $env:JWT_SECRET
    try {
        Set-Location (Join-Path $repoDir $Check.cwd)
        if ($Check.id -eq 'backend-coverage') {
            if (-not (Test-Path 'cover.out')) {
                Write-Host 'cover.out was not produced by backend tests' -ForegroundColor Red
                return 1
            }
            $coverageOutput = go tool cover '-func=cover.out'
            if ($LASTEXITCODE -ne 0) { return $LASTEXITCODE }
            $totalLine = $coverageOutput | Where-Object { $_ -match '^total:' } | Select-Object -Last 1
            if (-not $totalLine -or $totalLine -notmatch '([0-9]+(?:\.[0-9]+)?)%') { return 1 }
            $actual = [double]$Matches[1]
            $required = [double]($Check.command -replace '^@coverage:', '')
            Write-Host "Backend coverage: $actual% (required >= $required%)"
            return $(if ($actual -ge $required) { 0 } else { 1 })
        }
        if ($Check.id -eq 'frontend-config') {
            npm run build-config | Out-Host
            if ($LASTEXITCODE -ne 0) { return [int]$LASTEXITCODE }
            git diff --exit-code -- knowledge-graph.config.json | Out-Host
            return [int]$LASTEXITCODE
        }
        if ($Check.id -eq 'backend-config' -and -not $env:JWT_SECRET) {
            $env:JWT_SECRET = 'local-check-secret'
        }
        $global:LASTEXITCODE = 0
        $command = [string]$Check.command
        Invoke-Expression $command | Out-Host
        return [int]$(if ($null -eq $LASTEXITCODE) { 0 } else { $LASTEXITCODE })
    } catch {
        Write-Host $_ -ForegroundColor Red
        return 1
    } finally {
        if ($null -ne $previousJwtSecret) {
            $env:JWT_SECRET = $previousJwtSecret
        } else {
            Remove-Item Env:JWT_SECRET -ErrorAction SilentlyContinue
        }
        Set-Location $originalLocation
    }
}

Write-Host '========================================' -ForegroundColor Cyan
Write-Host '  Knowledge Graph Local Core Checks' -ForegroundColor Cyan
Write-Host '========================================' -ForegroundColor Cyan

if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Register-Skip -Name 'Core workflow sync' -Reason 'node is not installed or not in PATH'
} else {
    node $syncScript $manifestPath $workflowPath
    Register-Phase -Name 'Core workflow sync' -ExitCode $LASTEXITCODE
}

foreach ($check in $checks) {
    if ($Quick -and $check.quick_skip -eq '1') {
        Register-Skip -Name $check.name -Reason 'quick mode skips integration checks'
        continue
    }
    $unavailableReason = Test-ToolAvailable -Tool $check.tool
    if ($unavailableReason) {
        Register-Skip -Name $check.name -Reason $unavailableReason
        continue
    }
    $exitCode = Invoke-Check -Check $check
    if ($check.id -eq 'backend-unit') { $coverCreated = $true }
    Register-Phase -Name $check.name -ExitCode $exitCode
}

if ($coverCreated) {
    Remove-Item (Join-Path $repoDir 'backend\cover.out') -ErrorAction SilentlyContinue
}

$failed = Test-AnyFailed
Write-FinalSummary -Success (-not $failed)
exit $(if ($failed) { 1 } else { 0 })
