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
