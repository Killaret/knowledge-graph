# Frontend Rules — Knowledge Graph (Коротко)

Краткий свод практик, не столь фундаментальных, как `knowledge-graph.config.json`, но важных для единого стиля разработки.

1. Конфигурация

- Использовать `src/shared/config/config.ts` и `knowledge-graph.config.json` как источник истины.

2. Компоненты

- Компоненты небольшие, переиспользуемые, тестируемые; избегать больших «монстров».
- Использовать `Modal.svelte` как базу для всех диалогов.

3. State и stores

- Глобальное состояние в `src/shared/stores/*`.
- Сервисную логику выносить в `src/shared/services/*`.

4. API

- Все HTTP-вызовы через `src/shared/api/*`.
- В тестах мокировать HTTP через MSW, не делать реальные HTTP в unit-тестах.

5. Тестирование

- Unit: `vitest` + `@testing-library/svelte`.
- E2E: `playwright` с `playwright.config.*.ts`.
- Для стабильности тестов: использовать `vitest-setup.ts` mocks (animate, RAF, canvas, ResizeObserver).

6. Визуальные тесты и анимации

- Отключать/мокировать анимации в тестовом режиме.
- Скриншоты сохранять в `frontend/test-results/temp/screenshots/`.

7. Accessibility

- Проверять aria-атрибуты и клавиатурную навигацию для новых компонентов.

8. Styling

- Темизация через CSS-переменные, поддержка `galactic` messaging.

9. CI и артефакты

- CI очищает только `frontend/test-results/temp/`.
- `frontend/test-results/baseline/` версионируется вручную и не очищается автоматом.

10. Малые правила кодирования

- Имя файлов: PascalCase для Svelte-компонентов (SearchBar.svelte).
- Тесты colocated: `Component.spec.ts` рядом с компонентом.
- Использовать `src/shared/test-utils` для общих хелперов.

11. Domain Value Objects

- Визуальные/доменные параметры celestial-узлов (`CelestialBody`) хранятся в `src/shared/lib/domain/`.
- Запрещено дублировать цвета, emoji, label или switch-ветви по типам в компонентах/рендерерах — использовать `CelestialBody.fromString()` и `CelestialBody.UI_TYPES`.

---

## 🌐 Language Policy

**All user-facing content MUST be in English:**

- ✅ **Note titles and content** — all user-created notes
- ✅ **Annotations and descriptions** — any text fields for users
- ✅ **UI strings** — buttons, labels, placeholders, error messages
- ✅ **Toast messages and tooltips** — GalacticLexicon messages
- ✅ **Comments in code** — public API docs, README files

**Exceptions:**

- Internal code comments — brief explanations in any language OK

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```
