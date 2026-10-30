#!/usr/bin/env python3
"""BACKUP-3: worker_personal must write event-driven backups into the synced
host folder, not into its own container layer.

Machine check, not YAML-reading-by-eye: renders the compose file through
`docker compose config --format json` and asserts that worker_personal both
mounts a host directory at /backups and points BACKUP_LOCAL_PATH at it.
Without the mount the files land in the container layer - invisible to the
host folder and lost on `docker compose down`.

The compose file declares `env_file: .env`, and `config` fails when the file
is absent (CI runners and fresh clones have none). An empty .env is created
for the duration of the render and removed afterwards; every interpolated
variable in the file carries a `:-` default, so the render is identical.

Exit 0 - mount and env are wired. Exit 1 - either is missing.
Exit 2 - the check itself could not run (no docker, render failed).
"""

import json
import shutil
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
COMPOSE_FILE = REPO_ROOT / "docker-compose.personal.yml"
ENV_FILE = REPO_ROOT / ".env"
SERVICE = "worker_personal"
MOUNT_TARGET = "/backups"
ENV_VAR = "BACKUP_LOCAL_PATH"


def main() -> int:
    if shutil.which("docker") is None:
        print("[ERROR] docker CLI not found - cannot render compose config")
        return 2

    created_env = False
    if not ENV_FILE.exists():
        ENV_FILE.touch()
        created_env = True
    try:
        result = subprocess.run(
            [
                "docker", "compose", "-f", str(COMPOSE_FILE),
                "config", "--format", "json",
            ],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
        )
    finally:
        if created_env:
            ENV_FILE.unlink()

    if result.returncode != 0:
        print(f"[ERROR] docker compose config failed:\n{result.stderr.strip()}")
        return 2

    config = json.loads(result.stdout)
    service = (config.get("services") or {}).get(SERVICE) or {}

    problems = []

    volumes = service.get("volumes") or []
    mounted = any(
        (isinstance(v, dict) and v.get("target") == MOUNT_TARGET)
        or (isinstance(v, str) and v.split(":")[-1] == MOUNT_TARGET)
        for v in volumes
    )
    if not mounted:
        problems.append(
            f"{SERVICE} has no volume mounted at {MOUNT_TARGET} - "
            "event-driven backups would land in the container layer"
        )

    env = service.get("environment") or {}
    if isinstance(env, list):
        env = dict(item.split("=", 1) for item in env)
    if env.get(ENV_VAR) != MOUNT_TARGET:
        problems.append(
            f"{SERVICE}.{ENV_VAR} is {env.get(ENV_VAR)!r}, expected {MOUNT_TARGET!r}"
        )

    if problems:
        print("Worker backup mount check failed:")
        for problem in problems:
            print(f"  {problem}")
        return 1

    print(
        f"Worker backup mount OK: {SERVICE} mounts a host directory at "
        f"{MOUNT_TARGET} and {ENV_VAR}={MOUNT_TARGET}."
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
