import nltk
import yake
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

# Модель эмбеддингов
embedding_model = SentenceTransformer('all-MiniLM-L6-v2')
logger.info("Embedding model loaded")

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