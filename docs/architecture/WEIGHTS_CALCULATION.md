# Расчёт весов связей и рекомендаций

## Прямая схожесть (direct weight) для явных связей

`direct(A,B) = explicit(A,B) × 1.0 + (1 - explicit(A,B)) × (α×content_sim + β×tag_sim + γ×keyword_sim)`

Для MVP используются только явные связи (explicit = 1) или только content_sim (если explicit = 0).

## Распространение весов (BFS)

Для пути длины k:

`score_path = w1 × (λ^(k-1)) × w2 × ... × wk`

- `λ` — коэффициент затухания (`RECOMMENDATION_DECAY`), применяется начиная со второго уровня.
- Для прямой связи (k=1) затухание не применяется.

## Комбинирование с эмбеддингами

Для кандидата C, найденного через семантическое сходство:

`score = β × cosine_sim(embedding_A, embedding_C)`

Для кандидата, достигнутого по явной связи:

`score = α × explicit_weight`

Если кандидат достижим и по явной связи, и через эмбеддинги, веса складываются.

## Нормализация

Косинусное сходство эмбеддингов приводится к диапазону [0,1] по формуле:  
`similarity = clamp(1 - distance, 0, 1)`, где `distance` — косинусное расстояние (оператор `<=>` в pgvector). Реализация: `GREATEST(0.0, LEAST(1.0, 1 - (e1.embedding <=> e2.embedding)))` в `backend/internal/infrastructure/db/postgres/embedding_repo.go`.