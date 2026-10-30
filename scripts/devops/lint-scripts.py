#!/usr/bin/env python3
"""Cross-platform linter dispatcher for repository scripts.

Runs PSScriptAnalyzer on PowerShell scripts when available and shellcheck on Bash
scripts when available.
"""

import os
import shutil
import subprocess
import sys
from pathlib import Path


def project_root() -> Path:
    return Path(__file__).resolve().parents[2]


def run_powershell_lint(scripts_dir: Path) -> int:
    """Run PSScriptAnalyzer if available."""
    try:
        subprocess.run(
            ["pwsh", "-Command", "Get-Command Invoke-ScriptAnalyzer"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("PSScriptAnalyzer not available; skipping PowerShell lint")
        return 0

    print("Running PSScriptAnalyzer on scripts/*.ps1")
    result = subprocess.run(
        [
            "pwsh",
            "-Command",
            f"Invoke-ScriptAnalyzer -Path '{scripts_dir}' -Recurse -Severity Error",
        ],
        cwd=project_root(),
    )
    return result.returncode


def run_shell_lint(scripts_dir: Path) -> int:
    """Run shellcheck on Bash scripts if available."""
    shellcheck = shutil.which("shellcheck")  # noqa: F821
    if not shellcheck:
        print("shellcheck not available; skipping Bash lint")
        return 0

    sh_files = list(scripts_dir.rglob("*.sh"))
    if not sh_files:
        print("No Bash scripts found")
        return 0

    print(f"Running shellcheck on {len(sh_files)} Bash scripts")
    result = subprocess.run([shellcheck] + [str(f) for f in sh_files], cwd=project_root())
    return result.returncode


def main() -> int:
    scripts_dir = project_root() / "scripts"
    ps_rc = run_powershell_lint(scripts_dir)
    sh_rc = run_shell_lint(scripts_dir)
    return max(ps_rc, sh_rc)


if __name__ == "__main__":
    sys.exit(main())
