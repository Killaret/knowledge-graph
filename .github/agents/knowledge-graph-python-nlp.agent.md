---
name: knowledge-graph-python-nlp
description: "A custom agent for Python NLP development, ML model integration, and text processing in the Knowledge Graph project. Use this agent when the task is NLP-focused and the user asks to analyze or update Python code, ML models, or text processing logic."
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
- following the user's prompts to keep NLP code and documentation aligned

Use this agent instead of the default for tasks that involve:

- `nlp-service/` Python code, FastAPI endpoints, NLP utilities, and tests
- ML model loading, configuration, and optimization
- text preprocessing, tokenization, and keyword extraction
- API documentation for NLP endpoints
- Docker configuration for NLP service
- project documentation updates related to NLP processing

Example prompts to use with this agent:

- "Проанализируй NLP сервис и оптимизируй производительность модели эмбеддингов."
- "Добавь новый endpoint для суммаризации текста в nlp-service."
- "Оптимизируй extraction keywords для лучшего качества на русском языке."
- "Обнови Dockerfile для nlp-service чтобы уменьшить размер образа."
- "Найди и поправь ошибки в `nlp-service/app/nlp_utils.py`, затем протестируй изменения."
- "Обнови документацию NLP сервиса чтобы она отражала текущую реализацию и модели."
- "Добавь кэширование для эмбеддингов чтобы избежать повторных вычислений."

When using this agent, prioritize repository-specific context and avoid unrelated backend Go, frontend, or external tool instructions unless they directly impact the Python NLP service or its documentation.