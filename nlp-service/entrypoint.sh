#!/bin/bash

# Entrypoint script for NLP service
# Downloads model on first run if not cached

set -e

echo "Starting NLP service..."

# Canonical model name env var with backward-compatible fallback.
NLP_MODEL_NAME="${NLP_MODEL_NAME:-${MODEL_NAME:-all-MiniLM-L6-v2}}"
export MODEL_NAME="${NLP_MODEL_NAME}"

echo "Using embedding model: ${NLP_MODEL_NAME}"

# HuggingFace cache directory for sentence-transformers models.
# The directory name convention is models--sentence-transformers--<model_name>.
HF_CACHE_DIR="${HF_HOME:-/root/.cache/huggingface}/hub"
CACHE_NAME="models--sentence-transformers--${NLP_MODEL_NAME}"

# Download model if not already cached
if [ ! -d "${HF_CACHE_DIR}/${CACHE_NAME}" ]; then
    echo "Model not found in cache. Downloading..."
    python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('${NLP_MODEL_NAME}')"
    echo "Model downloaded and cached successfully."
else
    echo "Model found in cache. Skipping download."
fi

# Start the application
exec "$@"
