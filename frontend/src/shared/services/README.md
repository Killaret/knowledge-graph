# PreloadService - Фоновая предзагрузка данных

PreloadService - это сервис для фоновой загрузки публичных данных пока пользователь не аутентифицирован, что обеспечивает мгновенное отображение интерфейса после входа.

## Основной принцип

1. **На странице входа**: Запускается фоновая загрузка публичных данных (граф, достижения)
2. **После входа**: Предзагруженные данные немедленно отображаются, затем обновляются персональными личными данными через `/api/v1/me/graph/fresh` и `delta`
3. **При выходе**: Кэш очищается для безопасности

## Использование

### Базовое использование в компонентах

```typescript
import { getGraphWithPreload, getAchievementsWithPreload } from "$shared/hooks/usePreloadedData";

// Получение графа с приоритетом на предзагруженные данные
const graphData = await getGraphWithPreload(1000);

// Получение достижений
const achievements = await getAchievementsWithPreload(); // публичные
const personalAchievements = await getAchievementsWithPreload(true); // персональные
```

### Мгновенное отображение после входа

```typescript
import { useInstantData } from "$shared/hooks/usePreloadedData";

// Получить предзагруженные данные мгновенно
const instantData = useInstantData();

if (instantData.hasInstantData) {
  // Показать UI немедленно с предзагруженными данными
  graphData = instantData.graph;
  achievements = instantData.achievements;
}
```

### Комбинированная загрузка данных приложения

```typescript
import { loadAppData } from "$shared/hooks/usePreloadedData";

const appData = await loadAppData({
  limit: 1000,
  usePersonalAchievements: false,
  fallbackToServer: true,
});

console.log("Used preloaded data:", appData.usedPreloaded);
```

## API Reference

### PreloadService

Основной класс сервиса (синглтон):

```typescript
import { PreloadService } from "$shared/services/PreloadService";

// Запуск фоновой предзагрузки
await PreloadService.startPreload();

// Получение предзагруженных данных
const graph = PreloadService.getPreloadedGraph();
const achievements = PreloadService.getPreloadedAchievements();

// Управление кэшем
PreloadService.clearCache();
PreloadService.invalidateGraphCache();
PreloadService.invalidateAchievementsCache();

// Статус
const isPreloading = PreloadService.isPreloadingData();
const hasData = PreloadService.hasPreloadedData();
const stats = PreloadService.getStats();
```

### Удобные функции

```typescript
import {
  startPreload,
  getPreloadedGraph,
  getPreloadedAchievements,
  clearPreloadCache,
  hasPreloadedData,
} from "$shared/services/PreloadService";
```

## Конфигурация

### TTL (Time To Live)

- **Граф**: 5 минут (300,000 мс)
- **Достижения**: 10 минут (600,000 мс)

### Ограничения

- Граф загружается с лимитом 1000 узлов для производительности
- Только публичные данные предзагружаются (без аутентификации)

## Интеграция с приложением

### 1. Layout (+layout.svelte)

```typescript
import { startPreload } from "$shared/services/PreloadService";

$effect(() => {
  initAuth();

  // Запуск предзагрузки если не аутентифицирован
  if (!isAuthenticated()) {
    startPreload();
  }
});
```

### 2. Auth Store

```typescript
import { clearPreloadCache } from "$shared/services/PreloadService";

export async function logout(): Promise<void> {
  // ... логика выхода ...
  clearPreloadCache(); // Очистка кэша
  // ... редирект ...
}
```

### 3. Главная страница (+page.svelte)

```typescript
import { getGraphWithPreload, useInstantData } from "$shared/hooks/usePreloadedData";

async function loadDataParallel() {
  // Мгновенное получение предзагруженных данных
  const instantData = useInstantData();

  if (instantData.hasInstantData) {
    graphData = instantData.graph; // Немедленное отображение
  }

  // Загрузка свежих данных параллельно
  const freshData = await getGraphWithPreload();
  if (freshData !== graphData) {
    graphData = freshData; // Обновление если данные изменились
  }
}
```

## Мониторинг и отладка

### Логирование

Сервис использует детальное логирование для отладки:

```
[PreloadService] Starting background preload...
[PreloadService] Graph preloaded successfully
[PreloadService] Achievements preloaded successfully
[usePreloadedData] Using preloaded graph data
```

### Статистика

```typescript
const stats = PreloadService.getStats();
console.log("Preload stats:", stats);
/*
{
  hasGraph: true,
  hasAchievements: true,
  graphAge: 123456, // возраст в мс
  achievementsAge: 234567,
  isPreloading: false
}
*/
```

## Производительность

### Преимущества

- **Мгновенный UI**: Данные отображаются сразу после входа
- **Параллельная загрузка**: Граф и достижения загружаются одновременно
- **Интеллектуальное кэширование**: TTL предотвращает устаревание данных
- **Безопасность**: Кэш очищается при выходе

### Накладные расходы

- Дополнительный сетевой запрос на странице входа
- Память для кэша (несколько МБ для графа)
- Фоновая обработка

## Best Practices

1. **Используйте хуки**: Предпочитайте `usePreloadedData` хуки прямому вызову PreloadService
2. **Проверяйте актуальность**: Всегда проверяйте `hasInstantData` перед использованием
3. **Обрабатывайте ошибки**: Предзагрузка может завершиться с ошибкой
4. **Тестируйте**: Убедитесь что UI работает как с предзагруженными, так и с свежими данными
5. **Мониторьте**: Используйте `getStats()` для отладки производительности

## Пример полного цикла

1. **Пользователь заходит на страницу входа**
   - Запускается `startPreload()`
   - Фоново загружаются граф и достижения

2. **Пользователь вводит логин/пароль**
   - Выполняется вход
   - Предзагруженные данные немедленно отображаются

3. **После входа**
   - UI показывается мгновенно с предзагруженными данными
   - Параллельно загружаются персональные данные
   - При необходимости UI обновляется свежими данными

4. **Выход пользователя**
   - Вызывается `clearPreloadCache()`
   - Все предзагруженные данные удаляются

Это обеспечивает оптимальный пользовательский опыт с минимальными задержками после аутентификации.
