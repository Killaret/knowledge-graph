# Setup branch protection rules for knowledge-graph repository
# Requires: gh CLI installed and authenticated

$REPO = "Killaret/knowledge-graph"

Write-Host "Setting up branch protection for $REPO..." -ForegroundColor Green

# 1. Protect main branch - only Killaret can push, requires PR and review
Write-Host "Setting up protection for main branch..." -ForegroundColor Yellow
$mainProtection = @{
    required_status_checks = @{
        strict = $true
        contexts = @("Backend Tests", "Frontend Tests", "NLP Service Tests", "Integration Tests", "Security Scan")
    }
    enforce_admins = $true
    required_pull_request_reviews = @{
        required_approving_review_count = 1
        dismiss_stale_reviews = $true
        require_code_owner_reviews = $false
        require_last_push_approval = $true
    }
    restrictions = @{
        users = @("Killaret")
        teams = @()
    }
    allow_force_pushes = $false
    allow_deletions = $false
    required_conversation_resolution = $true
    require_signed_commits = $false
} | ConvertTo-Json -Depth 10

$mainProtection | gh api repos/$REPO/branches/main/protection --method PUT --input -

# 2. Protect windsurf branch - only Killaret
Write-Host "Setting up protection for windsurf branch..." -ForegroundColor Yellow
$windsurfProtection = @{
    required_status_checks = $null
    enforce_admins = $false
    required_pull_request_reviews = $null
    restrictions = @{
        users = @("Killaret")
        teams = @()
    }
    allow_force_pushes = $false
    allow_deletions = $false
} | ConvertTo-Json -Depth 10

$windsurfProtection | gh api repos/$REPO/branches/windsurf/protection --method PUT --input -

# 3. Protect sourceTextHandler branch - only alximac
Write-Host "Setting up protection for sourceTextHandler branch..." -ForegroundColor Yellow
$sourceProtection = @{
    required_status_checks = $null
    enforce_admins = $false
    required_pull_request_reviews = $null
    restrictions = @{
        users = @("alximac")
        teams = @()
    }
    allow_force_pushes = $false
    allow_deletions = $false
} | ConvertTo-Json -Depth 10

$sourceProtection | gh api repos/$REPO/branches/sourceTextHandler/protection --method PUT --input -

Write-Host "`nBranch protection setup complete!" -ForegroundColor Green
Write-Host "`nSummary:" -ForegroundColor Cyan
Write-Host "- main: Only Killaret, requires PR + 1 review + status checks" -ForegroundColor White
Write-Host "- windsurf: Only Killaret" -ForegroundColor White
Write-Host "- sourceTextHandler: Only alximac" -ForegroundColor White
