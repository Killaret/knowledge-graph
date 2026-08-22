# Шпаргалка: Связи в Knowledge Graph

## 🎨 Типы связей (актуальные цвета)

| Тип | Цвет | Стиль | Weight multiplier |
|-----|------|-------|-------------------|
| `reference` | `#3366ff` (синий) | Сплошная `[]` | ×1.0 |
| `dependency` | `#ff6600` (оранжевый) | `[10, 3]` | ×1.0 |
| `related` | `#999999` (серый) | Сплошная или `[6, 4]` если weight<0.3 | ×1.0 |
| `custom` | `#ff66ff` (розовый) | `[2, 6]` | ×1.0 |

**Файл:** `frontend/src/components/organisms/GraphCanvas/renderer.ts`

---

## ⚖️ Расчёт веса

### Пользовательские связи
```text
weight = задаётся вручную (по умолчанию 0.5)
```

### Рекомендации (автоматические)
```text
final_weight = α×graph + β×semantic + γ×keyword
```

| Параметр | Значение | Описание |
|----------|----------|----------|
| **α (alpha)** | 0.5 | BFS-обход графа (decay=0.5) |
| **β (beta)** | 0.5 | Косинусное сходство (pgvector) |
| **γ (gamma)** | 0.2 | Jaccard similarity (ключевые слова) |

**Файл:** `knowledge-graph.config.json` → `backend.recommendation`

---

## 📏 Визуализация

### Толщина линии
```typescript
lineWidth = max(1, weight × 4)
// Без множителей для типов
```

| weight | толщина (px) |
|--------|--------------|
| 0.1 | 1.0 |
| 0.5 | 2.0 |
| 0.8 | 3.2 |
| 1.0 | 4.0 |

### Прозрачность
```typescript
opacity = 0.4 + weight × 0.4
```

| weight | opacity |
|--------|---------|
| 0.1 | 0.44 |
| 0.5 | 0.60 |
| 0.8 | 0.72 |
| 1.0 | 0.80 |

---

## 🔐 Безопасность прокси

### ✅ Разрешено
```typescript
['accept', 'content-type', 'x-request-id']
```

### ❌ Заблокировано
```typescript
['cookie', 'authorization', 'connection', 'proxy-',
 'transfer-encoding', 'keep-alive', 'upgrade', 'te', 'host']
```

**Почему:** Cookie и токены управляются client-side API клиентом.

**Файл:** `frontend/src/hooks.server.ts`

---

## 📁 Файлы

| Компонент | Файл |
|-----------|------|
| **Рендеринг** | `frontend/src/components/organisms/GraphCanvas/renderer.ts` |
| **Расчёт веса** | `backend/internal/application/recommendation/refresh_service.go` |
| **Jaccard** | `backend/internal/application/recommendation/keyword_similarity.go` |
| **BFS** | `backend/internal/domain/graph/bfs.go` |
| **Конфиг** | `knowledge-graph.config.json` |
| **Прокси** | `frontend/src/hooks.server.ts` |

---

## 🧪 Тесты

```bash
npm run test:unit -- GraphCanvas.links
```

| Файл | Описание |
|------|----------|
| `GraphCanvas.links.spec.ts` | Базовые тесты |
| `GraphCanvas.rendering.spec.ts` | Визуальные тесты |
| `GraphCanvas.links.connection.spec.ts` | Тесты соединений |
| `GraphCanvas.links-detailed.spec.ts` | Детальные тесты |

---

## 🔗 Ссылки

- [Полная документация](GRAPH_LINKS_VISUALIZATION.md)
- [Recommendation Architecture](RECOMMENDATION_ARCHITECTURE.md)
- [Recommendation Troubleshooting](RECOMMENDATION_TROUBLESHOOTING.md)
