#!/usr/bin/env python3
"""Generate synthetic notes with known cluster labels and embed them via the
Personal NLP service (/embed). Output: synth.json in eval dataset format."""
import json
import urllib.request
import uuid

NLP = "http://127.0.0.1:5001/embed"

NOTES = [
    # education — system analysis / courses
    ("Coursera: System Analysis and Design — неделя 3", "education"),
    ("Специализация «Системный аналитик» — программа курса", "education"),
    ("Лекция: декомпозиция требований и границы системы", "education"),
    ("Конспект: нотации моделирования процессов BPMN", "education"),
    ("Курс по базам данных: нормализация и индексы", "education"),
    ("Учебник по архитектуре: шаблоны корпоративных приложений", "education"),
    ("Методичка: сбор требований у стейкхолдеров", "education"),
    ("Видеокурс: проектирование распределённых систем", "education"),
    ("Статья: доменно-ориентированное проектирование на практике", "education"),
    ("Шпаргалка по UML: диаграммы классов и последовательностей", "education"),
    # jobs — job boards / career
    ("LinkedIn Jobs — вакансии системного аналитика в Европе", "jobs"),
    ("Indeed: remote backend developer positions", "jobs"),
    ("Glassdoor — отзывы о компаниях и зарплаты", "jobs"),
    ("HH.ru: резюме и отклики на вакансии", "jobs"),
    ("Wellfound (AngelList) — стартапы и вакансии", "jobs"),
    ("Career profile: CV на английском, шаблоны резюме", "jobs"),
    ("StepStone — работа в Германии для IT-специалистов", "jobs"),
    ("RelocateMe — вакансии с релокацией", "jobs"),
    # books — book download / reading pages
    ("Читать онлайн: «Фундаментальные подходы к архитектуре ПО»", "books"),
    ("Скачать PDF: «Чистая архитектура» Роберт Мартин", "books"),
    ("ЛитРес: «Мастер и Маргарита» — читать фрагмент", "books"),
    ("Flibusta — библиотека epub и fb2", "books"),
    ("Bookmate: «Атлант расправил плечи» — читать онлайн", "books"),
    ("Гутенберг: классическая литература на английском", "books"),
    ("Аннотация книги «Системный анализ» — оглавление и отзывы", "books"),
    ("OZON: «Грокаем алгоритмы» — купить бумажную", "books"),
    # manga
    ("ReadManga: Берсерк, том 41 — читать онлайн", "manga"),
    ("MangaLib: Ван-Пис, глава 1090", "manga"),
    ("MangaPlus: официальные главы на английском", "manga"),
    ("Манхва «Всеведущий читатель» — новые главы", "manga"),
    ("Ранобэ и веб-новеллы: каталог переводов", "manga"),
    ("Anime News Network — новости аниме и манги", "manga"),
    # english — language learning
    ("EnglishSpace: грамматика, времена глаголов", "english"),
    ("Lingualeo — тренажёр слов и чтение текстов", "english"),
    ("Cambridge Dictionary — определения и произношение", "english"),
    ("Puzzle English: видеоуроки и тренировки", "english"),
    ("WooordHunt — словарь с примерами и транскрипцией", "english"),
    ("IELTS preparation: writing task 2 samples", "english"),
    ("Ted talks with subtitles — listening practice", "english"),
    ("Grammarly blog: правила пунктуации в английском", "english"),
    # devtools — tools/services
    ("Docker Hub — официальные образы контейнеров", "devtools"),
    ("GitHub: репозиторий проекта и issues", "devtools"),
    ("Postman — тестирование REST API", "devtools"),
    ("JetBrains: документация GoLand", "devtools"),
    ("Stack Overflow — вопросы по Go и Postgres", "devtools"),
    ("MDN Web Docs — справочник по веб-платформе", "devtools"),
    ("Yandex Cloud: консоль управления ВМ", "devtools"),
    ("Habr: статьи по разработке и DevOps", "devtools"),
]


def embed(text: str) -> list:
    req = urllib.request.Request(
        NLP,
        data=json.dumps({"text": text}).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=60) as r:
        return json.load(r)["embedding"]


def main():
    notes, embeddings = [], []
    for title, cluster in NOTES:
        nid = str(uuid.uuid4())
        vec = embed(title)
        notes.append({"id": nid, "title": title, "folder": "", "cluster": cluster})
        embeddings.append({"note_id": nid, "model": "current@128", "vec": "[" + ",".join(map(repr, vec)) + "]"})
        print(f"{cluster:10s} {title[:50]}")
    out = {"notes": notes, "embeddings": embeddings, "keywords": [], "links": []}
    with open("work-w1/synth.json", "w", encoding="utf-8") as f:
        json.dump(out, f, ensure_ascii=False, indent=1)
    print(f"\n{len(notes)} synthetic notes -> work-w1/synth.json")


if __name__ == "__main__":
    main()
