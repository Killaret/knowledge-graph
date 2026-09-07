#!/usr/bin/env python3
"""Regression tests for the Personal stack destructive-command guard.

Run: python scripts/devops/test_guard_personal_data.py
"""

from __future__ import annotations

import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import time
from pathlib import Path

GUARD = Path(__file__).with_name("guard-personal-data.py")

spec = importlib.util.spec_from_file_location("guard_personal_data", GUARD)
guard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(guard)

# (command, expected to be treated as destructive to Personal data)
COMMAND_CASES: list[tuple[str, bool]] = [
    # Test and dev stack work must never be blocked.
    ("docker compose -f docker-compose.test.yml down -v", False),
    ("docker compose -f docker-compose.test.yml down --volumes", False),
    ("docker compose -f docker-compose.personal.yml down", False),
    ("docker volume rm kg_test_pgdata", False),
    ("cd backend && go test ./...", False),
    ("npm run test:unit", False),
    ("rm -rf frontend/node_modules", False),
    # Quoting a dangerous command is not running it.
    ("grep -rn personal docker-compose.personal.yml", False),
    ('git commit -m "fix: docker volume prune && down -v в тексте"', False),
    ('git commit -m "docker compose -f docker-compose.personal.yml down -v"', False),
    ("echo 'docker volume prune' > notes.txt", False),
    # Genuinely destructive to Personal data.
    ("docker compose -f docker-compose.personal.yml down -v", True),
    ("docker compose -f docker-compose.personal.yml down --volumes", True),
    ("docker compose -p knowledge-graph-personal down -v", True),
    ("docker volume prune -f", True),
    ("docker system prune -a", True),
    ("docker volume rm pgdata_personal", True),
    ("docker volume rm redisdata_personal mongodbdata_personal", True),
    ("rm -rf ./backups/personal-old", True),
    ("cd /d/knowledge-graph && docker volume prune", True),
    ("DOCKER_HOST=tcp://x docker volume prune", True),
    # A chain is caught even when only the tail is destructive.
    ("docker compose -f docker-compose.personal.yml down -v && docker volume prune", True),
    ("npm run test:unit && docker volume prune", True),
    # A shell wrapper is not a bypass.
    ('bash -c "docker volume prune"', True),
    ('pwsh -Command "docker volume rm pgdata_personal"', True),
    ('bash -c "npm run test:unit"', False),
]


def check_commands() -> int:
    failures = 0
    for command, expected in COMMAND_CASES:
        actual = guard.targets_personal_data(command)
        if actual != expected:
            failures += 1
            print(f"FAIL targets_personal_data expected={expected} actual={actual} | {command}")
    return failures


def check_newest_backup() -> int:
    """newest_backup must ignore empty files and pick the newest real one."""
    failures = 0
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        backups = root / "backups"
        backups.mkdir()

        found, _, _ = guard.newest_backup(root)
        if found is not None:
            failures += 1
            print(f"FAIL newest_backup on empty dir returned {found}")

        (backups / "backup-personal-2026-01-01.sql.gz").write_bytes(b"")
        found, _, _ = guard.newest_backup(root)
        if found is not None:
            failures += 1
            print(f"FAIL newest_backup counted a zero-byte file: {found}")

        old = backups / "backup-personal-2026-01-02.sql.gz"
        old.write_bytes(b"old")
        os.utime(old, (time.time() - 7200, time.time() - 7200))
        new = backups / "personal-volumes-raw-2026-01-03.tar.gz"
        new.write_bytes(b"new")

        found, _, size = guard.newest_backup(root)
        if found != new:
            failures += 1
            print(f"FAIL newest_backup picked {found}, expected {new.name}")
        if size != 3:
            failures += 1
            print(f"FAIL newest_backup reported size {size}, expected 3")
    return failures


def run_guard(cwd: Path, command: str, env_extra: dict[str, str] | None = None) -> dict:
    """Run the guard as the hook does and return its parsed output."""
    env = {**os.environ, **(env_extra or {})}
    payload = json.dumps({"tool_name": "Bash", "tool_input": {"command": command}})
    result = subprocess.run(
        [sys.executable, str(GUARD)],
        input=payload,
        capture_output=True,
        text=True,
        cwd=cwd,
        env=env,
    )
    out = result.stdout.strip()
    return json.loads(out) if out else {}


def decision_of(output: dict) -> str:
    hook = output.get("hookSpecificOutput", {})
    if hook.get("permissionDecision") == "deny":
        return "deny"
    if "systemMessage" in output:
        return "allow"
    return "pass"


def check_backup_gate() -> int:
    """The deny branches of main() are the safety-critical path."""
    failures = 0
    destructive = "docker volume prune -f"

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        subprocess.run(["git", "init", "-q"], cwd=root, capture_output=True)

        got = decision_of(run_guard(root, destructive))
        if got != "deny":
            failures += 1
            print(f"FAIL no backup at all: expected deny, got {got}")

        backups = root / "backups"
        backups.mkdir()
        (backups / "backup-personal-empty.sql.gz").write_bytes(b"")
        got = decision_of(run_guard(root, destructive))
        if got != "deny":
            failures += 1
            print(f"FAIL only an empty backup: expected deny, got {got}")

        stale = backups / "backup-personal-stale.sql.gz"
        stale.write_bytes(b"data")
        old = time.time() - 72 * 3600
        os.utime(stale, (old, old))
        got = decision_of(run_guard(root, destructive))
        if got != "deny":
            failures += 1
            print(f"FAIL stale backup: expected deny, got {got}")

        # The same stale backup passes once the threshold is raised.
        got = decision_of(run_guard(root, destructive, {"KG_BACKUP_MAX_AGE_HOURS": "999"}))
        if got != "allow":
            failures += 1
            print(f"FAIL raised threshold: expected allow, got {got}")

        fresh = backups / "backup-personal-fresh.sql.gz"
        fresh.write_bytes(b"data")
        got = decision_of(run_guard(root, destructive))
        if got != "allow":
            failures += 1
            print(f"FAIL fresh backup: expected allow, got {got}")

        # A harmless command is never gated, backup or not.
        got = decision_of(run_guard(root, "npm run test:unit"))
        if got != "pass":
            failures += 1
            print(f"FAIL harmless command: expected pass, got {got}")
    return failures


def main() -> int:
    failures = check_commands() + check_newest_backup() + check_backup_gate()
    total = len(COMMAND_CASES) + 4 + 6
    if failures:
        print(f"\n{failures} checks failed")
        return 1
    print(f"All checks passed ({total} assertions)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
