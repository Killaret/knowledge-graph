#!/bin/bash
# Setup branch protection rules for knowledge-graph repository
# Requires: gh CLI installed and authenticated

REPO="Killaret/knowledge-graph"

echo "Setting up branch protection for $REPO..."

# 1. Protect main branch - only Killaret can push, requires PR and review
echo "Setting up protection for main branch..."
gh api repos/$REPO/branches/main/protection --method PUT \
  --input - <<< '{
    "required_status_checks": {
      "strict": true,
      "contexts": ["Backend Tests", "Frontend Tests", "NLP Service Tests", "Integration Tests", "Security Scan"]
    },
    "enforce_admins": true,
    "required_pull_request_reviews": {
      "required_approving_review_count": 1,
      "dismiss_stale_reviews": true,
      "require_code_owner_reviews": false,
      "require_last_push_approval": true
    },
    "restrictions": {
      "users": ["Killaret"],
      "teams": []
    },
    "allow_force_pushes": false,
    "allow_deletions": false,
    "required_conversation_resolution": true,
    "require_signed_commits": false
  }'

# 2. Protect windsurf branch - only Killaret + status checks
echo "Setting up protection for windsurf branch..."
gh api repos/$REPO/branches/windsurf/protection --method PUT \
  --input - <<< '{
    "required_status_checks": {
      "strict": true,
      "contexts": ["Backend Tests", "Frontend Tests", "NLP Service Tests", "Integration Tests", "Security Scan"]
    },
    "enforce_admins": false,
    "required_pull_request_reviews": null,
    "restrictions": {
      "users": ["Killaret"],
      "teams": []
    },
    "allow_force_pushes": false,
    "allow_deletions": false,
    "required_conversation_resolution": false
  }'

# 3. Protect sourceTextHandler branch - only alximac + status checks
echo "Setting up protection for sourceTextHandler branch..."
gh api repos/$REPO/branches/sourceTextHandler/protection --method PUT \
  --input - <<< '{
    "required_status_checks": {
      "strict": true,
      "contexts": ["Backend Tests", "Frontend Tests", "NLP Service Tests", "Integration Tests", "Security Scan"]
    },
    "enforce_admins": false,
    "required_pull_request_reviews": null,
    "restrictions": {
      "users": ["alximac"],
      "teams": []
    },
    "allow_force_pushes": false,
    "allow_deletions": false,
    "required_conversation_resolution": false
  }'

echo "Branch protection setup complete!"
echo ""
echo "Summary:"
echo "- main: Only Killaret, requires PR + 1 review + status checks"
echo "- windsurf: Only Killaret + status checks"
echo "- sourceTextHandler: Only alximac + status checks"
