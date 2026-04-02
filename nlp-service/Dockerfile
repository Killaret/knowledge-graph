FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*
# Обновляем pip
RUN pip install --upgrade pip

COPY requirements.txt .

RUN pip install --no-cache-dir --default-timeout=100 --retries=5 \
    torch==2.0.1 --index-url https://download.pytorch.org/whl/cpu

RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2

RUN pip install --no-cache-dir -r requirements.txt

RUN pip install --no-cache-dir transformers tqdm scikit-learn scipy sentencepiece huggingface-hub

RUN pip install --no-cache-dir torchvision==0.15.2 --index-url https://download.pytorch.org/whl/cpu || true

RUN python -c "import nltk; nltk.download('punkt'); nltk.download('stopwords')"

COPY . .

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]