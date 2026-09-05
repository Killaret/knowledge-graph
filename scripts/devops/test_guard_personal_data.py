#!/usr/bin/env python3
"""Regression tests for the Personal stack destructive-command guard.

Run: python scripts/devops/test_guard_personal_data.py
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

GUARD = Path(__file__).with_name("guard-personal-data.py")

spec = importlib.util.spec_from_file_location("guard_personal_data", GUARD)
guard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(guard)

# (command, expected to be treated as destructive to Personal data)
CASES: list[tuple[str, bool]] = [
    # Test and dev stack work must never be blocked.
    ("docker compose -f docker-compose.test.yml down -v", False),
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
    ("docker volume prune -f", True),
    ("docker system prune -a", True),
    ("docker volume rm pgdata_personal", True),
    ("rm -rf ./backups/personal-old", True),
    ("cd /d/knowledge-graph && docker volume prune", True),
    ("DOCKER_HOST=tcp://x docker volume prune", True),
]


def main() -> int:
    failures = 0
    for command, expected in CASES:
        actual = guard.targets_personal_data(command)
        if actual != expected:
            failures += 1
            print(f"FAIL expected={expected} actual={actual} | {command}")

    if failures:
        print(f"\n{failures} of {len(CASES)} cases failed")
        return 1
    print(f"All {len(CASES)} cases passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
