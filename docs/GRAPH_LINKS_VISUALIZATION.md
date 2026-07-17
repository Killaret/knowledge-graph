# Визуализация связей графа

## 1. Типы связей

| Тип | Назначение | Цвет линии | Стиль линии |
|-----|------------|------------|-------------|
| **reference** | Прямая ссылка между заметками (по умолчанию) | Синий `#3366ff` | Сплошная `[]` |
| **dependency** | Зависимость (одна заметка опирается на другую) | Оранжевый `#ff6600` | Штрих-пунктир `[10, 3]` |
| **related** | Связанная тема (слабая связь) | Серый `#999999` | Сплошная (при весе < 0.3 — пунктир `[6, 4]`) |
| **custom** | Пользовательская связь | Розовый `#ff66ff` | Точечный пунктир `[2, 6]` |

## 2. Расчёт веса связи

**Вес связи (weight)** — число от `0.0` до `1.0`, отражающее "силу притяжения" между двумя заметками.

### Пользовательские связи
- Задаётся вручную при создании через `LinkCreator`
- **По умолчанию:** `0.5`
- Сохраняется в таблице `links`
- **Не пересчитывается** автоматически

### Рекомендации (автоматические связи)
Вес пересчитывается через `RefreshService`:

```text
final_weight = α × graph_weight + β × semantic_weight + γ × keyword_weight
```

| Компонент | Коэффициент | Описание |
|-----------|-------------|----------|
| **graph_weight** | α (alpha) = 0.5 | Близость по графу явных связей (BFS-обход, decay=0.5) |
| **semantic_weight** | β (beta) = 0.5 | Косинусное сходство векторных эмбеддингов (pgvector) |
| **keyword_weight** | γ (gamma) = 0.2 | Сходство по ключевым словам (Jaccard Similarity) |

**Важно:** Комбинированный вес используется **только для рекомендаций**. Явная связь сохраняет вес, заданный пользователем.

## 3. Отображение веса на графе

### Толщина линии
```typescript
const lineWidth = Math.max(1, weight * 4);
// Без множителей для типов — только от веса
```

| Вес | Толщина (px) | Визуально |
|-----|--------------|-----------|
| 0.1 | 1.0 | Тонкая |
| 0.5 | 2.0 | Средняя |
| 0.8 | 3.2 | Толстая |
| 1.0 | 4.0 | Очень толстая |

### Прозрачность (opacity)
```typescript
const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
// weight=0.1 → opacity=0.44
// weight=0.5 → opacity=0.60
// weight=0.8 → opacity=0.72
// weight=1.0 → opacity=0.80
```

**Прямые связи (direct links)** — созданы пользователем явно:
- Отображаются **сплошной линией** (если тип `reference` или `related` с весом ≥ 0.3)
- Вес — точное значение, заданное пользователем
- При наведении — тултип с типом связи (планируется)

## 4. Прямые vs Рекомендации в UI

| Характеристика | Прямые связи | Рекомендации |
|----------------|--------------|--------------|
| **Источник** | Таблица `links` | Таблица `note_recommendations` |
| **Создание** | Вручную через UI/API | Автоматически worker'ом |
| **Тип** | Явный (reference/dependency/related/custom) | Вычисляемый (обычно related) |
| **Вес** | Задаётся пользователем (0.0-1.0) | Вычисляется (α×Graph + β×Semantic + γ×Keyword) |
| **Отображение** | Яркие цвета | Бледнее (fadeOpacity) |
| **Стиль** | По типу связи | По весу (слабые = пунктир) |

## 5. Код рендеринга

Основной файл: [`frontend/src/components/organisms/GraphCanvas/renderer.ts`](../frontend/src/components/organisms/GraphCanvas/renderer.ts)

### Функция getLinkColor
```typescript
function getLinkColor(weight: number, linkType?: string, fadeOpacity: number = 1): string {
  const effectiveType = linkType || 'related';
  const color = linkTypeColors[effectiveType] || linkTypeColors['related'];
  const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
  const finalOpacity = baseOpacity * fadeOpacity;

  // Convert hex to rgba
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${finalOpacity})`;
}
```

### Функция drawLink
```typescript
export function drawLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  sourceNode: SimulationNode,
  targetNode: SimulationNode,
  opacity: number = 1
): void {
  ctx.beginPath();
  ctx.moveTo(sourceNode.x!, sourceNode.y!);
  ctx.lineTo(targetNode.x!, targetNode.y!);

  const weight = link.weight ?? 0.5;
  const linkType = link.link_type;

  // Line thickness per specification: Math.max(1, weight * 4)
  const lineWidth = Math.max(1, weight * 4);

  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = getLinkColor(weight, linkType, opacity);

  // Dash pattern
  const dash = getLineDash(linkType, weight);
  if (dash.length > 0) {
    ctx.setLineDash(dash);
  } else {
    ctx.setLineDash([]);
  }

  ctx.stroke();
  ctx.setLineDash([]);
}
```

## 6. Где это в коде

| Компонент | Файл | Описание |
|-----------|------|----------|
| **Типы и цвета** | `frontend/src/components/organisms/GraphCanvas/renderer.ts` | Функции `getLinkColor`, `getLineDash`, `drawLink` |
| **Расчёт комбинированного веса** | `backend/internal/application/recommendation/refresh_service.go` | `RefreshService.RefreshRecommendations` |
| **Jaccard Similarity** | `backend/internal/application/recommendation/keyword_similarity.go` | `CalculateKeywordSimilarity` |
| **BFS-обход** | `backend/internal/domain/graph/bfs.go` | `runBFS` |
| **Конфигурация α/β/γ** | `knowledge-graph.config.json` → `backend.recommendation` | `alpha`, `beta`, `gamma` |

## 7. Тесты

| Тип тестов | Файл |
|------------|------|
| **Визуальные тесты** | `frontend/src/components/organisms/GraphCanvas.rendering.spec.ts` |
| **Детальные тесты** | `frontend/src/components/organisms/GraphCanvas.links-detailed.spec.ts` |
| **Тесты соединений** | `frontend/src/components/organisms/GraphCanvas.links.connection.spec.ts` |

## 8. Безопасность прокси (hooks.server.ts)

Прокси SvelteKit **блокирует чувствительные заголовки** для безопасности:

### ✅ Разрешённые заголовки
```typescript
const allowedHeaders = ['accept', 'content-type', 'x-request-id'];
```

### ❌ Заблокированные заголовки
```typescript
const blockedHeaders = [
  'cookie',              // Сессионные cookies
  'authorization',       // JWT токены
  'connection',          // Hop-by-hop
  'proxy-',              // Proxy headers
  'transfer-encoding',   // Hop-by-hop
  'keep-alive',          // Connection management
  'upgrade',             // Protocol upgrade
  'te',                  // Trailer encoding
  'host'                 // Host header (передаётся отдельно для graph-service)
];
```

**Почему?**
- **cookie** и **authorization** управляются **client-side** через API клиент (`frontend/src/shared/api/client.ts`)
- Прокси не должен передавать чувствительные данные между сервисами
- Предотвращает утечку токенов при компрометации одного из сервисов

**Где:** `frontend/src/hooks.server.ts`

### backendApiProxy
- Проксирует `/api/v1/*` → `http://backend_personal:8080`
- Блокирует `cookie`, `authorization`
- Передаёт только: `accept`, `content-type`, `x-request-id`

### graphServiceProxy
- Проксирует `/graph-service/*` → `http://graph-service:9091`
- Блокирует `cookie`, `authorization`
- Добавляет `x-internal-auth` (если настроен)
- Передаёт `host` для правильного проксирования

## 9. Ссылки

- [Recommendation Architecture](RECOMMENDATION_ARCHITECTURE.md) — архитектура системы рекомендаций
- [Recommendation Troubleshooting](RECOMMENDATION_TROUBLESHOOTING.md) — диагностика проблем
- [Knowledge Graph Config](../knowledge-graph.config.json) — конфигурация весов α, β, γ
- [Backend Patterns](../backend/BACKEND_PATTERNS.md) — паттерны бэкенда
