#!/bin/bash

# Entrypoint script for NLP service
# Downloads model on first run if not cached

set -e

echo "Starting NLP service..."

# Canonical model name env var with backward-compatible fallback.
NLP_MODEL_NAME="${NLP_MODEL_NAME:-paraphrase-multilingual-MiniLM-L12-v2}"
export NLP_MODEL_NAME

echo "Using embedding model: ${NLP_MODEL_NAME}"

# HuggingFace cache directory for sentence-transformers models.
# The directory name convention is models--sentence-transformers--<model_name>.
HF_CACHE_DIR="${HF_HOME:-/root/.cache/huggingface}/hub"
CACHE_NAME="models--sentence-transformers--${NLP_MODEL_NAME}"

# Download model if not already cached. snapshot_download must populate the
# same HuggingFace hub cache that nlp_utils.py reads at runtime.
if [ ! -d "${HF_CACHE_DIR}/${CACHE_NAME}" ]; then
    echo "Model not found in cache. Downloading..."
    python -c "import os; from huggingface_hub import snapshot_download; snapshot_download(repo_id='sentence-transformers/${NLP_MODEL_NAME}', cache_dir='${HF_CACHE_DIR}')"
    echo "Model downloaded and cached successfully."
else
    echo "Model found in cache. Skipping download."
fi

# Start the application
exec "$@"
