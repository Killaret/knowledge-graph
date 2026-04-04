FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

# Устанавливаем CPU-версию PyTorch (легковесная)
RUN pip install --no-cache-dir torch==2.2.0 --index-url https://download.pytorch.org/whl/cpu

# Устанавливаем sentence-transformers без зависимостей, чтобы не переустанавливал torch
RUN pip install --no-cache-dir --no-deps sentence-transformers==2.2.2

# Устанавливаем остальные зависимости из requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Устанавливаем недостающие зависимости (без torch)
RUN pip install --no-cache-dir transformers tqdm scikit-learn scipy sentencepiece huggingface-hub

# Torchvision не нужна для CPU, удаляем или пропускаем
# RUN pip install --no-cache-dir torchvision==0.15.2 --index-url https://download.pytorch.org/whl/cpu || true

RUN python -c "import nltk; nltk.download('punkt'); nltk.download('stopwords')"

# Скачиваем модель sentence-transformers прямо во время сборки образа
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('all-MiniLM-L6-V2')"

COPY . .

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "5000"]