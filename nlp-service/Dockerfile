FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu
RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --no-cache-dir transformers tqdm scikit-learn scipy sentencepiece huggingface-hub

# Pre-download model during build to avoid runtime issues
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('all-MiniLM-L6-V2')"

RUN python -c "import nltk; nltk.download('punkt'); nltk.download('stopwords')"

COPY . .

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]