---
name: knowledge-graph-nlp
description: "A custom agent for Python NLP development, FastAPI endpoints, HuggingFace sentence-transformers, and lazy model loading in the Knowledge Graph project. Use this agent when the task is NLP-focused."
applyTo:
  - "nlp-service/**"
  - "**/*.py"
  - "requirements.txt"
  - "Dockerfile"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- Python NLP code refinements, bug fixes, or feature work
- ML model integration and optimization (sentence-transformers, YAKE, NLTK)
- FastAPI endpoint development and testing
- text processing and keyword extraction algorithms
- Docker containerization for ML services
- reading, correcting, or extending documentation related to NLP service

Key constraints:
- Model loading is **lazy** (`ensure_model_loaded()` / `get_embedding_model()` in `nlp_utils.py`); never load at import time.
- Use `HF_HUB_OFFLINE=1` and `HF_HOME=/path/to/cache` in all environments.
- Uvicorn starts in ~1s; first `/health` call triggers model load (~15s from cache).

Example prompts:

- "Analyze NLP service and optimize embedding model performance."
- "Add a new endpoint for text summarization in nlp-service."
- "Optimize keyword extraction for better quality on Russian language."
- "Update Dockerfile for nlp-service to reduce image size."
- "Find and fix errors in `nlp-service/app/nlp_utils.py`, then test changes."

When using this agent, prioritize repository-specific context and avoid unrelated backend Go, frontend, or external tool instructions unless they directly impact the Python NLP service or its documentation.
