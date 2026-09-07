# Multi-stage build for NLP service
FROM python:3.11-slim AS builder

WORKDIR /app

# Keep cache under /root to match runtime expectations and compose mounts
ENV HF_HOME=/root/.cache/huggingface
ENV NLTK_DATA=/root/nltk_data

# Build-only dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Create a virtual environment and install all Python dependencies
RUN python -m venv /app/.venv
ENV PATH="/app/.venv/bin:$PATH" VIRTUAL_ENV=/app/.venv

COPY requirements.txt .

RUN pip install --no-cache-dir --upgrade pip
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --no-cache-dir transformers tqdm scikit-learn scipy sentencepiece

# Set HuggingFace / NLTK environment before downloading data
ENV HF_HUB_DISABLE_TELEMETRY=1
ENV HF_HUB_ENABLE_HF_TRANSFER=0
ENV HF_HUB_DISABLE_XET=1

# Build-time embedding model selection.
# When the test stack or a rebuild passes a different model via build args,
# the image will preload that one instead.
ARG NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
ENV NLP_MODEL_NAME=${NLP_MODEL_NAME}

# Download NLTK data during build
RUN python -c "import nltk; nltk.download('punkt', download_dir='/root/nltk_data'); nltk.download('stopwords', download_dir='/root/nltk_data')"

# Pre-download embedding model so the service can start offline (HF_HUB_OFFLINE=1)
RUN python -c "from huggingface_hub import snapshot_download; snapshot_download(repo_id='sentence-transformers/${NLP_MODEL_NAME}', cache_dir='${HF_HOME}/hub')"

COPY . .

# Runtime stage
FROM python:3.11-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    dos2unix \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

ARG NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
ENV HF_HOME=/root/.cache/huggingface
ENV NLTK_DATA=/root/nltk_data
ENV HF_HUB_DISABLE_TELEMETRY=1
ENV HF_HUB_ENABLE_HF_TRANSFER=0
ENV HF_HUB_DISABLE_XET=1
ENV NLP_MODEL_NAME=${NLP_MODEL_NAME}
ENV PATH="/app/.venv/bin:$PATH" VIRTUAL_ENV=/app/.venv

COPY --from=builder /app/.venv /app/.venv
COPY --from=builder /root/.cache/huggingface /root/.cache/huggingface
COPY --from=builder /root/nltk_data /root/nltk_data
COPY --from=builder /app/app /app/app
COPY --from=builder /app/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh && dos2unix /app/entrypoint.sh

EXPOSE 5000

HEALTHCHECK --interval=30s --timeout=10s --retries=30 --start-period=600s \
  CMD curl -f http://localhost:5000/health || exit 1

ENTRYPOINT ["./entrypoint.sh"]
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000", "--workers", "1", "--timeout-keep-alive", "5", "--timeout-graceful-shutdown", "30"]
