# BOOTSTRAP — точка входа для агента на новой машине

Скорми этот файл агенту (Devin, Claude Code, любой другой) — он содержит
полный маршрут развёртывания проекта. Детали намеренно не дублируются:
нормативные источники указаны по ссылкам, агент обязан читать их, а не
действовать по памяти.

## Контракт

Ты — агент, разворачивающий Knowledge Graph на машине, где его ещё нет.
Твоя задача: довести машину до состояния «стек поднят, проверки зелёные»
и отчитаться по чек-листу в конце этого файла. Ничего не пропускай молча —
если шаг невозможен, остановись и спроси.

## Шаг 0. Прочитать нормативку

До любых действий прочитай:

1. `.windsurfrules` — единственный нормативный источник (архитектура,
   запреты, версии инструментов).
2. `docs/agents/AI_AGENT_SETUP.md` — из чего состоит обвязка агентов и
   какие ловушки уже пойманы.
3. `DEPLOY.md` — практический гид по стекам; все команды ниже — выдержки
   из него, при расхождении верен он.

## Шаг 1. Окружение

Проверь и доложи версии (минимумы — в `.windsurfrules`):

```powershell
docker --version            # Docker Desktop с WSL2
docker compose version
git --version
go version                  # Go 1.25+
node --version              # Node LTS
python --version            # обязателен на PATH: без него не работает
                            # PreToolUse-хук guard-personal-data.py
gh --version                # GitHub CLI, для пушей
```

Чего-то нет — остановись, скажи что ставить.

## Шаг 2. Клон и ветка

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
git checkout ai-agents
```

Рабочая ветка для агентов — `ai-agents`. `main` догоняется мерджем,
не коммить туда напрямую.

## Шаг 3. Конфигурация

```powershell
cp .env.example .env
```

Открой `.env`, заполни обязательное (таблица — в `DEPLOY.md`, раздел
«Setting up .env»): `JWT_SECRET` (32+ символов), `POSTGRES_PASSWORD`,
`PERSONAL_POSTGRES_PASSWORD`, `TEST_POSTGRES_PASSWORD`, `MONGO_URL`,
`MONGO_DATABASE`. Секреты не коммитить и не логировать.

## Шаг 4. Стек

Выбери один — по задаче, не «на всякий случай»:

| Стек | Команда | Когда |
|---|---|---|
| dev | `docker compose up -d --build` | разработка |
| test | `pwsh scripts/testing/start-test.ps1` | изолированные прогоны, BDD |
| personal | `docker compose -f docker-compose.personal.yml up -d --build` | **только по явному запросу владельца** — там личные данные |

Для personal-стека NLP-модель качается один раз (команда в `DEPLOY.md`,
Quick Start). Порты и URL после старта — там же.

## Шаг 5. Верификация

```powershell
docker compose ps                     # все сервисы healthy
pwsh scripts/testing/check-all.ps1    # полный прогон; -Quick для быстрой проверки
```

`check-all` обязан дойти до конца. Единственный допустимый SKIP —
`golangci-lint`, если не установлен.

## Чек-лист отчёта

- [ ] Версии окружения соответствуют `.windsurfrules`
- [ ] `.env` заполнен, секреты не попали в git
- [ ] `docker compose ps` — все сервисы healthy
- [ ] `check-all` — все фазы PASS (golangci-lint SKIP допустим)
- [ ] Стек personal не поднимался без явного запроса

## Ловушки

- **Personal-стек** не стартуется без явного запроса владельца — там
  личные данные, см. `docs/agents/AI_AGENT_SETUP.md`.
- **Python на PATH** — обязателен даже если не собираешься трогать
  NLP: хук `guard-personal-data.py` висит на каждой команде.
- **Хук должен работать, а не отсутствовать**: проверь, что запрещённая
  команда реально останавливается (`AI_AGENT_SETUP.md`, раздел
  «Развернуть этот проект»).
- Порты, конфликты, WSL2-грабли, журнал типовых сообщений —
  `DEPLOY.md`, разделы Troubleshooting и «Windows pitfalls».
