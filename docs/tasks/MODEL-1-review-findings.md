# MODEL-1. Результаты замера моделей эмбеддингов

Данные: 113 реальных заметок владельца (экспорт Personal-базы 2026-09-18, только чтение). Замер внутри контейнера образа `knowledge-graph-nlp-personal`, CPU. Команда воспроизведения — в конце файла.

> Отступление от постановки: вместо «22 закладки + тест-сид» взят полный экспорт Personal-базы — строгое надмножество (все исходные закладки входят), корпус чище после нормализации IMP-6/7/8 и ближе к реальной нагрузке. Personal-стек не изменялся, запросы только на чтение.
>
> Оговорка по интерпретации: `current@128` видит только ~80–100 первых слов каждой заметки (усечение `max_seq_length`), у остальных вариантов окно ≥512. Различия в выдаче — смесь «модель» и «сколько текста она прочитала».

## Таблица 0. Параметры моделей (сверено по файлам)

| Вариант | hidden | max_position | окно в файле | окно замера | лицензия |
|---|---|---|---|---|---|
| current@128 | 384 | 512 | 128 | 128 | Apache-2.0 |
| current@512 | 384 | 512 | 128 | 512 | Apache-2.0 |
| e5-small | 384 | 512 | 512 | 512 | MIT |
| rubert-tiny2 | 312 | 2048 | 2048 | 2048 | MIT |
| e5-base | 768 | 514 | 512 | 512 | MIT |

## Таблица 1. Поиск — top-10 по каждому запросу

### Запрос 1: «книги по архитектуре и микросервисам»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | Figma: микросервисы (дизайн-файл) | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | Figma: микросервисы (дизайн-файл) | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) |
| 2 | Курс Архитектор решений (Solution architect) в «Специалист» | Coursera: Software Engineering — Software Design and Project | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | Coursera: Software Engineering — Software Design and Project | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf |
| 3 | Чистая архитектура. Искусство разработки программного обеспе | Сергей Константинов. API | «Эволюционная архитектура. Поддержка непрерывных изменений ( | Coursera: Modeling Software Systems Using UML | Чистая архитектура. Искусство разработки программного обеспе |
| 4 | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: | Личный кабинет PurpleSchool | Чистая архитектура. Искусство разработки программного обеспе | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | «Эволюционная архитектура. Поддержка непрерывных изменений ( |
| 5 | Figma: микросервисы (дизайн-файл) | Knowledge Core | Курс Архитектор решений (Solution architect) в «Специалист» | GETANALYST | DDD и Event Storming - архитектура для системно | Figma: микросервисы (дизайн-файл) |
| 6 | «Эволюционная архитектура. Поддержка непрерывных изменений ( | EnglishSpace: мнемотехника для английского | «Release it! Проектирование и дизайн ПО для тех, кому не всё | HTX — криптобиржа | «Release it! Проектирование и дизайн ПО для тех, кому не всё |
| 7 | Coursera: Software Engineering — Software Design and Project | Coursera: Software Processes | Figma: микросервисы (дизайн-файл) | Чистая архитектура. Искусство разработки программного обеспе | Курс Архитектор решений (Solution architect) в «Специалист» |
| 8 | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | SHAREWOOD.TECH (ex .BIZ) (Шервуд) 🔥 Слив Курсов – Скачать Бе | Хекслет: урок «Структуры» (Go Basics) | Курс Системный дизайн - обучение проектированию систем - Кар |
| 9 | «Release it! Проектирование и дизайн ПО для тех, кому не всё | Coursera: Modeling Software Systems Using UML | Скачать книгу «Изучаем PostgreSQL 10» [id:655852] в формате | «Release it! Проектирование и дизайн ПО для тех, кому не всё | Сергей Константинов. API |
| 10 | Курс Системный дизайн - обучение проектированию систем - Кар | 100 форекс книг | Курс Системный дизайн - обучение проектированию систем - Кар | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | GETANALYST | DDD и Event Storming - архитектура для системно |

### Запрос 2: «английский язык грамматика и времена»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | EnglishSpace: мнемотехника для английского | WooordHunt (Вордхант) — ваш помощник в мире английского язык | Изучение английского языка самостоятельно: бесплатные матери | EnglishSpace: мнемотехника для английского | Изучение английского языка самостоятельно: бесплатные матери |
| 2 | 4ege.ru: таблица 12 времён английского | EnglishSpace: мнемотехника для английского | 100 статей за 3 месяца: авторский план изучения грамматики с | WooordHunt (Вордхант) — ваш помощник в мире английского язык | 4ege.ru: таблица 12 времён английского |
| 3 | WooordHunt (Вордхант) — ваш помощник в мире английского язык | 100 статей за 3 месяца: авторский план изучения грамматики с | 4ege.ru: таблица 12 времён английского | 100 статей за 3 месяца: авторский план изучения грамматики с | 100 статей за 3 месяца: авторский план изучения грамматики с |
| 4 | 100 статей за 3 месяца: авторский план изучения грамматики с | 4ege.ru: таблица 12 времён английского | Степени сравнения английских прилагательных: правила degrees | 4ege.ru: таблица 12 времён английского | Степени сравнения английских прилагательных: правила degrees |
| 5 | Коллекция Перемещение во времени, перерождени и другой мир - | 423 Be Supposed To Exercises [21 Online Tests] - Grammarism | EnglishSpace: мнемотехника для английского | Степени сравнения английских прилагательных: правила degrees | WooordHunt (Вордхант) — ваш помощник в мире английского язык |
| 6 | Passive Voice в английском языке: правила, примеры, формула | Фермерская жизнь в ином мире / Isekai Nonbiri Nouka | EnglishClass101 — бесплатные материалы месяца | Изучение английского языка самостоятельно: бесплатные матери | EnglishSpace: мнемотехника для английского |
| 7 | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | Listen in English – Free ESL Listening Practice | Passive Voice в английском языке: правила, примеры, формула | Passive Voice в английском языке: правила, примеры, формула | EnglishClass101 — бесплатные материалы месяца |
| 8 | Изучение английского языка самостоятельно: бесплатные матери | Голосовой блокнот - Speechpad.ru | WooordHunt (Вордхант) — ваш помощник в мире английского язык | Голосовой блокнот - Speechpad.ru | Passive Voice в английском языке: правила, примеры, формула |
| 9 | English Corpora: most widely used online corpora. Billions o | Коллекция Перемещение во времени, перерождени и другой мир - | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: | 100 форекс книг | Listen in English – Free ESL Listening Practice |
| 10 | 423 Be Supposed To Exercises [21 Online Tests] - Grammarism | «Эволюционная архитектура. Поддержка непрерывных изменений ( | English Corpora: most widely used online corpora. Billions o | HTX — криптобиржа | English news and easy articles for students of English |

### Запрос 3: «вакансии и работа в Европе»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Eurojobs.com | Jobs across Europe | EURES — европейский портал вакансий | EURES — европейский портал вакансий | EURES — европейский портал вакансий | EURES — европейский портал вакансий |
| 2 | EURES | Find jobs abroad with languages | Europe Language Jobs | Eurojobs.com | Jobs across Europe | Find jobs abroad with languages | Europe Language Jobs | EURES |
| 3 | EURES — европейский портал вакансий | Плати по всему миру | EURES | EURES | Find jobs abroad with languages | Europe Language Jobs |
| 4 | Find jobs abroad with languages | Europe Language Jobs | Eurojobs.com | Jobs across Europe | Find jobs abroad with languages | Europe Language Jobs | Плати по всему миру | Хабр Карьера: избранное |
| 5 | Jobs That Fit | Resumes That Stand Out | JobLeads | 4ege.ru: таблица 12 времён английского | Изучение английского языка самостоятельно: бесплатные матери | Хабр Карьера: избранное | Eurojobs.com | Jobs across Europe |
| 6 | Плати по всему миру | Чтение Манга Неторопливый фермер в другом мире - Farming Lif | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | Jobs That Fit | Resumes That Stand Out | JobLeads | Jobs That Fit | Resumes That Stand Out | JobLeads |
| 7 | Career Community & Conversations | Glassdoor | EURES | DevTools: как вызвать в браузере консоль - функции и возможн | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: |
| 8 | Google UX Design Professional Certificate | Coursera | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | Курс Архитектор решений (Solution architect) в «Специалист» | Изучение английского языка самостоятельно: бесплатные матери |
| 9 | Курс Архитектор решений (Solution architect) в «Специалист» | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат | Jobs That Fit | Resumes That Stand Out | JobLeads | Eurojobs.com | Jobs across Europe | SHAREWOOD.TECH (ex .BIZ) (Шервуд) 🔥 Слив Курсов – Скачать Бе |
| 10 | WooordHunt (Вордхант) — ваш помощник в мире английского язык | Чистая архитектура. Искусство разработки программного обеспе | Основы PHP для начинающих: бесплатный курс по PHP | HTX — криптобиржа | Курс Архитектор решений (Solution architect) в «Специалист» |

### Запрос 4: «манга которую я читаю»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Восемьдесят шесть — Манга | Восемьдесят шесть — Манга | Читать мангу на русском Леди-Дьявол (Devilman Lady). Нагаи Г | Хоть я и бездарная злодейка — аниме | Читать мангу на русском Леди-Дьявол (Devilman Lady). Нагаи Г |
| 2 | Senkuro — Платформа манги, ранобэ и комиксов | Читать мангу на русском Леди-Дьявол (Devilman Lady). Нагаи Г | Манга «Став эволюционирующим космическим монстром» — глава | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит | Восемьдесят шесть — Манга |
| 3 | Читать мангу на русском Леди-Дьявол (Devilman Lady). Нагаи Г | Хоть я и бездарная злодейка — аниме | Чтение Манга Неторопливый фермер в другом мире - Farming Lif | Я влюбился в тебя, когда ты бежала в лунной ночи — аниме | Senkuro — Платформа манги, ранобэ и комиксов |
| 4 | Хоть я и бездарная злодейка — аниме | Операция «Панда» (2024) | Senkuro — Платформа манги, ранобэ и комиксов | EnglishSpace: мнемотехника для английского | Чтение Манга Неторопливый фермер в другом мире - Farming Lif |
| 5 | Операция «Панда» (2024) | Необъятный океан 2 / Grand Blue Season 2 | Хоть я и бездарная злодейка — аниме | 100 форекс книг | Коллекция Перемещение во времени, перерождени и другой мир - |
| 6 | Я влюбился в тебя, когда ты бежала в лунной ночи — аниме | Я влюбился в тебя, когда ты бежала в лунной ночи — аниме | Я влюбился в тебя, когда ты бежала в лунной ночи — аниме | Восемьдесят шесть — Манга | Манга «Став эволюционирующим космическим монстром» — глава |
| 7 | Необъятный океан 2 / Grand Blue Season 2 | Чтение Манга Неторопливый фермер в другом мире - Farming Lif | Кот и дракон — аниме | «Эволюционная архитектура. Поддержка непрерывных изменений ( | Читать Мачеха и ее подруги! |
| 8 | Чтение Манга Неторопливый фермер в другом мире - Farming Lif | Ты и я — полные противоположности, 2 сезон — аниме | Коллекция Перемещение во времени, перерождени и другой мир - | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат | Nagi no Asu kara / Аниме |
| 9 | Nagi no Asu kara / Аниме | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | Восемьдесят шесть — Манга | Читать мангу на русском Леди-Дьявол (Devilman Lady). Нагаи Г | Хоть я и бездарная злодейка — аниме |
| 10 | Читать Мачеха и ее подруги! | Чёрный факел — аниме | Ты и я — полные противоположности, 2 сезон — аниме | Книга "Узнай лжеца по выражению лица" — купить в интернет-ма | Кот и дракон — аниме |

### Запрос 5: «keycloak openid authentication»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Final: OpenID Connect Dynamic Client Registration 1.0 incorp | LetsExchange — крипто-обменник | Keycloak Admin REST API | SafelyChange | Keycloak Admin REST API |
| 2 | Keycloak Admin REST API | SafelyChange | Auth Code Flow pt. 2 | OneLogin Developers | Stacks: sBTC Dual Stacking | Final: OpenID Connect Dynamic Client Registration 1.0 incorp |
| 3 | Запуск Keycloak и подключение LDAP каталога - ViHelp | SecurityLab: эксплуатация уязвимостей IDOR | Auth Code Flow + PKCE | OneLogin Developers | LetsExchange — крипто-обменник | Auth Code Flow + PKCE | OneLogin Developers |
| 4 | LetsExchange — крипто-обменник | Knowledge Core | Final: OpenID Connect Dynamic Client Registration 1.0 incorp | HTX — криптобиржа | Запуск Keycloak и подключение LDAP каталога - ViHelp |
| 5 | Stake with Lido | Lido | HTX — криптобиржа | Запуск Keycloak и подключение LDAP каталога - ViHelp | JSON Web Tokens - jwt.io | Auth Code Flow pt. 2 | OneLogin Developers |
| 6 | ТОП-10 ошибок в спецификации OpenAPI и как их избежать | Сергей Константинов. API | JSON Web Tokens - jwt.io | Auth Code Flow pt. 2 | OneLogin Developers | API Design Patterns - JJ Geewax |
| 7 | SafelyChange | Stacks: sBTC Dual Stacking | JSON Web Token Introduction - jwt.io | SecurityLab: эксплуатация уязвимостей IDOR | Jobs That Fit | Resumes That Stand Out | JobLeads |
| 8 | SecurityLab: эксплуатация уязвимостей IDOR | ТОП-10 ошибок в спецификации OpenAPI и как их избежать | API Design Patterns - JJ Geewax | Аудиокнига Первый игрок скачать торрент бесплатно mp3 | JSON Web Tokens - jwt.io |
| 9 | Knowledge Core | EnglishSpace: мнемотехника для английского | English Corpora: most widely used online corpora. Billions o | Keycloak Admin REST API | Design Patterns | Coursera |
| 10 | Auth Code Flow pt. 2 | OneLogin Developers | Ателье колдовских колпаков / Tongari Boushi no Atelier | Jobs That Fit | Resumes That Stand Out | JobLeads | Postman Documenter: API Docs | English Corpora: most widely used online corpora. Billions o |

### Запрос 6: «coursera software design courses»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Coursera: Software Engineering — Software Design and Project | Coursera: Software Engineering — Software Design and Project | Software Design and Architecture | Coursera | Coursera: Software Engineering — Software Design and Project | Coursera: Software Engineering — Software Design and Project |
| 2 | Coursera: Software Processes | Coursera: Software Processes | Design Patterns | Coursera | Coursera: Modeling Software Systems Using UML | Software Design and Architecture | Coursera |
| 3 | Software Design and Architecture | Coursera | Coursera: Modeling Software Systems Using UML | Coursera: Software Engineering — Software Design and Project | Coursera: Software Processes | Google UX Design Professional Certificate | Coursera |
| 4 | Coursera: Modeling Software Systems Using UML | Хекслет: урок «Требования» (Java Development Overview) | Google UX Design Professional Certificate | Coursera | Software Design and Architecture | Coursera | Design Patterns | Coursera |
| 5 | Design Patterns | Coursera | Курсы Романа Горбачёва | Coursera: Modeling Software Systems Using UML | Хекслет: урок «Требования» (Java Development Overview) | Coursera: Software Processes |
| 6 | Курс Системный дизайн - обучение проектированию систем - Кар | Личный кабинет PurpleSchool | Coursera: Software Processes | Google UX Design Professional Certificate | Coursera | Coursera: Modeling Software Systems Using UML |
| 7 | Курс Архитектор решений (Solution architect) в «Специалист» | Software Design and Architecture | Coursera | API Design Patterns - JJ Geewax | Design Patterns | Coursera | English Corpora: most widely used online corpora. Billions o |
| 8 | Google UX Design Professional Certificate | Coursera | Хекслет: введение (Intro to Git) | Курс Системный дизайн - обучение проектированию систем - Кар | Understanding the Barrel Pattern in JavaScript/TypeScript - | API Design Patterns - JJ Geewax |
| 9 | Основы PHP для начинающих: бесплатный курс по PHP | Хекслет: введение (Go Basics) | Курс Архитектор решений (Solution architect) в «Специалист» | Free Online Mermaid Editor — Flowcharts, Sequence Diagrams & | Career Community & Conversations | Glassdoor |
| 10 | Хекслет: урок «Требования» (Java Development Overview) | Figma: микросервисы (дизайн-файл) | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: | Figma: микросервисы (дизайн-файл) | Курс Системный дизайн - обучение проектированию систем - Кар |

### Запрос 7: «криптовалютные биржи и обменники»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | MEXC — покупка криптовалюты | MEXC — покупка криптовалюты | SafelyChange | MEXC — покупка криптовалюты | SafelyChange |
| 2 | HTX — криптобиржа | HTX — криптобиржа | LetsExchange — крипто-обменник | LetsExchange — крипто-обменник | LetsExchange — крипто-обменник |
| 3 | LetsExchange — крипто-обменник | LetsExchange — крипто-обменник | HTX — криптобиржа | HTX — криптобиржа | HTX — криптобиржа |
| 4 | Morpho Vaults — DeFi кредитование | Morpho Vaults — DeFi кредитование | Купить Tether (USDT) по карте за рубли и доллары | P2P | Bit | SafelyChange | Купить Tether (USDT) по карте за рубли и доллары | P2P | Bit |
| 5 | Купить Tether (USDT) по карте за рубли и доллары | P2P | Bit | Купить Tether (USDT) по карте за рубли и доллары | P2P | Bit | MEXC — покупка криптовалюты | Плати по всему миру | MEXC — покупка криптовалюты |
| 6 | Плати по всему миру | Плати по всему миру | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: | Morpho Vaults — DeFi кредитование | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) |
| 7 | SafelyChange | SafelyChange | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) | Купить Tether (USDT) по карте за рубли и доллары | P2P | Bit | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf |
| 8 | BingX: пополнение баланса | BingX: пополнение баланса | SHAREWOOD.TECH (ex .BIZ) (Шервуд) 🔥 Слив Курсов – Скачать Бе | BingX: пополнение баланса | Проектирование архитектуры и интеграций (API, брокеры) + ИИ: |
| 9 | Коллекция Перемещение во времени, перерождени и другой мир - | Коллекция Перемещение во времени, перерождени и другой мир - | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | Коллекция Перемещение во времени, перерождени и другой мир - | SHAREWOOD.TECH (ex .BIZ) (Шервуд) 🔥 Слив Курсов – Скачать Бе |
| 10 | Операция «Панда» (2024) | Операция «Панда» (2024) | BingX: пополнение баланса | 100 форекс книг | BingX: пополнение баланса |

### Запрос 8: «chrome devtools debugging tips»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Chrome DevTools: как открыть и как работать с инструментами | Stacks: sBTC Dual Stacking | Гид по инструментам разработчика в браузере (Chrome, Mozilla | Stacks: sBTC Dual Stacking | Гид по инструментам разработчика в браузере (Chrome, Mozilla |
| 2 | Профессиональное применение инструментов разработчика Chrome | DevTools: как вызвать в браузере консоль - функции и возможн | Chrome DevTools: как открыть и как работать с инструментами | JSON Web Tokens - jwt.io | Chrome DevTools: как открыть и как работать с инструментами |
| 3 | DevTools: как вызвать в браузере консоль - функции и возможн | Chrome DevTools: как открыть и как работать с инструментами | DevTools: как вызвать в браузере консоль - функции и возможн | Morpho Vaults — DeFi кредитование | DevTools: как вызвать в браузере консоль - функции и возможн |
| 4 | Гид по инструментам разработчика в браузере (Chrome, Mozilla | Профессиональное применение инструментов разработчика Chrome | Профессиональное применение инструментов разработчика Chrome | Knowledge Core | API Design Patterns - JJ Geewax |
| 5 | Stacks: sBTC Dual Stacking | Гид по инструментам разработчика в браузере (Chrome, Mozilla | API Design Patterns - JJ Geewax | Хекслет: введение (Go Basics) | Обзор инструментов разработки в браузерах - Изучение веб-раз |
| 6 | Шаблонный метод | Хекслет: введение (Go Basics) | Knowledge Core | LetsExchange — крипто-обменник | JSON Web Tokens - jwt.io |
| 7 | ТОП-10 ошибок в спецификации OpenAPI и как их избежать | Хекслет: урок «Структуры» (Go Basics) | Обзор инструментов разработки в браузерах - Изучение веб-раз | HTX — криптобиржа | Профессиональное применение инструментов разработчика Chrome |
| 8 | Хекслет: введение (Go Basics) | LetsExchange — крипто-обменник | JSON Web Tokens - jwt.io | Free Online Mermaid Editor — Flowcharts, Sequence Diagrams & | Jobs That Fit | Resumes That Stand Out | JobLeads |
| 9 | Хекслет: урок «Структуры» (Go Basics) | Morpho Vaults — DeFi кредитование | Keycloak Admin REST API | Хекслет: урок «Структуры» (Go Basics) | AI Text Formatter – Format Text Online Free (Unlimited) |
| 10 | LetsExchange — крипто-обменник | Stepik: Postman для тестирования API — урок «Введение» | Understanding the Barrel Pattern in JavaScript/TypeScript - | 100 форекс книг | Design Patterns | Coursera |

### Запрос 9: «книга про психологию лжи»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит | Пол Экман, Психология лжи. Обмани меня, если сможешь – купит |
| 2 | Книга "Узнай лжеца по выражению лица" — купить в интернет-ма | Книга "Узнай лжеца по выражению лица" — купить в интернет-ма | Книга "Узнай лжеца по выражению лица" — купить в интернет-ма | 100 форекс книг | Книга "Узнай лжеца по выражению лица" — купить в интернет-ма |
| 3 | Читать Мачеха и ее подруги! | Ателье колдовских колпаков / Tongari Boushi no Atelier | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | «Release it! Проектирование и дизайн ПО для тех, кому не всё |
| 4 | Краснов Сергей скачать аудиокниги исполнителя онлайн » | Краснов Сергей скачать аудиокниги исполнителя онлайн » | Скачать книгу «Изучаем PostgreSQL 10» [id:655852] в формате | «Эволюционная архитектура. Поддержка непрерывных изменений ( | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат |
| 5 | Фермерская жизнь в ином мире / Isekai Nonbiri Nouka | Фермерская жизнь в ином мире / Isekai Nonbiri Nouka | «Release it! Проектирование и дизайн ПО для тех, кому не всё | Скачать книгу «Изучаем PostgreSQL 10» [id:655852] в формате | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) |
| 6 | Ателье колдовских колпаков / Tongari Boushi no Atelier | прием врача игнатова кардиолог 30.09.26 17 40 | «Эволюционная архитектура. Поддержка непрерывных изменений ( | EnglishSpace: мнемотехника для английского | 100 форекс книг |
| 7 | прием врача игнатова кардиолог 30.09.26 17 40 | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | «Создание микросервисов (pdf+epub)» Сэм Ньюмен – скачать pdf | Чистая архитектура. Искусство разработки программного обеспе |
| 8 | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | Кейнз Олег скачать аудиокниги исполнителя онлайн » | 100 форекс книг | Чистая архитектура. Искусство разработки программного обеспе | Изучение английского языка самостоятельно: бесплатные матери |
| 9 | «Release it! Проектирование и дизайн ПО для тех, кому не всё | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат | Изучение английского языка самостоятельно: бесплатные матери | Morpho Vaults — DeFi кредитование | «Микросервисы. Паттерны разработки и рефакторинга (pdf+epub) |
| 10 | Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачат | 100 форекс книг | SHAREWOOD.TECH (ex .BIZ) (Шервуд) 🔥 Слив Курсов – Скачать Бе | Хоть я и бездарная злодейка — аниме | Восемьдесят шесть — Манга |

### Запрос 10: «english vocabulary dictionaries»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base |
|---|---|---|---|---|---|
| 1 | EnglishSpace: мнемотехника для английского | EnglishSpace: мнемотехника для английского | English Corpora: most widely used online corpora. Billions o | Morpho Vaults — DeFi кредитование | WooordHunt (Вордхант) — ваш помощник в мире английского язык |
| 2 | 4ege.ru: таблица 12 времён английского | 4ege.ru: таблица 12 времён английского | Learn irregular verbs effectively | 100 форекс книг | English Corpora: most widely used online corpora. Billions o |
| 3 | English Corpora: most widely used online corpora. Billions o | WooordHunt (Вордхант) — ваш помощник в мире английского язык | Изучение английского языка самостоятельно: бесплатные матери | Хекслет: введение (Go Basics) | English news and easy articles for students of English |
| 4 | Обзор инструментов разработки в браузерах - Изучение веб-раз | Listen in English – Free ESL Listening Practice | English news and easy articles for students of English | Stacks: sBTC Dual Stacking | EnglishSpace: мнемотехника для английского |
| 5 | WooordHunt (Вордхант) — ваш помощник в мире английского язык | 100 статей за 3 месяца: авторский план изучения грамматики с | Listen in English – Free ESL Listening Practice | Терри Пратчетт, Нил Гейман - Добрые Предзнаменования (2006) | Find jobs abroad with languages | Europe Language Jobs |
| 6 | Listen in English – Free ESL Listening Practice | Гид по инструментам разработчика в браузере (Chrome, Mozilla | 4ege.ru: таблица 12 времён английского | Free Online Mermaid Editor — Flowcharts, Sequence Diagrams & | Learn irregular verbs effectively |
| 7 | Изучение английского языка самостоятельно: бесплатные матери | Голосовой блокнот - Speechpad.ru | API Design Patterns - JJ Geewax | Аудиокнига Первый игрок скачать торрент бесплатно mp3 | Listen in English – Free ESL Listening Practice |
| 8 | Степени сравнения английских прилагательных: правила degrees | Обзор инструментов разработки в браузерах - Изучение веб-раз | EnglishSpace: мнемотехника для английского | HTX — криптобиржа | 4ege.ru: таблица 12 времён английского |
| 9 | Passive Voice в английском языке: правила, примеры, формула | EnglishClass101 — бесплатные материалы месяца | EnglishClass101 — бесплатные материалы месяца | Learn irregular verbs effectively | Изучение английского языка самостоятельно: бесплатные матери |
| 10 | 100 статей за 3 месяца: авторский план изучения грамматики с | Шаблонный метод | Jobs That Fit | Resumes That Stand Out | JobLeads | Хекслет: введение (Intro to Git) | Степени сравнения английских прилагательных: правила degrees |

## Таблица 2. Близость пар — top-15 по 20 типичным заметкам

### current@128

| # | Пара | cos |
|---|---|---|
| current@128 | 384 | 512 | 128 | 128 | Apache-2.0 |

### current@512

| # | Пара | cos |
|---|---|---|
| current@512 | 384 | 512 | 128 | 512 | Apache-2.0 |

### e5-small

| # | Пара | cos |
|---|---|---|
| e5-small | 384 | 512 | 512 | 512 | MIT |

### rubert-tiny2

| # | Пара | cos |
|---|---|---|
| rubert-tiny2 | 312 | 2048 | 2048 | 2048 | MIT |

### e5-base

| # | Пара | cos |
|---|---|---|
| e5-base | 768 | 514 | 512 | 512 | MIT |

## Таблица 3. Ключевые слова — top-10 (гибрид: частотный шорт-лист → модель; yake — базовая линия)

### «Эволюционная архитектура. Поддержка непрерывных изменений (pdf+epub)»

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | pdf epub | текст pdf | pdf средний рейтинг | текст pdf pdf | текст pdf | PDF Средний рейтинг |
| 2 | pdf pdf | текст pdf pdf | pdf pdf средний | pdf средний рейтинг | pdf epub | PDF Средний |
| 3 | текст pdf pdf | текст | текст pdf | pdf pdf средний | текст pdf pdf | pdf |
| 4 | pdf | pdf pdf | pdf средний | текст pdf | pdf средний рейтинг | Текст PDF |
| 5 | текст pdf | pdf | pdf | pdf epub | текст | Поддержка непрерывных изменений |
| 6 | основе | pdf epub | текст pdf pdf | рейтинг основе оценок | pdf | Средний рейтинг |
| 7 | epub | основе | текст | pdf pdf | pdf pdf средний | Основной контент книги |
| 8 | pdf pdf средний | отзыв | pdf pdf | средний рейтинг основе | рейтинг основе оценок | Нил Форд |
| 9 | pdf средний | средний | отзыв | рейтинг основе | pdf средний | Эволюционная архитектура. Поддержка |
| 10 | средний | epub | pdf epub | основе оценок | рейтинг основе | Основной контент |

### «Release it! Проектирование и дизайн ПО для тех, кому не всё равно» Ма

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | текст pdf | текст | pdf средний рейтинг | текст pdf pdf | текст | PDF Средний рейтинг |
| 2 | pdf | подписке | текст pdf | pdf средний рейтинг | текст pdf | PDF Средний |
| 3 | текст pdf pdf | основе | pdf средний | pdf pdf средний | pdf средний рейтинг | Текст PDF |
| 4 | pdf pdf | текст pdf | pdf pdf средний | текст pdf | подписке | pdf |
| 5 | pdf epub | текст pdf pdf | текст pdf pdf | pdf pdf | pdf pdf средний | Средний рейтинг |
| 6 | текст | средний | pdf | средний рейтинг основе | pdf | Основной контент книги |
| 7 | основе | pdf | текст | рейтинг основе | pdf средний | равно Текст PDF |
| 8 | подписке | рейтинг основе | оценок подписке | pdf epub | текст pdf pdf | Основной контент |
| 9 | средний | рейтинг | средний рейтинг основе | основе оценок | рейтинг основе | Йонтс Текст PDF |
| 10 | epub | оценок подписке | рейтинг основе | оценок подписке | pdf epub | Нвайву Текст PDF |

### Проектирование архитектуры и интеграций (API, брокеры) + ИИ: Начальный

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | архитектуры | интеграций | интеграций | технических | курс | приложение Проектирование архитектуры |
| 2 | проектирование | курс | проектирование | проектирование | проектирование | Перейти в приложение |
| 3 | интеграций | технических | архитектуры | архитектуры | уровень | Проектирование архитектуры |
| 4 | технических | проектирование | технических | интеграций | брокеры | Начальный уровень |
| 5 | stepik | архитектуры | брокеры | stepik | rabbitmq | Сертификат Stepik Чему |
| 6 | брокеры | работы | сообщений | брокеры | работы | брокеры |
| 7 | курс | сообщений | kafka | уровень | stepik | архитектуры и интеграций |
| 8 | уровень | rabbitmq | курс | курс | архитектуры | брокеры сообщений |
| 9 | работы | уровень | уровень | рынок | рынок | API |
| 10 | rabbitmq | это | rabbitmq | api | данных | проектирование баз данных |

### Чистая архитектура. Искусство разработки программного обеспечения Робе

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | разработки программного обеспечения | разработки программного | разработки программного | разработки программного обеспечения | разработки программного обеспечения | Текст Текст Средний |
| 2 | разработки программного | текст | разработки программного обеспечения | разработки программного | разработки программного | Текст Средний рейтинг |
| 3 | программного обеспечения | программного | программного обеспечения | рейтинг основе оценок | программного обеспечения | Чистая архитектура. Искусство |
| 4 | программного | разработки программного обеспечения | программного | средний рейтинг основе | текст | Текст Текст |
| 5 | книги | разработки | разработки | рейтинг основе | программного | Искусство разработки программного |
| 6 | архитектура | книги | архитектура | архитектура | архитектура | разработки программного обеспечения |
| 7 | разработки | основе | обеспечения | основе оценок | рейтинг основе | Текст Средний |
| 8 | обеспечения | программного обеспечения | текст | программного обеспечения | книги | Основной контент книги |
| 9 | основе | чистая | рейтинг основе оценок | разработки | рейтинг основе оценок | Роберт Мартин Питер |
| 10 | искусство | обеспечения | основе оценок | книги | чистая | Средний рейтинг |

### Билли Саммерс Стивен Кинг Издательство АСТ – купить и скачать книгу в 

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | стивен кинг текст | текст | стивен кинг текст | стивен кинг текст | билли саммерс | Стивен Кинг Текст |
| 2 | стивен кинг | билли саммерс | стивен кинг | средний рейтинг основе | саммерс | Стивен Кинг |
| 3 | билли саммерс | средний | билли саммерс | средний рейтинг | стивен кинг текст | Билли Саммерс Объем |
| 4 | доступен | текст доступен | кинг текст | кинг текст | билли | Читай-городе и Буквоеде |
| 5 | саммерс | саммерс | стивен | рейтинг основе | стивен кинг | аудиоформат Средний рейтинг |
| 6 | стивен | кинг текст | текст | текст доступен | стивен | Саммерс Стивен Кинг |
| 7 | средний | стивен кинг текст | саммерс | рейтинг | текст | доступен аудиоформат Средний |
| 8 | основе | основе | билли | текст | кинг текст | Билли Саммерс Стивен |
| 9 | текст | доступен | кинг | доступен аудиоформат | кинг | Текст Текст Средний |
| 10 | текст доступен | оценок | средний рейтинг основе | оценок | текст доступен | Билли Саммерс |

### Find jobs abroad with languages | Europe Language Jobs

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | jobs abroad | jobs abroad | language jobs | language jobs | danish | next job abroad |
| 2 | find jobs | language jobs | languages europe | jobs abroad | languages | Europe Language Jobs |
| 3 | language jobs | find jobs | jobs abroad | abroad languages | language jobs | Europe Language |
| 4 | abroad languages | abroad languages | europe language | find jobs | europe | Find jobs abroad |
| 5 | languages europe | languages europe | find jobs | languages europe | languages europe | Europe |
| 6 | europe language | europe language | abroad languages | europe language | apply | On-Demand Transportation Service |
| 7 | jobs | abroad | europe | companies hiring | abroad | abroad |
| 8 | hiring | jobs | hiring | job | jobs abroad | Transportation Service |
| 9 | abroad | hiring | abroad | jobs | jobs | jobs abroad |
| 10 | job | europe | languages | hiring | language | Find |

### JSON Web Tokens - jwt.io

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | json web tokens | json web tokens | json web tokens | json web tokens | json web tokens | JSON Web Tokens |
| 2 | json web | json web | web tokens | web tokens | jwt | JSON Web |
| 3 | jwt | jwt | json web | json web | json web | Web Tokens |
| 4 | json | json | token | claims breakdown | web tokens | Web Token |
| 5 | web tokens | web tokens | claims breakdown | encoder | json | JSON JSON Copy |
| 6 | decoder | decoder | json | decoder | verify | Web |
| 7 | encoded | encoded | decode | jwt | tokens | generate JSON Web |
| 8 | decode | decode | tokens | decode | token | JSON |
| 9 | encoder | encoder | jwt | verify | web | JWT |
| 10 | tokens | web | web | json | decode | Claims Breakdown Copy |

### Фермерская жизнь в ином мире / Isekai Nonbiri Nouka

| # | current@128 | current@512 | e5-small | rubert-tiny2 | e5-base | yake |
|---|---|---|---|---|---|---|
| 1 | бог | бог | хираку | жизнь | хираку | Хираку Матио попал |
| 2 | жизнь | жизнь | жизнь | бог | бог | Матио попал |
| 3 | хираку | хираку | бог | хираку | жизнь | Хираку Матио |
| 4 | — | — | — | — | — | Прожив недолгую |
| 5 | — | — | — | — | — | полную страданий |
| 6 | — | — | — | — | — | неизлечимой болезни |
| 7 | — | — | — | — | — | аудиенцию к богу-творцу |
| 8 | — | — | — | — | — | страданий и боли |
| 9 | — | — | — | — | — | умерев от неизлечимой |
| 10 | — | — | — | — | — | попал на аудиенцию |

## Таблица 4. Скорость (CPU, внутри контейнера)

| Вариант | embed p50, мс | embed p95, мс | весь корпус, с | keybert p50, мс | keybert p95, мс |
|---|---|---|---|---|---|
| current@128 | 95 | 158 | 11.0 | 89 | 104 |
| current@512 | 239 | 368 | 24.2 | 86 | 92 |
| e5-small | 256 | 381 | 23.1 | 105 | 135 |
| rubert-tiny2 | 137 | 603 | 40.2 | 56 | 115 |
| e5-base | 2320 | 3427 | 234.5 | 1058 | 1792 |

## Таблица 5. Память контейнера с загруженной моделью

| Вариант | RSS процесса, МБ | прирост над baseline, МБ |
|---|---|---|
| current@128 | 1091 | 758 |
| current@512 | 1092 | 757 |
| e5-small | 1166 | 760 |
| rubert-tiny2 | 545 | 157 |
| e5-base | 1768 | 1374 |

## Воспроизведение

```
docker run --rm -e HF_HUB_OFFLINE=0 -e HF_HOME=/root/.cache/huggingface   -v "D:/knowledge-graph/huggingface_cache:/root/.cache/huggingface"   -v "<workdir>:/work" knowledge-graph-nlp-personal sh -c   "pip install -q keybert; for v in current@128 current@512 e5-small rubert-tiny2 e5-base; do      python /work/measure_models.py --dataset /work/notes_dataset.json --out /work/MODEL-1-findings.md --only $v; done"
```

Скрипт: `nlp-service/scripts/measure_models.py`. Датасет: экспорт `notes` Personal-базы (113 заметок).
