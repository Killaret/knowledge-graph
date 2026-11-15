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

import gzip
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import time
from pathlib import Path

PERSONAL_VOLUMES = ("pgdata_personal", "redisdata_personal", "mongodbdata_personal")

# A usable backup exists per datastore: Postgres AND MongoDB. A single fresh
# file must not cover the other store (BACKUP-4 item 2). Raw volume archives
# are named per volume, so they count only for the datastore they hold.
PG_PATTERNS = ("backup-personal-*", "personal-volumes-raw-*pgdata*")
MONGO_PATTERNS = ("backup-personal-*.mongo.gz", "personal-volumes-raw-*mongodbdata*")
MONGO_DUMP_SUFFIX = ".mongo.gz"

# Content proof, not just non-zero size (BACKUP-4 item 1): a 913-byte pg
# envelope with zero tables used to pass the freshness check.
GZIP_MAGIC = b"\x1f\x8b"
NOTES_SQL = re.compile(rb"\bnotes\b")
DUMP_TABLE_RE = re.compile(rb"\bTABLE\s+\S+\s+notes\b")
MONGO_NS = b"knowledge_personal."

# Commands that sweep every unused volume, personal ones included.
PRUNE_PATTERNS = (
    re.compile(r"\bdocker\s+volume\s+prune\b"),
    re.compile(r"\bdocker\s+system\s+prune\b"),
)

# Commands that destroy volumes belonging to the Personal stack specifically.
#
# Known bypass this guard cannot see: scripts/cleanup/cleanup-docker.ps1 and
# cleanup-docker.sh run `docker volume rm` / `docker system prune` inside a
# script file, so the hook payload (`pwsh -File cleanup-docker.ps1 ...`) does
# not match any pattern. That path is guarded inside the script itself via
# scripts/devops/check-personal-backup.{ps1,sh} — the check must live where
# the deletion happens, because the hook only sees agent commands.
PERSONAL_PATTERNS = (
    re.compile(r"\bdocker\s+volume\s+rm\b"),
    re.compile(r"\bdown\b.*(?:\s-v\b|--volumes\b)"),
    re.compile(r"\brm\s+-[a-z]*r"),
)


def _max_age_hours() -> float:
    """Freshness threshold for a usable backup, in hours.

    Single source of truth: backup-policy.env next to this script; the
    KG_BACKUP_MAX_AGE_HOURS environment variable overrides it per run.
    """
    env_value = os.environ.get("KG_BACKUP_MAX_AGE_HOURS")
    if env_value:
        return float(env_value)
    policy = Path(__file__).resolve().parent / "backup-policy.env"
    if policy.is_file():
        for line in policy.read_text().splitlines():
            if line.strip().startswith("KG_BACKUP_MAX_AGE_HOURS"):
                return float(line.split("=", 1)[1].strip())
    return 24.0


def resolve_backup_dir() -> Path:
    """Absolute path to the local backup directory.

    Single source of truth: the KG_BACKUP_DIR value from backup-policy.env
    next to this script. The KG_BACKUP_DIR environment variable overrides it.
    A leading ~ is expanded; a relative tail is joined to the user's home
    directory; an absolute path is used as-is.
    """
    env_value = os.environ.get("KG_BACKUP_DIR")
    value = env_value
    if not value:
        policy = Path(__file__).resolve().parent / "backup-policy.env"
        if policy.is_file():
            for line in policy.read_text().splitlines():
                if line.strip().startswith("KG_BACKUP_DIR"):
                    value = line.split("=", 1)[1].strip()
                    break
    if not value:
        raise RuntimeError(
            "KG_BACKUP_DIR is not set and not found in backup-policy.env"
        )
    value = os.path.expanduser(value)
    p = Path(value)
    if p.is_absolute():
        return p
    return Path.home() / p


MAX_AGE_HOURS = _max_age_hours()


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

# Shells that carry a nested command in an argument; the payload is inspected
# instead of the wrapper, so `bash -c "docker volume prune"` is not a bypass.
SHELL_WRAPPERS = {"bash", "sh", "zsh", "pwsh", "powershell", "cmd"}
SHELL_COMMAND_FLAGS = {"-c", "-Command", "-command", "/c", "/C"}
MAX_WRAPPER_DEPTH = 3


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


def nested_command(segment: list[str]) -> str | None:
    """Payload of a shell wrapper such as `bash -c "..."`, if this is one."""
    words = [w for w in segment if not ENV_ASSIGNMENT.match(w)]
    if not words:
        return None
    name = words[0].rsplit("/", 1)[-1].rsplit("\\", 1)[-1].lower()
    if name.removesuffix(".exe") not in SHELL_WRAPPERS:
        return None
    for index, word in enumerate(words[1:], start=1):
        if word in SHELL_COMMAND_FLAGS and index + 1 < len(words):
            return " ".join(words[index + 1 :])
    return None


def targets_personal_data(command: str, depth: int = 0) -> bool:
    """True when the command can destroy Personal stack volumes.

    Only segments that actually invoke a risky program are inspected, so a
    commit message or a doc edit that merely quotes such a command is ignored.
    A shell wrapper is unwrapped and its payload inspected instead. A command
    that cannot be parsed is treated as risky.
    """
    try:
        segments = split_segments(command)
    except ValueError:
        segments = [command.split()]

    for segment in segments:
        if depth < MAX_WRAPPER_DEPTH:
            payload = nested_command(segment)
            if payload and targets_personal_data(payload, depth + 1):
                return True

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


def _is_gzip(path: Path) -> bool:
    try:
        with path.open("rb") as fh:
            return fh.read(2) == GZIP_MAGIC
    except OSError:
        return False


def _sql_dump_has_notes(path: Path) -> bool:
    """Plain pg_dump output (gzipped or not) counts only when it carries a
    COPY or INSERT for the notes table — an envelope alone is not a backup."""
    try:
        opener = gzip.open if _is_gzip(path) else open
        with opener(path, "rb") as fh:  # type: ignore[arg-type]
            for line in fh:
                if (line.startswith(b"COPY ") or line.startswith(b"INSERT INTO ")) and NOTES_SQL.search(line):
                    return True
    except (OSError, EOFError, gzip.BadGzipFile):
        return False
    return False


def _custom_dump_has_tables(path: Path) -> bool:
    """pg custom-format dumps are listed by pg_restore. No pg_restore (or a
    failed listing) means the file cannot be verified — refuse, not guess."""
    pg_restore = shutil.which("pg_restore")
    if pg_restore is None:
        return False
    try:
        out = subprocess.run([pg_restore, "--list", str(path)], capture_output=True, timeout=60)
    except (OSError, subprocess.SubprocessError):
        return False
    return out.returncode == 0 and DUMP_TABLE_RE.search(out.stdout) is not None


def _mongo_dump_has_namespace(path: Path) -> bool:
    """A mongodump archive must mention the personal DB namespace, either in
    plaintext archive headers or inside a whole-file gzip stream."""
    try:
        if MONGO_NS in path.read_bytes():
            return True
        if _is_gzip(path):
            with gzip.open(path, "rb") as fh:
                return MONGO_NS in fh.read(8 * 1024 * 1024)
    except (OSError, EOFError, gzip.BadGzipFile):
        return False
    return False


def backup_has_data(candidate: Path) -> bool:
    """A backup counts only if it provably holds data, not just bytes."""
    if not candidate.is_file() or candidate.stat().st_size == 0:
        return False
    name = candidate.name
    if name.endswith((".sql", ".sql.gz")):
        return _sql_dump_has_notes(candidate)
    if name.endswith(".dump"):
        return _custom_dump_has_tables(candidate)
    if name.endswith(MONGO_DUMP_SUFFIX):
        return _mongo_dump_has_namespace(candidate)
    # Raw per-volume archives are emergency copies; non-empty is all they
    # can prove without being mounted.
    if name.startswith("personal-volumes-raw-"):
        return True
    return False


def newest_usable_backup(backup_dir: Path, kind: str) -> tuple[Path | None, float, int]:
    """Newest file that passes backup_has_data for kind 'pg' or 'mongo'."""
    patterns = MONGO_PATTERNS if kind == "mongo" else PG_PATTERNS
    newest: Path | None = None
    newest_mtime = 0.0
    if backup_dir.is_dir():
        for pattern in patterns:
            for candidate in backup_dir.glob(pattern):
                if kind == "pg" and candidate.name.endswith(MONGO_DUMP_SUFFIX):
                    continue
                if not backup_has_data(candidate):
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

    try:
        backup_dir = resolve_backup_dir()
    except RuntimeError as exc:
        deny(f"Команда может уничтожить данные Personal-стека, но не удалось "
             f"определить каталог бэкапа: {exc}. Проверьте "
             "scripts/devops/backup-policy.env.")

    missing = []
    stale = []
    seen = []
    for kind, label in (("pg", "Postgres"), ("mongo", "MongoDB")):
        backup, mtime, size = newest_usable_backup(backup_dir, kind)
        if backup is None:
            missing.append(label)
            continue
        age_hours = (time.time() - mtime) / 3600
        if age_hours > MAX_AGE_HOURS:
            stale.append(f"{backup.name} ({age_hours:.1f} ч)")
            continue
        seen.append(f"{backup.name} ({size / 1024:.0f} КБ, {age_hours:.1f} ч)")

    if missing:
        deny(
            "Команда может уничтожить данные Personal-стека, а пригодного "
            f"бэкапа нет: {', '.join(missing)} — в {backup_dir} нужен дамп "
            "с реальным содержимым (таблица notes / пространство имён "
            "knowledge_personal) или сырой архив тома. Выполните "
            "scripts/devops/backup-personal.ps1 и убедитесь, что файлы созданы."
        )

    if stale:
        deny(
            f"Последний бэкап устарел: {', '.join(stale)} при допустимых "
            f"{MAX_AGE_HOURS:.0f} ч. Сделайте свежий бэкап через "
            "scripts/devops/backup-personal.ps1, либо поднимите порог "
            "переменной KG_BACKUP_MAX_AGE_HOURS."
        )

    json.dump(
        {
            "systemMessage": (
                f"Деструктивная команда разрешена: бэкапы на месте — "
                f"{'; '.join(seen)}."
            )
        },
        sys.stdout,
    )


if __name__ == "__main__":
    main()
