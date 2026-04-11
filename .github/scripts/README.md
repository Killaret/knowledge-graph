# Branch Protection Setup Scripts

These scripts automate the setup of branch protection rules for the knowledge-graph repository.

## Prerequisites

1. Install GitHub CLI:
   ```bash
   # macOS
   brew install gh
   
   # Windows
   winget install --id GitHub.cli
   
   # Or download from https://github.com/cli/cli/releases
   ```

2. Authenticate with GitHub:
   ```bash
   gh auth login
   # Select HTTPS → Authenticate with web browser
   ```

3. Verify you have admin access to the repository

## Usage

### Linux/macOS (Bash)

```bash
cd .github/scripts
chmod +x setup-branch-protection.sh
./setup-branch-protection.sh
```

### Windows (PowerShell)

```powershell
cd .github/scripts
.\setup-branch-protection.ps1
```

Or use Git Bash to run the shell script.

## What Gets Configured

### main branch
- ✅ Only `Killaret` can push directly
- ✅ Pull requests required
- ✅ 1 approving review required
- ✅ Status checks must pass (Backend/Frontend/NLP/Integration/Security tests)
- ✅ Admin enforcement enabled
- ✅ Conversation resolution required
- ❌ Force pushes disabled
- ❌ Deletions disabled

### windsurf branch
- ✅ Only `Killaret` can push
- ❌ Force pushes disabled
- ❌ Deletions disabled

### sourceTextHandler branch
- ✅ Only `alximac` can push
- ❌ Force pushes disabled
- ❌ Deletions disabled

## Manual Setup (Alternative)

If you prefer manual setup via web interface:

1. Go to: `https://github.com/Killaret/knowledge-graph/settings/branches`
2. Click **"Add rule"**
3. Configure settings as described above
4. Repeat for each branch

## Verification

To verify protection is applied:

```bash
# Check main branch protection
gh api repos/Killaret/knowledge-graph/branches/main/protection

# Check windsurf branch protection  
gh api repos/Killaret/knowledge-graph/branches/windsurf/protection

# Check sourceTextHandler branch protection
gh api repos/Killaret/knowledge-graph/branches/sourceTextHandler/protection
```

## Troubleshooting

### "Could not resolve to a Repository"
- Ensure you're authenticated: `gh auth status`
- Run `gh auth login` if needed

### "Validation Failed" or "422 Unprocessable Entity"
- Make sure branches exist before applying protection
- Verify you have admin permissions
- Check that usernames are correct

### "Resource not accessible by integration"
- Your token needs `repo` scope
- Re-authenticate with: `gh auth login --scopes repo`

## Modifying Rules

To update rules later, edit the JSON in the scripts and re-run, or use:

```bash
# Remove protection (use with caution!)
gh api repos/Killaret/knowledge-graph/branches/BRANCH_NAME/protection --method DELETE
```

## Security Notes

- Keep these scripts in version control for transparency
- Review changes before running
- Consider adding CODEOWNERS file for additional review requirements
- Enable 2FA on your GitHub account for extra security
