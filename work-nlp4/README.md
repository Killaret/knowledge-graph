# work-nlp4 — замер итеративной нормализации для NLP-4

Цель: эмпирически закрыть открытый вопрос спеки `docs/tasks/NLP-4-note-logical-form-normalization.md` —
нужна ли итеративная нормализация и какой критерий остановки брать.

## Харнес

`nlp-service/scripts/measure_normalization.py` — 4 прохода над датасетом, считает:

- токены/символы на итерацию, долю неизменившихся заметок;
- распределение итерации остановки при гипотезе «бюджет 3 + дельта токенов < 5%»;
- деградации (вывод < 100 символов → откат);
- при `--with-embeddings`: косинусную близость эмбеддинга к исходному тексту (e5-base, внутри контейнера NLP).

Формат датасета: `[{"id": "...", "title": "...", "content": "..."}]`.

## Проверка харнеса (сделано)

- `fixture_boilerplate.json` (4 синтетические заметки с бойлерплейтом): итерация 1 чистит,
  итерации 2+ — неподвижная точка; «Tiny after strip» ловится деградацией (<100 символов → откат).
- `seed_dataset.json` (100 заметок тест-сида): 0 изменений на всех итерациях, cos = 1.0 —
  чистый текст правила не ломают.

## Чего не хватает — корпус владельца

Тела 113 заметок из work-w1/work-nlp2 не сохранились (только счётчики). Экспорт Personal-стека:

```bash
docker compose -f docker-compose.personal.yml up -d postgres
docker exec kg-postgres-personal psql -U personal -d knowledge_personal -t -A \
  -c "SELECT json_agg(json_build_object('id',id,'title',title,'content',content)) FROM notes WHERE length(content) >= 100" \
  > work-nlp4/notes_dataset.json
```

Затем:

```bash
python nlp-service/scripts/measure_normalization.py --dataset work-nlp4/notes_dataset.json \
  --out work-nlp4/measurements.md --raw work-nlp4/raw_notes.json
# ветка с эмбеддингами — внутри kg-test-nlp (модель уже в кэше)
```

Запуск Personal-стека — только по явному разрешению владельца.
