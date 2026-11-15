"""CONFIG-AUDIT-1 / решение 102: NLP reads the shared knowledge-graph.config.json.

Precedence: environment > config file > code default.
"""
import json
import os
import sys
from pathlib import Path

import pytest

# Add the parent directory to the path to import app modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from app import config as app_config


def _write_config(tmp_path: Path, nlp: dict) -> Path:
    cfg = tmp_path / "knowledge-graph.config.json"
    cfg.write_text(json.dumps({"nlp": nlp}), encoding="utf-8")
    return cfg


class TestSeedEnvFromConfig:
    def test_env_wins_over_file(self, tmp_path, monkeypatch):
        cfg = _write_config(tmp_path, {"model_name": "from-file"})
        monkeypatch.setenv(app_config.CONFIG_PATH_ENV, str(cfg))
        monkeypatch.setenv("NLP_MODEL_NAME", "from-env")

        assert app_config.seed_env_from_config() == cfg
        assert os.environ["NLP_MODEL_NAME"] == "from-env"

    def test_file_wins_over_default(self, tmp_path, monkeypatch):
        cfg = _write_config(
            tmp_path,
            {
                "model_name": "from-file",
                "max_text_length": 123,
                "hf_hub_offline": False,
                "hf_hub_disable_telemetry": True,
                "hf_home": "/tmp/hf-from-file",
            },
        )
        monkeypatch.setenv(app_config.CONFIG_PATH_ENV, str(cfg))
        for name in (
            "NLP_MODEL_NAME",
            "NLP_MAX_TEXT_LENGTH",
            "HF_HUB_OFFLINE",
            "HF_HUB_DISABLE_TELEMETRY",
            "HF_HOME",
        ):
            monkeypatch.delenv(name, raising=False)

        app_config.seed_env_from_config()

        assert os.environ["NLP_MODEL_NAME"] == "from-file"
        assert os.environ["NLP_MAX_TEXT_LENGTH"] == "123"
        assert os.environ["HF_HUB_OFFLINE"] == "0"  # bool -> "1"/"0"
        assert os.environ["HF_HUB_DISABLE_TELEMETRY"] == "1"
        assert os.environ["HF_HOME"] == "/tmp/hf-from-file"

    def test_missing_file_returns_none_and_seeds_nothing(self, tmp_path, monkeypatch):
        missing = tmp_path / "absent.json"
        monkeypatch.setattr(app_config, "_candidates", lambda: iter([missing]))
        monkeypatch.delenv("NLP_MODEL_NAME", raising=False)

        assert app_config.seed_env_from_config() is None
        assert "NLP_MODEL_NAME" not in os.environ

    def test_unreadable_json_is_skipped(self, tmp_path, monkeypatch):
        bad = tmp_path / "bad.json"
        bad.write_text("{not json", encoding="utf-8")
        monkeypatch.setattr(app_config, "_candidates", lambda: iter([bad]))

        assert app_config.seed_env_from_config() is None
