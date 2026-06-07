import nltk
import yake
import os
from sentence_transformers import SentenceTransformer
import logging

logger = logging.getLogger(__name__)

# Загрузка стоп-слов и токенизатора (один раз)
try:
    nltk.data.find('tokenizers/punkt')
except LookupError:
    nltk.download('punkt')
try:
    nltk.data.find('corpora/stopwords')
except LookupError:
    nltk.download('stopwords')

stop_words = set(nltk.corpus.stopwords.words('russian') + nltk.corpus.stopwords.words('english'))

# Модель эмбеддингов с обработкой ошибок
embedding_model = None
try:
    # Try loading with cache folder first
    cache_dir = os.environ.get('SENTENCE_TRANSFORMERS_HOME', '/root/.cache/torch/sentence_transformers')
    model_name = 'all-MiniLM-L6-v2'
    
    # Disable ETag check for mirror compatibility
    os.environ['HF_HUB_DISABLE_SYMLINKS'] = '1'
    os.environ['HF_HUB_DISABLE_TELEMETRY'] = '1'
    
    embedding_model = SentenceTransformer(model_name, cache_folder=cache_dir)
    logger.info("Embedding model loaded successfully")
except Exception as e:
    logger.error(f"Failed to load embedding model: {e}")
    # Try without cache folder as fallback
    try:
        embedding_model = SentenceTransformer(model_name)
        logger.info("Embedding model loaded without cache folder")
    except Exception as e2:
        logger.error(f"Failed to load embedding model without cache: {e2}")
        raise

# Экстрактор ключевых слов YAKE
kw_extractor = yake.KeywordExtractor(lan="ru", top=20, stopwords=stop_words)

def extract_keywords(text: str, top_n: int = 10) -> list:
    """Возвращает список кортежей (keyword, weight), где weight в [0..1]."""
    if not text or not text.strip():
        return []
    keywords = kw_extractor.extract_keywords(text)
    result = []
    for kw, score in keywords[:top_n]:
        weight = max(0.0, min(1.0, 1.0 - score))
        result.append((kw, weight))
    return result
