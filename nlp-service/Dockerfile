FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    curl \
    dos2unix \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --no-cache-dir transformers tqdm scikit-learn scipy sentencepiece

# Set environment variables for HuggingFace
ENV HF_HUB_DISABLE_TELEMETRY=1
ENV HF_HUB_ENABLE_HF_TRANSFER=0

# Download NLTK data during build
RUN python -c "import nltk; nltk.download('punkt'); nltk.download('stopwords')"

COPY . .

HEALTHCHECK --interval=30s --timeout=10s --retries=30 --start-period=600s \
  CMD curl -f http://localhost:5000/health || exit 1

ENTRYPOINT ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]
