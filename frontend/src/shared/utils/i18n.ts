/**
 * Internationalization (i18n) helper
 * Provides message formatting with language support
 */

export type Locale = "en" | "ru";
export type MessageParams = Record<string, string | number>;

// Message dictionary with language support
const messages: Record<Locale, Record<string, string>> = {
  en: {
    // Success messages
    "note.created": 'Note "{{title}}" created successfully.',
    "note.updated": 'Note "{{title}}" updated.',
    "note.deleted": "Note deleted.",
    "link.created": 'Link from "{{source}}" to "{{target}}" created.',
    "link.deleted": "Link deleted.",
    "settings.saved": "Settings saved.",
    "login.success": "Login successful.",
    "logout.success": "Logout successful.",

    // Error messages
    "validation.error": 'Invalid value in field "{{field}}".',
    "duplicate.link": "This link already exists.",
    "note.notFound": "Note not found.",
    "link.notFound": "Link not found.",
    unauthorized: "Authorization required.",
    forbidden: "Insufficient permissions.",
    "server.error": "Server error. Please try again later.",
    "network.error": "Connection error. Check your internet.",
    "invalid.credentials": "Invalid username or password.",
    "graph.notFound": "Graph not found. The note may have been deleted.",
    "graph.serverError": "Server error while loading the graph. Please try again later.",
    "graph.networkError": "Could not connect to the server. Check your internet connection.",
    "graph.loadError": "Graph loading error: {{message}}.",
    "graph.unknownError": "Unknown error while loading the graph.",

    // Info messages
    "empty.graph":
      "Graph is empty. Create your first note to see the starry sky.",
    loading: "Loading data...",
    "no.results": "Nothing found.",
    "search.hint": "Enter search query to search the universe.",
    "first.note.hint": "Create your first star in the universe of knowledge!",

    // Warning messages
    "unsaved.changes": "There are unsaved changes.",
    "delete.confirm": 'Are you sure you want to delete "{{item}}"?',
    "leave.page": "Leaving the page will lose unsaved changes.",

    // Filters and sorting
    "filter.all": "All",
    "filter.inbox": "Inbox",
    "filter.type.dust": "Cosmic Dust",
    "filter.type.blackhole": "Black Holes",
    "filter.type.unknown": "Unknown",
    "filter.type.star": "Stars",
    "filter.type.planet": "Planets",
    "filter.type.moon": "Moons",
    "filter.type.comet": "Comets",
    "filter.type.galaxy": "Galaxies",
    "filter.type.nebula": "Nebulae",
    "filter.type.asteroid": "Asteroids",
    "filter.type.satellite": "Satellites",
    "filter.type.debris": "Debris",
    "filter.type.technical": "Technical",
    "filter.type.reality_rift": "Reality Rifts",
    "filter.type.chromatic_maw": "Chromatic Maws",
    "filter.type.void_whisper": "Void Whispers",
    "filter.type.cosmic_abomination": "Cosmic Abominations",
    "sort.created": "Created (newest first)",
    "sort.updated": "Updated (recent first)",
    "sort.type": "Type (alphabetical)",
    "notes.loadError": "Failed to load notes",

    // Auth labels and errors
    "auth.signInTitle": "Sign in",
    "auth.loginPasswordMode": "Login / Password",
    "auth.apiKeyMode": "API Key",
    "auth.apiKeyLabel": "API Key",
    "auth.apiKeyPlaceholder": "Enter your API key",
    "auth.loginLabel": "Login",
    "auth.loginPlaceholder": "Enter login",
    "auth.emailLabel": "Email",
    "auth.emailPlaceholder": "Enter email",
    "auth.passwordLabel": "Password",
    "auth.confirmPasswordLabel": "Confirm Password",
    "auth.passwordPlaceholder": "Enter password",
    "auth.signInButton": "Sign in",
    "auth.signingInButton": "Signing in...",
    "auth.invalidApiKey": "Invalid API key",
    "auth.enterLoginAndPassword": "Please enter login and password",
    "auth.invalidCredentials": "Invalid credentials",
    "auth.registerTitle": "Registration",
    "auth.loginRequired": "Login is required",
    "auth.emailRequired": "Email is required",
    "auth.passwordRequirementsNotMet": "Password does not meet requirements",
    "auth.passwordsDoNotMatch": "Passwords do not match",
    "auth.registrationFailed": "Registration failed",
    "auth.chooseLoginPlaceholder": "Choose a login",
    "auth.enterEmailPlaceholder": "Enter email",
    "auth.createPasswordPlaceholder": "Create a password",
    "auth.repeatPasswordPlaceholder": "Repeat password",
    "auth.passwordRequirementsTitle": "Password requirements:",
    "auth.passwordMinChars": "Minimum 10 characters",
    "auth.passwordUppercase": "Uppercase letter",
    "auth.passwordLowercase": "Lowercase letter",
    "auth.passwordNumber": "Number",
    "auth.passwordSpecial": "Special character (!@#$%^&*)",
    "auth.registerButton": "Register",
    "auth.registeringButton": "Registering...",
    "auth.alreadyHaveAccount": "Already have an account? Sign in",

    // UI labels
    "create.note": "Create Note",
    "edit.note": "Edit Profile",
    "delete.note": "Delete Note",
    cancel: "Cancel",
    save: "Save",
    search: "Search",
    filter: "Filter",
    sort: "Sort",
    all: "All",
    type: "Type",
    title: "Title",
    content: "Content",
    links: "Links",
    settings: "Settings",
    help: "Help",
    close: "Close",
    readonly: "Read only",
    "login.readonly": "Login cannot be changed",
    "email.placeholder": "Enter your email",
    "password.confirm": "Please enter your password for confirmation",
    "confirm.delete": "Confirm Delete",
    "delete.account": "Delete Account",
  },
  ru: {
    // Success messages
    "note.created": 'Заметка "{{title}}" успешно создана.',
    "note.updated": 'Заметка "{{title}}" обновлена.',
    "note.deleted": "Заметка удалена.",
    "link.created": 'Связь от "{{source}}" к "{{target}}" создана.',
    "link.deleted": "Связь удалена.",
    "settings.saved": "Настройки сохранены.",
    "login.success": "Вход выполнен успешно.",
    "logout.success": "Выход выполнен успешно.",

    // Error messages
    "validation.error": 'Неверное значение в поле "{{field}}".',
    "field.password": "пароль",
    "duplicate.link": "Такая связь уже существует.",
    "note.notFound": "Заметка не найдена.",
    "link.notFound": "Связь не найдена.",
    unauthorized: "Требуется авторизация.",
    forbidden: "Недостаточно прав.",
    "server.error": "Ошибка сервера. Попробуйте позже.",
    "network.error": "Ошибка соединения. Проверьте интернет.",
    "invalid.credentials": "Неверный логин или пароль.",
    "graph.notFound": "Граф не найден. Возможно, заметка была удалена.",
    "graph.serverError": "Ошибка сервера при загрузке графа. Попробуйте позже.",
    "graph.networkError": "Не удалось подключиться к серверу. Проверьте интернет-соединение.",
    "graph.loadError": "Ошибка загрузки графа: {{message}}.",
    "graph.unknownError": "Неизвестная ошибка при загрузке графа.",

    // Info messages
    "empty.graph":
      "Граф пуст. Создайте первую заметку, чтобы увидеть звёздное небо.",
    loading: "Загрузка данных...",
    "no.results": "Ничего не найдено.",
    "search.hint": "Введите поисковый запрос для поиска по вселенной.",
    "first.note.hint": "Создайте свою первую звезду во вселенной знаний!",

    // Warning messages
    "unsaved.changes": "Есть несохраненные изменения.",
    "delete.confirm": 'Вы уверены, что хотите удалить "{{item}}"?',
    "leave.page": "При уходе со страницы изменения будут потеряны.",

    // Filters and sorting
    "filter.all": "Все",
    "filter.inbox": "Входящие",
    "filter.type.dust": "Космическая пыль",
    "filter.type.blackhole": "Чёрные дыры",
    "filter.type.unknown": "Неизвестное",
    "filter.type.star": "Звёзды",
    "filter.type.planet": "Планеты",
    "filter.type.moon": "Луны",
    "filter.type.comet": "Кометы",
    "filter.type.galaxy": "Галактики",
    "filter.type.nebula": "Туманности",
    "filter.type.asteroid": "Астероиды",
    "filter.type.satellite": "Спутники",
    "filter.type.debris": "Обломки",
    "filter.type.technical": "Техническое",
    "filter.type.reality_rift": "Разломы реальности",
    "filter.type.chromatic_maw": "Хроматические пасти",
    "filter.type.void_whisper": "Шёпот пустоты",
    "filter.type.cosmic_abomination": "Космические чудовища",
    "sort.created": "Создано (сначала новые)",
    "sort.updated": "Обновлено (недавние)",
    "sort.type": "Тип (по алфавиту)",
    "notes.loadError": "Не удалось загрузить заметки",

    // Auth labels and errors
    "auth.signInTitle": "Вход",
    "auth.loginPasswordMode": "Логин / Пароль",
    "auth.apiKeyMode": "API-ключ",
    "auth.apiKeyLabel": "API-ключ",
    "auth.apiKeyPlaceholder": "Введите ваш API-ключ",
    "auth.loginLabel": "Логин",
    "auth.loginPlaceholder": "Введите логин",
    "auth.emailLabel": "Email",
    "auth.emailPlaceholder": "Введите email",
    "auth.passwordLabel": "Пароль",
    "auth.confirmPasswordLabel": "Подтвердите пароль",
    "auth.passwordPlaceholder": "Введите пароль",
    "auth.signInButton": "Войти",
    "auth.signingInButton": "Вход...",
    "auth.invalidApiKey": "Неверный API-ключ",
    "auth.enterLoginAndPassword": "Введите логин и пароль",
    "auth.invalidCredentials": "Неверные учётные данные",
    "auth.registerTitle": "Регистрация",
    "auth.loginRequired": "Введите логин",
    "auth.emailRequired": "Введите email",
    "auth.passwordRequirementsNotMet": "Пароль не соответствует требованиям",
    "auth.passwordsDoNotMatch": "Пароли не совпадают",
    "auth.registrationFailed": "Ошибка регистрации",
    "auth.chooseLoginPlaceholder": "Придумайте логин",
    "auth.enterEmailPlaceholder": "Введите email",
    "auth.createPasswordPlaceholder": "Придумайте пароль",
    "auth.repeatPasswordPlaceholder": "Повторите пароль",
    "auth.passwordRequirementsTitle": "Требования к паролю:",
    "auth.passwordMinChars": "Минимум 10 символов",
    "auth.passwordUppercase": "Заглавная буква",
    "auth.passwordLowercase": "Строчная буква",
    "auth.passwordNumber": "Цифра",
    "auth.passwordSpecial": "Специальный символ (!@#$%^&*)",
    "auth.registerButton": "Зарегистрироваться",
    "auth.registeringButton": "Регистрация...",
    "auth.alreadyHaveAccount": "Уже есть аккаунт? Войдите",

    // UI labels
    "create.note": "Создать заметку",
    "edit.note": "Редактировать профиль",
    "delete.note": "Удалить заметку",
    cancel: "Отмена",
    save: "Сохранить",
    search: "Поиск",
    filter: "Фильтр",
    sort: "Сортировка",
    all: "Все",
    type: "Тип",
    title: "Заголовок",
    content: "Содержание",
    links: "Связи",
    settings: "Настройки",
    help: "Справка",
    close: "Закрыть",
    readonly: "Только чтение",
    "login.readonly": "Логин нельзя изменить",
    "email.placeholder": "Введите ваш email",
    "password.confirm": "Введите пароль для подтверждения",
    "confirm.delete": "Подтвердить удаление",
    "delete.account": "Удалить аккаунт",
  },
};

/**
 * Format a message with parameters
 * @param key - Message key
 * @param locale - Language locale (default: 'en')
 * @param params - Parameters to replace in message (optional)
 * @returns Formatted message
 */
export function formatMessage(
  key: string,
  locale: Locale = "en",
  params?: MessageParams,
): string {
  const message = messages[locale]?.[key] || messages.en[key] || key;

  if (!params) {
    return message;
  }

  // Replace {{param}} placeholders
  return Object.entries(params).reduce((result, [paramKey, paramValue]) => {
    return result.replace(
      new RegExp(`{{${paramKey}}}`, "g"),
      String(paramValue),
    );
  }, message);
}

/**
 * Get current locale from localStorage or default
 * @returns Current locale
 */
export function getCurrentLocale(): Locale {
  if (typeof window === "undefined") return "en";

  const stored = localStorage.getItem("locale");
  if (stored === "en" || stored === "ru") {
    return stored;
  }

  return "en";
}

/**
 * Set current locale in localStorage
 * @param locale - Locale to set
 */
export function setLocale(locale: Locale): void {
  if (typeof window !== "undefined") {
    localStorage.setItem("locale", locale);
  }
}
