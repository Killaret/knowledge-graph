#!/usr/bin/env python3
"""
Скрипт для загрузки тестовых данных через API Knowledge Graph.
Запуск:
    python scripts/seed_data.py [--api-url http://localhost:8080]

Требуется Python 3.8+ и библиотеки:
    pip install requests tqdm faker
"""

import argparse
import random
import time
import sys
from typing import List, Dict, Any, Optional, Union

import requests
from faker import Faker
from tqdm import tqdm

# ----------------------------------------------------------------------
# Конфигурация
# ----------------------------------------------------------------------
DEFAULT_API_URL = "http://localhost:8080"
CATEGORIES = ["book", "movie", "anime", "manga", "adaptation"]
NOTES_PER_CATEGORY = 30
TOTAL_NOTES = NOTES_PER_CATEGORY * len(CATEGORIES)

# ----------------------------------------------------------------------
# Типы связей в Knowledge Graph
# ----------------------------------------------------------------------
# link_type   | Описание                    | Рекомендуемый вес
# ------------|-----------------------------|------------------
# reference   | Прямая ссылка/цитирование   | 0.7 – 1.0
# dependency  | Зависимость                 | 0.5 – 0.9
# related     | Общая ассоциация            | 0.3 – 0.6
# custom      | Пользовательская связь      | 0.1 – 1.0
# ----------------------------------------------------------------------
LINK_TYPES = {
    "reference": (0.7, 1.0),   # Прямая ссылка/цитирование
    "dependency": (0.5, 0.9),  # Зависимость
    "related": (0.3, 0.6),     # Общая ассоциация
    "custom": (0.1, 1.0),      # Пользовательская связь
}

fake = Faker('ru_RU')

# ----------------------------------------------------------------------
# Шаблоны контента (~500 слов)
# ----------------------------------------------------------------------
CONTENT_TEMPLATES = {
    "book": """Роман «{title}» — это глубокое философское произведение, исследующее темы любви, потери и искупления. Главный герой, {hero}, оказывается втянут в череду загадочных событий после получения таинственного письма. Действие разворачивается в {location}, где каждый персонаж хранит свои секреты. Автор мастерски плетёт интригу, заставляя читателя гадать до последней страницы. В книге поднимаются вопросы морального выбора, цены успеха и истинной дружбы. Второстепенные линии добавляют глубины: история юной {heroine}, её борьба за независимость, и старый профессор {mentor}, чьи советы меняют судьбу героя. Описания природы и городских пейзажей создают неповторимую атмосферу. Критики отмечают неожиданную развязку и сильный эмоциональный финал. Произведение уже сравнивают с классикой жанра, а экранизация по книге находится в разработке. Читатели по всему миру обсуждают скрытые символы и аллюзии на современное общество. Эта книга обязательна к прочтению для всех ценителей умной прозы.

Дополнительные детали сюжета: {hero} встречает таинственного незнакомца, который раскрывает ему секрет древнего артефакта. В поисках правды герой отправляется в путешествие, полное опасностей и открытий. Каждая глава добавляет новые слои к характерам персонажей, а диалоги наполнены скрытым смыслом. Роман получил престижную литературную премию и был переведён на двадцать языков. Критики называют его «прорывом десятилетия» и предрекают долгую жизнь на полках книжных магазинов.""",

    "movie": """Фильм «{title}» режиссёра {director} — это визуальный шедевр, который держит в напряжении от первых кадров до финальных титров. В центре сюжета — {hero}, чья размеренная жизнь рушится после случайной встречи с загадочной незнакомкой. События разворачиваются в {location}, где переплетаются судьбы нескольких персонажей. Операторская работа завораживает: каждый кадр можно ставить на паузу и рассматривать как картину. Музыкальное сопровождение усиливает эмоциональное воздействие, а актёрский состав подобран безупречно. Критики особо выделяют игру {actor}, исполнившего роль второго плана. Сценарий полон неожиданных поворотов, а диалоги остроумны и реалистичны. Фильм уже собрал несколько наград на международных фестивалях и претендует на «Оскар». Зрители отмечают, что после просмотра хочется немедленно обсудить увиденное с друзьями. Это тот редкий случай, когда блокбастер сочетает зрелищность с глубоким смыслом. Настоятельно рекомендуется к просмотру на большом экране.

В фильме также затрагиваются темы предательства и искупления. Персонаж {heroine} добавляет романтическую линию, которая органично вплетена в основной сюжет. Спецэффекты выполнены на высочайшем уровне, а экшн-сцены сняты с использованием новейших технологий. Бюджет картины составил более ста миллионов долларов, и каждый цент виден на экране. После премьеры фильм возглавил прокат в тридцати странах и получил восторженные отзывы как от зрителей, так и от профессиональных критиков.""",

    "anime": """Аниме-сериал «{title}» студии {studio} стал настоящим хитом сезона. История повествует о {hero}, обычном старшекласснике, который неожиданно получает сверхъестественные способности. Теперь ему предстоит балансировать между школьной жизнью, первой любовью и битвами с демонами, угрожающими {location}. Визуальный стиль поражает проработкой деталей и плавностью анимации, особенно в боевых сценах. Музыкальное сопровождение от {composer} идеально ложится на настроение каждой серии. Персонажи прописаны с большой любовью: у каждого своя мотивация и трагичная предыстория. Зрители особенно полюбили цундэрэ {heroine} и загадочного сэнсэя {mentor}. Сюжет не скатывается в клише, а умело играет с жанровыми ожиданиями. Фансервис присутствует, но не отвлекает от основного повествования. Сериал уже продлён на второй сезон, а манга-первоисточник переживает новый всплеск популярности. Настоятельно рекомендуется всем поклонникам жанра сёнэн.

Каждая серия длится двадцать четыре минуты и заканчивается на самом интересном месте, заставляя зрителя с нетерпением ждать продолжения. Открывающая и закрывающая темы в исполнении популярных японских групп стали хитами и возглавили чарты. Фанаты активно создают фанарт и теории о дальнейшем развитии сюжета. Студия уже анонсировала полнометражный фильм, который выйдет в следующем году и расскажет предысторию одного из ключевых персонажей.""",

    "manga": """Манга «{title}» авторства {author} — это захватывающая история, действие которой разворачивается в альтернативном {location}. Главный герой, {hero}, мечтает стать величайшим {profession}, но его путь усеян препятствиями. Судьба сводит его с эксцентричным наставником {mentor} и загадочной девушкой {heroine}, которая хранит опасный секрет. Рисунок {author} отличается детализацией и динамикой: каждая панель передаёт движение и эмоции. Боевые сцены читаются на одном дыхании, а комедийные вставки разряжают обстановку. Сюжет не топчется на месте — каждая глава добавляет новые загадки и раскрывает характеры персонажей. Читатели отмечают глубокую проработку мира и его внутреннюю логику. Манга уже издаётся в десяти странах, а аниме-адаптация находится в производстве. Фанаты строят теории о финале и ждут каждый новый том с нетерпением. Это произведение обязано быть в коллекции каждого ценителя качественной манги.

На данный момент выпущено пятнадцать томов, и автор планирует завершить историю на двадцатом томе. Рейтинг манги на популярных сайтах составляет 9.2 из 10, а тиражи превысили пять миллионов экземпляров. В Японии манга регулярно попадает в еженедельные чарты продаж, а критики хвалят её за нестандартный подход к жанру и проработанных персонажей.""",

    "adaptation": """Экранизация «{title}», основанная на {source_type} {source_title}, вызывает горячие споры среди фанатов. С одной стороны, создатели бережно перенесли ключевые сцены и сохранили дух оригинала. С другой — некоторые сюжетные линии были изменены или вырезаны ради хронометража. Режиссёр {director} в интервью признался, что хотел рассказать историю {hero} так, чтобы она была понятна и новым зрителям, и преданным поклонникам. Визуальные эффекты на высоте: {location} выглядит именно так, как его описывали в книге. Актёрский состав в целом справляется, хотя выбор {actor} на роль {heroine} вызвал неоднозначную реакцию. Саундтрек от {composer} заслуживает отдельных похвал — музыка идеально подчёркивает эмоциональные моменты. Фильм уже собрал хорошую кассу в первый уикенд, а критики дают сдержанно-положительные отзывы. Стоит ли смотреть? Если вы фанат первоисточника — приготовьтесь к компромиссам. Если же вы новичок — это отличный способ познакомиться с удивительным миром, придуманным {author}.

Продолжительность фильма составляет два часа двадцать минут, и за это время зритель успевает погрузиться в атмосферу оригинала. Некоторые фанаты организовали петицию с требованием выпустить режиссёрскую версию, которая включала бы вырезанные сцены. Студия пока не комментирует эти запросы, но не исключает выпуска расширенного издания на цифровых носителях."""
}

# ----------------------------------------------------------------------
# Генераторы данных
# ----------------------------------------------------------------------
def generate_title(category: str, index: int) -> str:
    """Генерирует правдоподобное название для категории."""
    if category == "book":
        return f"{fake.catch_phrase()} — {fake.word().capitalize()} {fake.word().capitalize()}"
    elif category == "movie":
        return f"{fake.bs().title()} {random.randint(2, 5)}"
    elif category == "anime":
        suffixes = [": Хроники", ": Перерождение", "!!", "~", " Zero", " Kai"]
        return f"{fake.word().capitalize()} {fake.word().capitalize()}{random.choice(suffixes)}"
    elif category == "manga":
        return f"{fake.word().capitalize()} no {fake.word().capitalize()}"
    else:  # adaptation
        return f"{fake.catch_phrase()} (Экранизация)"

def generate_content(category: str, title: str) -> str:
    """Генерирует контент ~500 слов на основе шаблона."""
    template = CONTENT_TEMPLATES[category]
    # Подстановки для разнообразия
    subs = {
        "title": title,
        "hero": fake.name(),
        "heroine": fake.name_female(),
        "location": fake.city(),
        "mentor": fake.name(),
        "director": fake.name(),
        "actor": fake.name(),
        "studio": random.choice(["Kyoto Animation", "MAPPA", "Bones", "Ufotable", "Madhouse"]),
        "composer": fake.name(),
        "author": fake.name(),
        "profession": random.choice(["воином", "магом", "детективом", "поваром", "пилотом"]),
        "source_type": random.choice(["роману", "манге", "комиксу", "повести"]),
        "source_title": fake.sentence(nb_words=4).rstrip('.'),
    }
    content = template.format(**subs)
    # Добавляем ещё немного текста, если вдруг меньше 500 слов
    if len(content.split()) < 450:
        extra = "\n\n" + fake.paragraph(nb_sentences=10)
        content += extra
    return content

def generate_note_payload(category: str, index: int) -> Dict[str, Any]:
    """Создаёт payload для POST /notes."""
    title = generate_title(category, index)
    content = generate_content(category, title)
    # Разнообразные типы небесных тел для всех заметок
    ALL_CELESTIAL_TYPES = ["star", "planet", "moon", "comet", "galaxy", "nebula", "asteroid", "satellite", "blackhole"]
    return {
        "title": title,
        "content": content,
        "metadata": {
            "category": category,
            "word_count": len(content.split()),
            "type": random.choice(ALL_CELESTIAL_TYPES)  # Тип в metadata для backend
        }
    }

# ----------------------------------------------------------------------
# Исключения для обработки ошибок
# ----------------------------------------------------------------------
class APIError(Exception):
    """Ошибка API с подробным описанием"""
    def __init__(self, method: str, url: str, status_code: int = None, response_text: str = None, message: str = None):
        self.method = method
        self.url = url
        self.status_code = status_code
        self.response_text = response_text
        self.message = message
        super().__init__(self._format_message())
    
    def _format_message(self) -> str:
        parts = [f"[ERR] API Error: {self.method} {self.url}"]
        if self.status_code:
            parts.append(f"   Status Code: {self.status_code}")
        if self.message:
            parts.append(f"   Message: {self.message}")
        if self.response_text:
            parts.append(f"   Response: {self.response_text[:500]}")
        return "\n".join(parts)

class SeedingError(Exception):
    """Ошибка при загрузке тестовых данных"""
    pass

# ----------------------------------------------------------------------
# Работа с API
# ----------------------------------------------------------------------
class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.created_notes: List[str] = []  # Для отслеживания созданных заметок
        self.created_links: List[str] = []  # Для отслеживания созданных связей

    def _request(self, method: str, path: str, raise_on_error: bool = True, **kwargs) -> Optional[Union[Dict[str, Any], List[Dict[str, Any]]]]:
        """
        Выполняет HTTP запрос к API.
        
        Args:
            method: HTTP метод
            path: путь URL
            raise_on_error: если True, выбрасывает APIError при ошибке
            **kwargs: дополнительные параметры для requests
        
        Returns:
            JSON ответ или None при ошибке (если raise_on_error=False)
        """
        url = f"{self.base_url}{path}"
        try:
            resp = self.session.request(method, url, **kwargs)
            resp.raise_for_status()
            if resp.status_code == 204:
                return {}
            return resp.json()
        except requests.exceptions.HTTPError as e:
            status_code = e.response.status_code if e.response else None
            response_text = e.response.text if e.response else None
            # Log detailed error for 500 errors
            if status_code == 500:
                print(f"\n[ERR] SERVER ERROR 500:")
                print(f"   URL: {method} {url}")
                print(f"   Response: {response_text[:500] if response_text else 'No response body'}")
                print(f"   Request body: {kwargs.get('json', 'N/A')}")
            error = APIError(method, url, status_code, response_text, str(e))
            if raise_on_error:
                raise error
            print(str(error))
            return None
        except requests.exceptions.RequestException as e:
            error = APIError(method, url, message=str(e))
            if raise_on_error:
                raise error
            print(str(error))
            return None

    def get_all_notes(self, limit: int = 1000) -> List[Dict[str, Any]]:
        """GET /notes с пагинацией - получает все заметки"""
        all_notes = []
        offset = 0
        
        while True:
            result = self._request("GET", f"/notes?limit={limit}&offset={offset}")
            if not isinstance(result, dict):
                raise APIError("GET", f"/notes?limit={limit}&offset={offset}", 
                              message=f"Unexpected response format: expected dict, got {type(result).__name__}")
            
            notes = result.get("notes", [])
            if not isinstance(notes, list):
                raise APIError("GET", f"/notes?limit={limit}&offset={offset}",
                              message=f"Unexpected 'notes' field format: expected list, got {type(notes).__name__}")
            
            all_notes.extend(notes)
            total = result.get("total", 0)
            
            # Проверяем, получили ли все заметки
            if len(all_notes) >= total or len(notes) == 0:
                break
                
            offset += len(notes)
        
        return all_notes

    def delete_note(self, note_id: str, raise_on_error: bool = False) -> bool:
        """DELETE /notes/{id}"""
        try:
            resp = self._request("DELETE", f"/notes/{note_id}", raise_on_error=raise_on_error)
            return resp is not None
        except APIError:
            if raise_on_error:
                raise
            return False

    def create_note(self, payload: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """POST /notes"""
        result = self._request("POST", "/notes", json=payload)
        if isinstance(result, dict) and "id" in result:
            self.created_notes.append(result["id"])
            return result
        raise APIError("POST", "/notes", message=f"Invalid response format: missing 'id' field. Response: {result}")

    def create_link(self, source_id: str, target_id: str, description: str = "",
                    weight: float = 1.0, link_type: Optional[str] = None,
                    skip_duplicate: bool = True) -> Optional[Dict[str, Any]]:
        """
        POST /links — обязателен link_type (reference/dependency/related/custom).
        Если link_type не передан, он определяется по описанию.
        Если skip_duplicate=True, ошибка 409 (дубликат) игнорируется.
        """
        if link_type is None:
            if "экранизация" in description or "адаптация" in description:
                link_type = "reference"
            elif "основано" in description:
                link_type = "dependency"
            elif "похожее" in description:
                link_type = "related"
            else:
                link_type = "related"

        # Валидация link_type
        valid_types = ["reference", "dependency", "related", "custom"]
        if link_type not in valid_types:
            raise APIError("POST", "/links", 
                          message=f"Invalid link_type: '{link_type}'. Valid types: {valid_types}")

        payload = {
            "source_note_id": source_id,
            "target_note_id": target_id,
            "link_type": link_type,
            "weight": weight,
            "metadata": {"description": description}
        }
        
        try:
            time.sleep(0.2)  # Задержка для rate limit
            result = self._request("POST", "/links", json=payload)
            if isinstance(result, dict) and "id" in result:
                self.created_links.append(result["id"])
                return result
            raise APIError("POST", "/links", 
                          message=f"Invalid response format: missing 'id' field. Response: {result}")
        except APIError as e:
            if skip_duplicate and e.status_code == 409:
                return None
            raise

# ----------------------------------------------------------------------
# Логика загрузки
# ----------------------------------------------------------------------
def clear_database(api: APIClient) -> None:
    """Получает все заметки и удаляет их."""
    print("[CLEANUP] Очистка базы данных...")
    notes = api.get_all_notes()
    if not notes:
        print("   База уже пуста.")
        return

    print(f"   Найдено заметок: {len(notes)}. Удаление...")
    success = 0
    for note in tqdm(notes, desc="Удаление заметок", unit="note"):
        if api.delete_note(note["id"]):
            success += 1
        time.sleep(0.1)  # Задержка для rate limit
    print(f"[OK] Удалено {success} из {len(notes)} заметок.")
    time.sleep(2)

def cleanup_on_error(api: APIClient) -> None:
    """Очищает все созданные данные при ошибке"""
    print("\n[CLEANUP] ОЧИСТКА: Удаление созданных данных из-за ошибки...")
    
    # Удаляем созданные связи
    if api.created_links:
        print(f"   Удаление {len(api.created_links)} связей...")
        for link_id in api.created_links:
            try:
                time.sleep(0.1)  # Задержка для rate limit
                api._request("DELETE", f"/links/{link_id}", raise_on_error=False)
            except Exception:
                pass
    
    # Удаляем созданные заметки
    if api.created_notes:
        print(f"   Удаление {len(api.created_notes)} заметок...")
        for note_id in api.created_notes:
            try:
                time.sleep(0.1)  # Задержка для rate limit
                api.delete_note(note_id, raise_on_error=False)
                time.sleep(0.1)  # Задержка для rate limit
            except Exception:
                pass
    
    print("[OK] Очистка завершена")

def create_notes(api: APIClient) -> Dict[str, List[str]]:
    """
    Создаёт заметки по категориям.
    Возвращает словарь: category -> list of note IDs.
    При ошибке выбрасывает SeedingError.
    """
    created_ids = {cat: [] for cat in CATEGORIES}
    print(f"\n[NOTES] Создание {TOTAL_NOTES} заметок...")
    
    try:
        for cat in CATEGORIES:
            print(f"\n   Категория: {cat}")
            for i in tqdm(range(NOTES_PER_CATEGORY), desc=f"   {cat}", unit="note"):
                payload = generate_note_payload(cat, i)
                try:
                    result = api.create_note(payload)
                    if result:
                        created_ids[cat].append(result["id"])
                except APIError as e:
                    raise SeedingError(f"Ошибка создания заметки {i} в категории {cat}: {e}")
                time.sleep(0.15)
    except SeedingError:
        raise
    except Exception as e:
        raise SeedingError(f"Неожиданная ошибка при создании заметок: {e}")
    
    return created_ids

def create_links_between_categories(api: APIClient, ids_map: Dict[str, List[str]]) -> int:
    """
    Создаёт прямые связи между заметками разных категорий.
    Например: книга -> экранизация, манга -> аниме.
    При ошибке выбрасывает SeedingError.
    
    Returns:
        Количество созданных связей
    """
    print("\n[LINKS] Создание межкатегорийных связей...")
    total_links = 0

    try:
        # 1. Книги -> Экранизации
        if "book" in ids_map and "adaptation" in ids_map:
            pairs = min(len(ids_map["book"]), len(ids_map["adaptation"])) // 2
            for _ in range(pairs):
                src = random.choice(ids_map["book"])
                dst = random.choice(ids_map["adaptation"])
                try:
                    if api.create_link(src, dst, description="экранизация книги", weight=0.9):
                        total_links += 1
                    if api.create_link(dst, src, description="основано на книге", weight=0.8):
                        total_links += 1
                except APIError as e:
                    raise SeedingError(f"Ошибка создания связи book->adaptation: {e}")

        # 2. Манга -> Аниме
        if "manga" in ids_map and "anime" in ids_map:
            pairs = min(len(ids_map["manga"]), len(ids_map["anime"])) // 2
            for _ in range(pairs):
                src = random.choice(ids_map["manga"])
                dst = random.choice(ids_map["anime"])
                try:
                    if api.create_link(src, dst, description="аниме-адаптация", weight=0.95):
                        total_links += 1
                    if api.create_link(dst, src, description="основано на манге", weight=0.85):
                        total_links += 1
                except APIError as e:
                    raise SeedingError(f"Ошибка создания связи manga->anime: {e}")

        # 3. Фильмы -> Экранизации
        if "movie" in ids_map and "adaptation" in ids_map:
            pairs = min(len(ids_map["movie"]), len(ids_map["adaptation"])) // 3
            for _ in range(pairs):
                src = random.choice(ids_map["movie"])
                dst = random.choice(ids_map["adaptation"])
                try:
                    if api.create_link(src, dst, description="ремейк/адаптация", weight=0.7):
                        total_links += 1
                except APIError as e:
                    raise SeedingError(f"Ошибка создания связи movie->adaptation: {e}")
    except SeedingError:
        raise
    except Exception as e:
        raise SeedingError(f"Неожиданная ошибка при создании межкатегорийных связей: {e}")

    print(f"   Создано {total_links} межкатегорийных связей.")
    return total_links

def create_links_within_category(api: APIClient, ids_map: Dict[str, List[str]]) -> int:
    """Создаёт связи внутри одной категории (похожие произведения).
    При ошибке выбрасывает SeedingError.
    
    Returns:
        Количество созданных связей
    """
    print("\n[LINKS] Создание внутрикатегорийных связей...")
    total_links = 0
    
    try:
        for cat, ids in ids_map.items():
            if len(ids) < 2:
                continue
            for i in range(0, len(ids) - 1, 2):
                src = ids[i]
                dst = ids[i + 1]
                try:
                    if api.create_link(src, dst, description=f"похожее {cat}", weight=0.6, skip_duplicate=True):
                        total_links += 1
                    if api.create_link(dst, src, description=f"похожее {cat}", weight=0.6, skip_duplicate=True):
                        total_links += 1
                except APIError as e:
                    raise SeedingError(f"Ошибка создания связи в категории {cat}: {e}")
    except SeedingError:
        raise
    except Exception as e:
        raise SeedingError(f"Неожиданная ошибка при создании внутрикатегорийных связей: {e}")
    
    print(f"   Создано {total_links} внутрикатегорийных связей.")
    return total_links

def create_random_links(api: APIClient, all_ids: List[str], count: int = 20) -> int:
    """Создаёт случайные связи для увеличения связности графа.
    При ошибке выбрасывает SeedingError.
    
    Returns:
        Количество созданных связей
    """
    print(f"\n[LINKS] Создание {count} случайных связей...")
    created = 0
    
    try:
        for _ in tqdm(range(count), desc="Случайные связи", unit="link"):
            src = random.choice(all_ids)
            dst = random.choice(all_ids)
            if src == dst:
                continue
            # 30% случайных связей делаем custom
            if random.random() < 0.3:
                link_type = "custom"
                description = "пользовательская связь"
            else:
                link_type = "related"
                description = "связано"
            try:
                if api.create_link(src, dst, description=description, link_type=link_type, weight=0.3, skip_duplicate=True):
                    created += 1
            except APIError as e:
                raise SeedingError(f"Ошибка создания случайной связи: {e}")
    except SeedingError:
        raise
    except Exception as e:
        raise SeedingError(f"Неожиданная ошибка при создании случайных связей: {e}")
    
    print(f"   Создано {created} случайных связей.")
    return created

def main():
    parser = argparse.ArgumentParser(description="Загрузка тестовых данных в Knowledge Graph")
    parser.add_argument("--api-url", default=DEFAULT_API_URL, help="URL API бэкенда")
    parser.add_argument("--skip-clear", action="store_true", help="Пропустить очистку БД")
    parser.add_argument("--no-random-links", action="store_true", help="Не создавать случайные связи")
    parser.add_argument("--no-cleanup-on-error", action="store_true", help="Не очищать созданные данные при ошибке")
    args = parser.parse_args()

    api = APIClient(args.api_url)
    print(f"[START] Подключение к API: {args.api_url}")

    # Проверка доступности API
    try:
        health = api._request("GET", "/health")
        if health is None:
            print("[ERR] Не удалось подключиться к API. Убедитесь, что бэкенд запущен.")
            sys.exit(1)
    except APIError as e:
        print(f"[ERR] Не удалось подключиться к API: {e}")
        sys.exit(1)

    success = False
    try:
        if not args.skip_clear:
            clear_database(api)

        ids_by_category = create_notes(api)
        all_ids = [id for ids in ids_by_category.values() for id in ids]

        if not all_ids:
            raise SeedingError("Не создано ни одной заметки")

        create_links_between_categories(api, ids_by_category)
        create_links_within_category(api, ids_by_category)
        if not args.no_random_links:
            create_random_links(api, all_ids, count=30)

        success = True
        print("\n[OK] Загрузка тестовых данных завершена!")
        print(f"   Всего создано заметок: {len(all_ids)}")
        print(f"   Всего создано связей: {len(api.created_links)}")
        print(f"   Категории: {', '.join(f'{k}: {len(v)}' for k, v in ids_by_category.items())}")
        print("\n[INFO] Теперь можно проверить работу графа и рекомендаций через фронтенд.")
        
    except SeedingError as e:
        print(f"\n[ERR] ОШИБКА ЗАГРУЗКИ: {e}")
        if not args.no_cleanup_on_error:
            cleanup_on_error(api)
        sys.exit(1)
    except APIError as e:
        print(f"\n[ERR] ОШИБКА API: {e}")
        if not args.no_cleanup_on_error:
            cleanup_on_error(api)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n\n[WARN] Прервано пользователем (Ctrl+C)")
        if not args.no_cleanup_on_error:
            cleanup_on_error(api)
        sys.exit(130)
    except Exception as e:
        print(f"\n[ERR] НЕОЖИДАННАЯ ОШИБКА: {e}")
        import traceback
        traceback.print_exc()
        if not args.no_cleanup_on_error:
            cleanup_on_error(api)
        sys.exit(1)

if __name__ == "__main__":
    main()