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
Register-Phase -Name "Argos visual tests" -Skipped -Reason "ARGOS_TOKEN is not configured"
if (Test-AnyFailed) {
    Write-Host "Skipped phase must not count as a failure" -ForegroundColor Red
    exit 1
}
if ($script:PhaseResults["Argos visual tests"].Reason -ne "ARGOS_TOKEN is not configured") {
    Write-Host "Skipped phase reason was not preserved" -ForegroundColor Red
    exit 1
}

# Scenario 4: -Skipped combined with a non-zero exit code is a failure,
# not a skip (finding 6 of the A-3 round-3 tail).
$script:PhaseResults.Clear()
Register-Phase -Name "Broken restore" -ExitCode 2 -Skipped -Reason "stack was not started"
if (-not (Test-AnyFailed)) {
    Write-Host "Skipped phase with non-zero exit code must count as a failure" -ForegroundColor Red
    exit 1
}
if ($script:PhaseResults["Broken restore"].Status -ne 'fail') {
    Write-Host "Skipped phase with non-zero exit code must be registered as fail" -ForegroundColor Red
    exit 1
}

# Scenario 5: a phase with an unexpected status must still appear in the
# summary instead of vanishing silently (finding 5).
$script:PhaseResults.Clear()
$script:PhaseResults["Weird phase"] = @{ Status = 'weird'; ExitCode = 0 }
$summary = Write-FinalSummary -Success $true 6>&1 | Out-String
if ($summary -notmatch '\[\?\?\?\?\] Weird phase') {
    Write-Host "Unexpected phase status must be printed in the summary" -ForegroundColor Red
    exit 1
}

# Scenario 6: the shell library must report phases in registration order
# (finding 2). Skipped when bash is unavailable.
if (Get-Command bash -ErrorAction SilentlyContinue) {
    # Names a1..a7 are chosen because bash's associative-array hash order
    # shuffles them (a2 a3 a1 a6 a7 a4 a5), so a summary that iterates the
    # map instead of the insertion-order index fails this check.
    $names = 'a1','a2','a3','a4','a5','a6','a7'
    $registrations = ($names | ForEach-Object { "register_phase `"$_`" 0" }) -join '; '
    $shCommand = ". scripts/testing/lib/phase-tracking.sh; $registrations; write_final_summary true"
    $shText = bash -c $shCommand 2>$null | Out-String
    # Registration prints [PASS] lines in order; the summary is what must
    # preserve order, so only look at text after the summary header.
    $summaryStart = $shText.IndexOf('[Final Summary]')
    if ($summaryStart -lt 0) {
        Write-Host "Shell summary header not found" -ForegroundColor Red
        exit 1
    }
    $summaryText = $shText.Substring($summaryStart)
    $positions = foreach ($n in $names) { $summaryText.IndexOf("[PASS] $n") }
    if (($positions | Where-Object { $_ -lt 0 }).Count -gt 0) {
        Write-Host "Shell summary did not print all registered phases" -ForegroundColor Red
        exit 1
    }
    for ($i = 1; $i -lt $positions.Count; $i++) {
        if ($positions[$i] -le $positions[$i - 1]) {
            Write-Host "Shell summary lost the registration order of phases" -ForegroundColor Red
            exit 1
        }
    }
}

Write-Host "A-3 exit-code regression test passed" -ForegroundColor Green
exit 0
