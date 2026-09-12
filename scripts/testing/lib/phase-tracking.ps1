# Shared phase tracking for the full test cycle scripts.
# Dot-source this file before using Register-Phase / Write-FinalSummary / Test-AnyFailed.
#
# PhaseResults is a script-scoped ordered map:
#   Name -> @{ Status = 'pass' | 'fail' | 'skip'; ExitCode = int }
#
# The main script sets $script:SnapshotDir so Write-FinalSummary can report the
# snapshot location without each call site passing it.

if (-not (Get-Variable -Name PhaseResults -Scope Script -ErrorAction SilentlyContinue)) {
    $script:PhaseResults = [ordered]@{}
}

function Register-Phase {
    param(
        [Parameter(Mandatory)]
        [string]$Name,

        [int]$ExitCode = 0,

        [switch]$Skipped,

        [string]$Reason = ""
    )

    if ($Skipped -and $ExitCode -eq 0) {
        $script:PhaseResults[$Name] = @{
            Status   = 'skip'
            ExitCode = $ExitCode
            Reason   = $Reason
        }
        $reasonSuffix = if ($Reason) { " — $Reason" } else { "" }
        Write-Host "  [SKIP] $Name$reasonSuffix" -ForegroundColor Yellow
        return
    }

    $script:PhaseResults[$Name] = @{
        Status   = if ($ExitCode -eq 0) { 'pass' } else { 'fail' }
        ExitCode = $ExitCode
    }

    if ($ExitCode -ne 0) {
        Write-Host "  [FAIL] $Name (exit $ExitCode)" -ForegroundColor Red
    } else {
        Write-Host "  [PASS] $Name" -ForegroundColor Green
    }
}

function Write-FinalSummary {
    param([bool]$Success)

    Write-Host "`n[Final Summary] Test cycle summary" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    $skipped = @($script:PhaseResults.GetEnumerator() | Where-Object { $_.Value.Status -eq 'skip' })
    if ($Success -and $skipped.Count -gt 0) {
        Write-Host "  TEST CYCLE COMPLETE WITH SKIPS" -ForegroundColor Yellow
    } elseif ($Success) {
        Write-Host "  TEST CYCLE COMPLETE" -ForegroundColor Green
    } else {
        Write-Host "  TEST CYCLE FAILED" -ForegroundColor Red
    }
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""

    foreach ($entry in $script:PhaseResults.GetEnumerator()) {
        switch ($entry.Value.Status) {
            'skip'  {
                $reasonSuffix = if ($entry.Value.Reason) { " — $($entry.Value.Reason)" } else { "" }
                Write-Host "  [SKIP] $($entry.Key)$reasonSuffix" -ForegroundColor Yellow
            }
            'pass'  { Write-Host "  [PASS] $($entry.Key)" -ForegroundColor Green }
            'fail'  { Write-Host "  [FAIL] $($entry.Key) (exit $($entry.Value.ExitCode))" -ForegroundColor Red }
            default { Write-Host "  [????] $($entry.Key) (status: $($entry.Value.Status))" -ForegroundColor Magenta }
        }
    }
    Write-Host ""

    if ($script:SnapshotDir) {
        Write-Host "Snapshots saved to: $script:SnapshotDir" -ForegroundColor Yellow
        Write-Host ""
    }

    if ($skipped.Count -gt 0) {
        Write-Host "Skipped phases: $($skipped.Count)" -ForegroundColor Yellow
        foreach ($entry in $skipped) {
            $reason = if ($entry.Value.Reason) { $entry.Value.Reason } else { "no reason provided" }
            Write-Host "  - $($entry.Key): $reason" -ForegroundColor Yellow
        }
        Write-Host ""
    }

    if ($Success -and $skipped.Count -eq 0) {
        Write-Host "All stacks are stable and isolated testing completed successfully." -ForegroundColor Green
    } elseif ($Success) {
        Write-Host "No phases failed, but skipped checks require review." -ForegroundColor Yellow
    } else {
        Write-Host "One or more phases failed. See details above." -ForegroundColor Red
    }
}

function Test-AnyFailed {
    # Any phase registered as a hard failure (not a skip) fails the cycle.
    return @($script:PhaseResults.Values | Where-Object { $_.Status -eq 'fail' }).Count -gt 0
}
