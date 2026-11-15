# I18N-1 — разбор ревью

## Ревью — Claude Code, 2026-11-15 (ревью №9)

Работа: Devin, `ced4922`. **Вердикт: принято без замечаний.**

Три ключа — `field.password`, `graphOverlay.fogWarning`, `graphOverlay.fogRecovery` — есть в обоих языках.
Тест паритета `i18n.test.ts`, «locale parity», сравнивает полные наборы ключей `messages.en` и `messages.ru` в обе
стороны. На текущей ветке он зелёный: ключи, добавленные после 01.10, тоже совпадают.

### Мутации — на `9255429`

| Мутация | Результат |
|---|---|
| `field.password` убран из en (`messages/common.ts`) | красная: «en and ru expose exactly the same keys» |
| `graphOverlay.fogWarning` убран из en (`messages/graph.ts`) | красная: паритет и `overlay.test.ts` «shows fog danger warning» |

Находка Devin верна: два теста оверлея искали `/fog/i` и проходили на сыром ключе `graphOverlay.fog…`. Теперь они
ищут текст сообщения, и вторая мутация это подтверждает.

### Прогон

`check-all.ps1` без `-Quick` на `a2285dd` (2026-11-15): 34 PASS, 2 FAIL, 1 SKIP. Красные — вне задачи: *Frontend unit* — `home-page.svelte.test.ts` «switches views» (FREEZE-3D-1, ревью №8, отклонено); *NLP* — `test_long_paragraph_is_chunked_quickly[prose-400K, www-400K]`, порог скорости, на этой машине краснеет и отдельно (см. ревью PANEL-LINKS-1). Пропущен `golangci-lint` — локально не установлен. *Graph-service integration* зелёная.
