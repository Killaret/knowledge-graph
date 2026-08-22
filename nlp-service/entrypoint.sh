#!/bin/bash

# Entrypoint script for NLP service
# Downloads model on first run if not cached

set -e

echo "Starting NLP service..."

# Download model if not already cached
echo "Checking for cached model..."
if [ ! -d "${HF_HOME:-/root/.cache/huggingface}/hub/models--sentence-transformers--all-MiniLM-L6-v2" ]; then
    echo "Model not found in cache. Downloading..."
    python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('all-MiniLM-L6-v2')"
    echo "Model downloaded and cached successfully."
else
    echo "Model found in cache. Skipping download."
fi

# Start the application
exec "$@"
