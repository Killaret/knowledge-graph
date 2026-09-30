"""Knowledge Graph NLP service."""

from . import config as _config

# Seed NLP env vars from the shared knowledge-graph.config.json before any
# submodule reads os.environ at import time (nlp_utils, models).
_config.seed_env_from_config()
