"""Seed NLP settings from the shared knowledge-graph.config.json.

Precedence matches backend and graph-service (CONFIG-AUDIT-1, решение 102):
environment > config file > code default. At startup the nlp.* section is
mapped onto environment variables via os.environ.setdefault, so the existing
os.environ.get readers keep working and anything already exported wins.

The file is located via KG_CONFIG_PATH, then /app/knowledge-graph.config.json
(the path it is copied to in the image and mounted at in compose), then the
repository root relative to this module for local runs and tests. A missing
or unreadable file is not an error — code defaults apply.
"""

import json
import os
from pathlib import Path
from typing import Optional, Tuple

CONFIG_PATH_ENV = "KG_CONFIG_PATH"
_CONTAINER_PATH = "/app/knowledge-graph.config.json"

# nlp.* key -> environment variable it seeds
_ENV_MAP = {
    "model_name": "NLP_MODEL_NAME",
    "max_text_length": "NLP_MAX_TEXT_LENGTH",
    "hf_home": "HF_HOME",
    "hf_hub_disable_telemetry": "HF_HUB_DISABLE_TELEMETRY",
    "hf_hub_offline": "HF_HUB_OFFLINE",
}


def _candidates():
    override = os.environ.get(CONFIG_PATH_ENV)
    if override:
        yield Path(override)
    yield Path(_CONTAINER_PATH)
    yield Path(__file__).resolve().parents[2] / "knowledge-graph.config.json"


def _load() -> Tuple[Optional[dict], Optional[Path]]:
    for path in _candidates():
        try:
            if path.is_file():
                return json.loads(path.read_text(encoding="utf-8")), path
        except (OSError, json.JSONDecodeError):
            continue
    return None, None


def _as_env(value) -> str:
    if isinstance(value, bool):
        return "1" if value else "0"
    return str(value)


def seed_env_from_config() -> Optional[Path]:
    """Fill missing NLP env vars from nlp.* in the shared config file.

    Returns the path the file was loaded from, or None when no readable
    config was found.
    """
    config, path = _load()
    if not config:
        return None
    nlp = config.get("nlp") or {}
    for key, env_name in _ENV_MAP.items():
        if key in nlp:
            os.environ.setdefault(env_name, _as_env(nlp[key]))
    return path
