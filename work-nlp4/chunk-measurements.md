## Чанкер: статистика резки (108 заметок, target=96, max=128, model=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2)
- чанков/заметку: med 8, mean 8.0, max 20
- токенов/чанк: med 81, p90 109, max 128
- forced_split: 128 чанков
- виды блоков: {'prose': 869}

### Doc-вектор vs текущий усечённый /embed
- cos(mean_agg, truncated): med 0.744, min 0.362 — доля смысла, которую теряет усечение
- cos(weighted_agg, truncated): med 0.768, min 0.366
- cos(doc с title-инъекцией, doc без): med 0.911, min 0.749

### Соседи top-2 (прокси gamma-связей): усечение vs чанки
- top-2 score: truncated med 0.657 → chunked med 0.654
- заметок, у которых сменился хотя бы один top-2 сосед: 81/99 (82%)

### Хвост дрейфа (худшие 5 по mean_agg)
- 0.362 | note-065 | 15 чанков
- 0.425 | note-074 | 17 чанков
- 0.514 | note-108 | 16 чанков
- 0.532 | note-050 | 14 чанков
- 0.570 | note-014 | 16 чанков

## Чанкер: статистика резки (108 заметок, target=256, max=512, model=intfloat/multilingual-e5-base)
- чанков/заметку: med 3, mean 3.2, max 8
- токенов/чанк: med 231, p90 254, max 507
- forced_split: 2 чанков
- виды блоков: {'prose': 351}

### Doc-вектор vs текущий усечённый /embed
- cos(mean_agg, truncated): med 0.963, min 0.925 — доля смысла, которую теряет усечение
- cos(weighted_agg, truncated): med 0.967, min 0.931
- cos(doc с title-инъекцией, doc без): med 0.986, min 0.946

### Соседи top-2 (прокси gamma-связей): усечение vs чанки
- top-2 score: truncated med 0.880 → chunked med 0.921
- заметок, у которых сменился хотя бы один top-2 сосед: 67/99 (68%)

### Title-инъекция: влияние на top-2
- сменился top-2 сосед: 46/99 (46%)
- top-2 score: без title med 0.921 → с title med 0.897

### Хвост дрейфа (худшие 5 по mean_agg)
- 0.925 | note-017 | 6 чанков
- 0.933 | note-056 | 1 чанков
- 0.934 | note-030 | 2 чанков
- 0.942 | note-018 | 1 чанков
- 0.942 | note-049 | 1 чанков
