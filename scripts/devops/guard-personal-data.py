#!/usr/bin/env python3
"""PreToolUse guard: refuse destructive Docker commands without a fresh backup.

Reads the Claude Code hook payload from stdin and inspects the shell command
that is about to run. Commands that can destroy Personal stack data are allowed
only when a recent, non-empty backup exists in the backups directory.

Everything else passes through untouched, including work on the test and dev
stacks, so routine `docker compose -f docker-compose.test.yml down -v` is not
affected.
"""

from __future__ import annotations

import json
import os
import re
import shlex
import subprocess
import sys
import time
from pathlib import Path

PERSONAL_VOLUMES = ("pgdata_personal", "redisdata_personal", "mongodbdata_personal")

# Commands that sweep every unused volume, personal ones included.
PRUNE_PATTERNS = (
    re.compile(r"\bdocker\s+volume\s+prune\b"),
    re.compile(r"\bdocker\s+system\s+prune\b"),
)

# Commands that destroy volumes belonging to the Personal stack specifically.
PERSONAL_PATTERNS = (
    re.compile(r"\bdocker\s+volume\s+rm\b"),
    re.compile(r"\bdown\b.*(?:\s-v\b|--volumes\b)"),
    re.compile(r"\brm\s+-[a-z]*r"),
)

BACKUP_GLOBS = ("backup-personal-*", "personal-volumes-raw-*")
MAX_AGE_HOURS = float(os.environ.get("KG_BACKUP_MAX_AGE_HOURS", "24"))


def repo_root() -> Path:
    try:
        out = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if out.returncode == 0 and out.stdout.strip():
            return Path(out.stdout.strip())
    except (OSError, subprocess.SubprocessError):
        pass
    return Path(__file__).resolve().parents[2]


ENV_ASSIGNMENT = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")
OPERATORS = {"&&", "||", ";", "|", "&", "(", ")", "\n"}
RISKY_EXECUTABLES = {"docker", "docker-compose", "rm", "rmdir", "del"}


def split_segments(command: str) -> list[list[str]]:
    """Tokenize the shell command into segments, honouring quotes.

    A quoted string never starts a new segment, so a commit message that
    mentions `docker ... && ...` is one argument, not three commands.
    """
    lexer = shlex.shlex(command, posix=True, punctuation_chars=True)
    lexer.whitespace_split = True
    segments: list[list[str]] = [[]]
    for token in lexer:
        if token in OPERATORS:
            segments.append([])
        else:
            segments[-1].append(token)
    return segments


def executable_of(segment: list[str]) -> str:
    """First real word of a segment, ignoring leading env assignments."""
    for word in segment:
        if ENV_ASSIGNMENT.match(word):
            continue
        return word.rsplit("/", 1)[-1].rsplit("\\", 1)[-1]
    return ""


def targets_personal_data(command: str) -> bool:
    """True when the command can destroy Personal stack volumes.

    Only segments that actually invoke a risky program are inspected, so a
    commit message or a doc edit that merely quotes such a command is ignored.
    A command that cannot be parsed is treated as risky.
    """
    try:
        segments = split_segments(command)
    except ValueError:
        segments = [command.split()]

    for segment in segments:
        if executable_of(segment) not in RISKY_EXECUTABLES:
            continue
        text = " ".join(segment)
        if any(p.search(text) for p in PRUNE_PATTERNS):
            return True
        mentions_personal = "personal" in text.lower() or any(
            v in text for v in PERSONAL_VOLUMES
        )
        if mentions_personal and any(p.search(text) for p in PERSONAL_PATTERNS):
            return True
    return False


def newest_backup(root: Path) -> tuple[Path | None, float, int]:
    backups = root / "backups"
    newest: Path | None = None
    newest_mtime = 0.0
    if backups.is_dir():
        for pattern in BACKUP_GLOBS:
            for candidate in backups.glob(pattern):
                if not candidate.is_file() or candidate.stat().st_size == 0:
                    continue
                mtime = candidate.stat().st_mtime
                if mtime > newest_mtime:
                    newest, newest_mtime = candidate, mtime
    size = newest.stat().st_size if newest else 0
    return newest, newest_mtime, size


def deny(reason: str) -> None:
    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        },
        sys.stdout,
    )
    sys.exit(0)


def main() -> None:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return

    command = str(payload.get("tool_input", {}).get("command", ""))
    if not command or not targets_personal_data(command):
        return

    root = repo_root()
    backup, mtime, size = newest_backup(root)

    if backup is None:
        deny(
            "Команда может уничтожить данные Personal-стека, а непустого бэкапа "
            f"в {root / 'backups'} нет. Сначала выполните "
            "scripts/devops/backup-personal.ps1 и убедитесь, что файл создан."
        )

    age_hours = (time.time() - mtime) / 3600
    if age_hours > MAX_AGE_HOURS:
        deny(
            f"Последний бэкап Personal-стека — {backup.name}, ему "
            f"{age_hours:.1f} ч при допустимых {MAX_AGE_HOURS:.0f}. "
            "Сделайте свежий бэкап через scripts/devops/backup-personal.ps1, "
            "либо поднимите порог переменной KG_BACKUP_MAX_AGE_HOURS."
        )

    json.dump(
        {
            "systemMessage": (
                f"Деструктивная команда разрешена: бэкап {backup.name} "
                f"({size / 1024:.0f} КБ, возраст {age_hours:.1f} ч) на месте."
            )
        },
        sys.stdout,
    )


if __name__ == "__main__":
    main()
