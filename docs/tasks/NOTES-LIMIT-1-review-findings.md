# NOTES-LIMIT-1 — разбор исполнения (Devin, 2026-09-30)

## Что сделано

1. `frontend/src/shared/api/notes.ts` — `getNotes()` читает все страницы: цикл `while (all.length < total)` с `offset = all.length`, стоп на пустой странице. Бэкенд по-прежнему режет `limit` до `pagination.max_limit` (300) — фронт теперь дочитывает `total`, а не молча берёт первую страницу. Оба вызовителя (`home-page.svelte.ts`, `graphLoader.ts`) получают полный массив без изменения сигнатуры.
2. `frontend/src/entities/graph/model/filter-state.ts` — `filterGraphData` при невыбранных фильтрах (`!isTypeActive && !isSearchActive`) возвращает `graphData` как есть: узлы графа больше не выбрасываются только потому, что их заметок нет в списке.

## Тесты и мутация

- `notes.test.ts` — «fetch all pages when total exceeds the first page»: MSW отдаёт 2+1 заметки при `total=3`, проверяет порядок `offset` и итог `[1,2,3]`; «stop when a page comes back empty» — пустая страница при `total=5` завершает цикл.
- `filter-state.test.ts` — «keeps graph nodes not present in the notes list when no filter is active»: 3 узла при 2 заметках в списке — все 3 узла и 2 связи остаются; тест с `selectedType` фильтрует как раньше.
- Мутация: замена цикла на один запрос (`break` после первой страницы) → «fetch all pages…» красный. Восстановлено.
- Прогоны: `vitest notes.test.ts filter-state.test.ts` 40/40; `graphLoader.test.ts` + `home-page` 46/46; `svelte-check` 0 ошибок.

## Что не тронуто

- Лимит бэкенда и контракт `GET /v1/notes` не менялись — постановка указывает на фронтенд-пагинацию.
- UI-пагинации в списке нет — владелец её не заказывал; список получает полный массив, фильтрация/сортировка остаются клиентскими.
