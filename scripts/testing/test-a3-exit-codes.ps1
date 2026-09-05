# A-3 regression test: any non-zero phase exit code must fail the cycle.
# This is a lightweight unit-style check; it does not start the test stack.
#
# The test dot-sources the same lib/phase-tracking.ps1 that the production
# run-full-test-cycle.ps1 uses, so it exercises the real Register-Phase /
# Test-AnyFailed implementation rather than a copied version.

$ErrorActionPreference = 'Stop'

. "$PSScriptRoot\lib\phase-tracking.ps1"

# Scenario 1: every phase passes -> Test-AnyFailed must be false.
$script:PhaseResults.Clear()
Register-Phase -Name "Start test stack" -ExitCode 0
Register-Phase -Name "Seed test data" -ExitCode 0
if (Test-AnyFailed) {
    Write-Host "Expected no failures when every phase exits 0" -ForegroundColor Red
    exit 1
}

# Scenario 2: one phase exits with code 2 -> Test-AnyFailed must be true.
$script:PhaseResults.Clear()
Register-Phase -Name "Start test stack" -ExitCode 0
Register-Phase -Name "Seed test data" -ExitCode 2
Register-Phase -Name "Backend unit tests" -ExitCode 0
if (-not (Test-AnyFailed)) {
    Write-Host "Expected failure to be detected for exit code 2" -ForegroundColor Red
    exit 1
}

# Scenario 3: a skipped phase must not count as a failure.
$script:PhaseResults.Clear()
Register-Phase -Name "Start test stack" -ExitCode 0
Register-Phase -Name "Argos visual tests" -Skipped
if (Test-AnyFailed) {
    Write-Host "Skipped phase must not count as a failure" -ForegroundColor Red
    exit 1
}

Write-Host "A-3 exit-code regression test passed" -ForegroundColor Green
exit 0
