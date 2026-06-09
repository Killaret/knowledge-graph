export type Locale = 'ru' | 'en'
export type Mode = 'standard' | 'galactic'
/**
 * Galactic Lexicon - Cosmic-themed messaging system
 * Transforms technical messages into galactic metaphors
 */

export type MessageCategory = 'success' | 'error' | 'info' | 'warning';
export type MessageKey = string;

// Technical messages (default mode)
const technicalMessages = {
  success: {
    noteCreated: (title: string) => `Заметка "${title}" успешно создана.`,
    noteUpdated: (title: string) => `Заметка "${title}" обновлена.`,
    noteDeleted: () => 'Заметка удалена.',
    linkCreated: (source: string, target: string) => `Связь от "${source}" к "${target}" создана.`,
    linkDeleted: () => 'Связь удалена.',
    settingsSaved: () => 'Настройки сохранены.',
    achievementUnlocked: (title: string) => `Достижение получено: ${title}!`,
    shareCreated: () => 'Ссылка для доступа создана.',
    shareRevoked: () => 'Доступ отозван.',
    loginSuccess: () => 'Вход выполнен успешно.',
    logoutSuccess: () => 'Выход выполнен успешно.',
    accountDeleted: () => 'Аккаунт удален.',
    passwordChanged: () => 'Пароль изменен.',
  },
  error: {
    validation: (field: string) => `Неверное значение в поле "${field}".`,
    duplicateLink: () => 'Такая связь уже существует.',
    noteNotFound: () => 'Заметка не найдена.',
    linkNotFound: () => 'Связь не найдена.',
    unauthorized: () => 'Требуется авторизация.',
    forbidden: () => 'Недостаточно прав.',
    serverError: () => 'Ошибка сервера. Попробуйте позже.',
    networkError: () => 'Ошибка соединения. Проверьте интернет.',
    invalidCredentials: () => 'Неверный логин или пароль.',
    duplicateNote: () => 'Заметка с таким названием уже существует.',
    shareNotFound: () => 'Ссылка доступа не найдена или недействительна.',
    maxSharesReached: () => 'Достигнут лимит количества ссылок доступа.',
  },
  info: {
    emptyGraph: () => 'Граф пуст. Создайте первую заметку, чтобы увидеть звёздное небо.',
    loading: () => 'Загрузка данных...',
    noResults: () => 'Ничего не найдено.',
    searchHint: () => 'Введите поисковый запрос для поиска по вселенной.',
    firstNoteHint: () => 'Создайте свою первую звезду во вселенной знаний!',
    achievementProgress: (current: number, target: number) => `Прогресс: ${current}/${target}`,
    streakActive: (days: number) => `Вы активны уже ${days} дней подряд!`,
    newFeature: (feature: string) => `Новая функция доступна: ${feature}`,
  },
  warning: {
    unsavedChanges: () => 'Есть несохраненные изменения.',
    deleteConfirm: (item: string) => `Вы уверены, что хотите удалить "${item}"?`,
    leavePage: () => 'При уходе со страницы изменения будут потеряны.',
    sessionExpiring: () => 'Сессия скоро закончится. Сохраните данные.',
  },
};

// Galactic messages (cosmic mode)
const galacticMessages = {
  success: {
    noteCreated: (title: string) => `Звезда «${title}» зажжена в вашей галактике. ✨`,
    noteUpdated: (title: string) => `Свечение звезды «${title}» изменилось. 🌟`,
    noteDeleted: () => 'Звезда погасла, но её свет продолжает путешествовать. 💫',
    linkCreated: (source: string, target: string) => `Луч гравитации протянут от «${source}» к «${target}». 🌌`,
    linkDeleted: () => 'Гравитационный мост разрушен. 🌑',
    settingsSaved: () => 'Координаты галактики обновлены. 📍',
    achievementUnlocked: (title: string) => `⭐ Новая звезда на карте достижений: ${title}!`,
    shareCreated: () => 'Портал в другую галактику открыт. 🚪',
    shareRevoked: () => 'Портал закрыт навсегда. 🔒',
    loginSuccess: () => 'Телепортация в галактику завершена. 🚀',
    logoutSuccess: () => 'Возвращение на родную планету. 🌍',
    accountDeleted: () => 'Галактика поглощена чёрной дырой. Всё, что было создано, остаётся в звёздах. 🌌',
    passwordChanged: () => 'Код доступа к космическому кораблю обновлён. 🔐',
  },
  error: {
    validation: (field: string) => `Сенсоры зафиксировали аномалию в поле «${field}». Сигнал искажён. 📡`,
    duplicateLink: () => 'Этот гравитационный мост уже существует в пространстве-времени. 🌉',
    noteNotFound: () => 'Звезда не найдена в созвездии. Проверьте координаты. 🌠',
    linkNotFound: () => 'Гравитационный мост не обнаружен. Возможно, он коллапсировал. ⚫',
    unauthorized: () => 'Отказано в доступе к звёздной системе. Требуется авторизация. 🛡️',
    forbidden: () => 'Эта область галактики под защитой щита. Доступ запрещён. ⛔',
    serverError: () => 'Космическая аномалия! Сервера поглощены вспышкой сверхновой. Попробуйте позже. 🌋',
    networkError: () => 'Потеряно соединение с космической сетью. Проверьте гиперканал. 📡',
    invalidCredentials: () => 'Неверный код доступа к звездолёту. Капитан не узнаёт вас. 👨‍🚀',
    duplicateNote: () => 'Звезда с таким именем уже светит в этой галактике. ☀️',
    shareNotFound: () => 'Портал не найден или исчез в червоточине. 🌀',
    maxSharesReached: () => 'Достигнут лимит порталов в другие галактики. 🚫',
  },
  info: {
    emptyGraph: () => 'Ваше звёздное небо пусто. Создайте первую звезду, чтобы начать создавать вселенную! 🌌',
    loading: () => 'Сканируем галактику... Телескоп настраивается. 🔭',
    noResults: () => 'Сенсоры ничего не обнаружили в этой части космоса. 🛰️',
    searchHint: () => 'Введите координаты для поиска звёзд в нашей вселенной. 🌟',
    firstNoteHint: () => 'Будьте первооткрывателем! Создайте свою первую звезду во вселенной знаний! ⭐',
    achievementProgress: (current: number, target: number) => `Звёздная карта заполняется: ${current} из ${target} объектов открыто 🗺️`,
    streakActive: (days: number) => `Непрерывное путешествие: ${days} дней без потери связи с базой! 🚀`,
    newFeature: (feature: string) => `Новая технология доступна на борту: ${feature} 🛸`,
  },
  warning: {
    unsavedChanges: () => 'В бортовом журнале есть несохранённые записи. 📓',
    deleteConfirm: (item: string) => `Вы уверены, что хотите отправить «${item}» в чёрную дыру? Это необратимо. 🕳️`,
    leavePage: () => 'При выходе из гиперпространства несохранённые данные будут потеряны. ⚠️',
    sessionExpiring: () => 'Топливо для телепортации заканчивается. Сохраните координаты! ⛽',
  },
};

/**
 * Message formatter that supports both technical and galactic modes
 */
export class MessageFormatter {
  private useGalacticMode: boolean;

  constructor(useGalacticMode: boolean = false) {
    this.useGalacticMode = useGalacticMode;
  }

  /**
   * Format a message with the current mode
   */
  format(category: MessageCategory, key: MessageKey, ...args: any[]): string {
    const messages = this.useGalacticMode ? galacticMessages : technicalMessages;
    const categoryMessages = messages[category];
    
    if (!categoryMessages || !(key in categoryMessages)) {
      return `[${category}.${key}]`;
    }

    const messageFn = categoryMessages[key as keyof typeof categoryMessages];
    if (typeof messageFn === 'function') {
      return (messageFn as (...args: any[]) => string)(...args);
    }

    return String(messageFn);
  }

  /**
   * Get success message
   */
  success(key: MessageKey, ...args: any[]): string {
    return this.format('success', key, ...args);
  }

  /**
   * Get error message
   */
  error(key: MessageKey, ...args: any[]): string {
    return this.format('error', key, ...args);
  }

  /**
   * Get info message
   */
  info(key: MessageKey, ...args: any[]): string {
    return this.format('info', key, ...args);
  }

  /**
   * Get warning message
   */
  warning(key: MessageKey, ...args: any[]): string {
    return this.format('warning', key, ...args);
  }

  /**
   * Set galactic mode
   */
  setGalacticMode(enabled: boolean): void {
    this.useGalacticMode = enabled;
  }

  /**
   * Check if galactic mode is enabled
   */
  isGalacticMode(): boolean {
    return this.useGalacticMode;
  }
}

/**
 * Global galactic message helper for legacy message formatting
 */
export const GalacticMessageLibrary = {
  success: {
    noteCreated: (title: string, useGalactic = false) => 
      useGalactic ? galacticMessages.success.noteCreated(title) : technicalMessages.success.noteCreated(title),
    noteUpdated: (title: string, useGalactic = false) =>
      useGalactic ? galacticMessages.success.noteUpdated(title) : technicalMessages.success.noteUpdated(title),
    noteDeleted: (useGalactic = false) =>
      useGalactic ? galacticMessages.success.noteDeleted() : technicalMessages.success.noteDeleted(),
    linkCreated: (source: string, target: string, useGalactic = false) =>
      useGalactic ? galacticMessages.success.linkCreated(source, target) : technicalMessages.success.linkCreated(source, target),
    achievementUnlocked: (title: string, useGalactic = false) =>
      useGalactic ? galacticMessages.success.achievementUnlocked(title) : technicalMessages.success.achievementUnlocked(title),
    loginSuccess: (useGalactic = false) =>
      useGalactic ? galacticMessages.success.loginSuccess() : technicalMessages.success.loginSuccess(),
  },
  error: {
    validation: (field: string, useGalactic = false) =>
      useGalactic ? galacticMessages.error.validation(field) : technicalMessages.error.validation(field),
    duplicateLink: (useGalactic = false) =>
      useGalactic ? galacticMessages.error.duplicateLink() : technicalMessages.error.duplicateLink(),
    unauthorized: (useGalactic = false) =>
      useGalactic ? galacticMessages.error.unauthorized() : technicalMessages.error.unauthorized(),
    serverError: (useGalactic = false) =>
      useGalactic ? galacticMessages.error.serverError() : technicalMessages.error.serverError(),
  },
  info: {
    emptyGraph: (useGalactic = false) =>
      useGalactic ? galacticMessages.info.emptyGraph() : technicalMessages.info.emptyGraph(),
    firstNoteHint: (useGalactic = false) =>
      useGalactic ? galacticMessages.info.firstNoteHint() : technicalMessages.info.firstNoteHint(),
    streakActive: (days: number, useGalactic = false) =>
      useGalactic ? galacticMessages.info.streakActive(days) : technicalMessages.info.streakActive(days),
  },
  warning: {
    unsavedChanges: (useGalactic = false) =>
      useGalactic ? galacticMessages.warning.unsavedChanges() : technicalMessages.warning.unsavedChanges(),
    deleteConfirm: (item: string, useGalactic = false) =>
      useGalactic ? galacticMessages.warning.deleteConfirm(item) : technicalMessages.warning.deleteConfirm(item),
  },
};

/**
 * Create a formatter with galactic mode setting
 */
export function createFormatter(galacticMode: boolean): MessageFormatter {
  return new MessageFormatter(galacticMode);
}

/**
 * Get all available message keys
 */
export function getMessageKeys(): Record<MessageCategory, string[]> {
  return {
    success: Object.keys(technicalMessages.success),
    error: Object.keys(technicalMessages.error),
    info: Object.keys(technicalMessages.info),
    warning: Object.keys(technicalMessages.warning),
  };
}

/**
 * Legacy compatibility function for lexicon-settings.ts
 * Maps old API to new MessageFormatter system
 */
type LegacyCategory = 'success' | 'error' | 'info' | 'warning' | 'achievement'

// Mapping from old keys to new MessageFormatter methods
const keyToMethodMap: Record<LegacyCategory, Record<string, string>> = {
  success: {
    noteCreated: 'noteCreated',
    noteUpdated: 'noteUpdated',
    unlocked: 'achievementUnlocked',
  },
  error: {
    connectionExists: 'duplicateLink',
    generic: 'serverError',
  },
  info: {
    saved: 'settingsSaved',
  },
  warning: {
    unsavedChanges: 'unsavedChanges',
  },
  achievement: {
    unlocked: 'achievementUnlocked',
  },
}

export function getLexiconMessage(locale: Locale, mode: Mode, category: LegacyCategory, key: string, ...params: any[]): string {
  const formatter = createFormatter(mode === 'galactic')
  const methodKey = keyToMethodMap[category]?.[key] || key
  
  // Try to call the method dynamically
  try {
    const categoryObj = (formatter as any)[category]
    if (categoryObj && typeof categoryObj[methodKey] === 'function') {
      return categoryObj[methodKey](...params, mode === 'galactic')
    }
  } catch {
    // Fallback to generic message
  }
  
  // Fallback messages
  const fallbackMessages: Record<LegacyCategory, Record<string, (mode: Mode) => string>> = {
    success: {
      noteCreated: (m) => m === 'galactic' ? `Звезда зажжена в вашей галактике` : `Заметка создана`,
      noteUpdated: (m) => m === 'galactic' ? `Орбита скорректирована` : `Заметка обновлена`,
      unlocked: (m) => m === 'galactic' ? `✨ Новая звезда на карте` : `⭐ Достижение получено`,
    },
    error: {
      connectionExists: (m) => m === 'galactic' ? `Гравитационный мост уже существует` : `Связь уже существует`,
      generic: (m) => m === 'galactic' ? `Космический шторм` : `Произошла ошибка`,
    },
    info: {
      saved: (m) => m === 'galactic' ? `Данные зафиксированы в созвездии` : `Сохранено`,
    },
    warning: {
      unsavedChanges: (m) => m === 'galactic' ? `Коммуникация прервана` : `Есть несохранённые изменения`,
    },
    achievement: {
      unlocked: (m) => m === 'galactic' ? `✨ Вы покорили новую звезду: ${params[0] || ''}` : `⭐ Достижение получено: ${params[0] || ''}`,
    },
  }
  
  return fallbackMessages[category]?.[key]?.(mode) || '...'
}

/**
 * Legacy compatibility object for tests that expect GalacticLexicon.success.noteCreated('title', false)
 */
export const GalacticLexicon = {
  success: {
    noteCreated: (title: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.noteCreated(title) : technicalMessages.success.noteCreated(title),
    noteUpdated: (title: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.noteUpdated(title) : technicalMessages.success.noteUpdated(title),
    noteDeleted: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.noteDeleted() : technicalMessages.success.noteDeleted(),
    linkCreated: (source: string, target: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.linkCreated(source, target) : technicalMessages.success.linkCreated(source, target),
    linkDeleted: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.linkDeleted() : technicalMessages.success.linkDeleted(),
    settingsSaved: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.settingsSaved() : technicalMessages.success.settingsSaved(),
    achievementUnlocked: (title: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.achievementUnlocked(title) : technicalMessages.success.achievementUnlocked(title),
    shareCreated: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.shareCreated() : technicalMessages.success.shareCreated(),
    shareRevoked: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.shareRevoked() : technicalMessages.success.shareRevoked(),
    loginSuccess: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.loginSuccess() : technicalMessages.success.loginSuccess(),
    logoutSuccess: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.logoutSuccess() : technicalMessages.success.logoutSuccess(),
    accountDeleted: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.accountDeleted() : technicalMessages.success.accountDeleted(),
    passwordChanged: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.success.passwordChanged() : technicalMessages.success.passwordChanged(),
  },
  error: {
    validation: (field: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.validation(field) : technicalMessages.error.validation(field),
    duplicateLink: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.duplicateLink() : technicalMessages.error.duplicateLink(),
    noteNotFound: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.noteNotFound() : technicalMessages.error.noteNotFound(),
    linkNotFound: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.linkNotFound() : technicalMessages.error.linkNotFound(),
    unauthorized: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.unauthorized() : technicalMessages.error.unauthorized(),
    forbidden: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.forbidden() : technicalMessages.error.forbidden(),
    serverError: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.serverError() : technicalMessages.error.serverError(),
    networkError: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.networkError() : technicalMessages.error.networkError(),
    invalidCredentials: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.invalidCredentials() : technicalMessages.error.invalidCredentials(),
    duplicateNote: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.duplicateNote() : technicalMessages.error.duplicateNote(),
    shareNotFound: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.shareNotFound() : technicalMessages.error.shareNotFound(),
    maxSharesReached: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.error.maxSharesReached() : technicalMessages.error.maxSharesReached(),
  },
  info: {
    emptyGraph: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.emptyGraph() : technicalMessages.info.emptyGraph(),
    loading: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.loading() : technicalMessages.info.loading(),
    noResults: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.noResults() : technicalMessages.info.noResults(),
    searchHint: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.searchHint() : technicalMessages.info.searchHint(),
    firstNoteHint: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.firstNoteHint() : technicalMessages.info.firstNoteHint(),
    achievementProgress: (current: number, target: number, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.achievementProgress(current, target) : technicalMessages.info.achievementProgress(current, target),
    streakActive: (days: number, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.streakActive(days) : technicalMessages.info.streakActive(days),
    newFeature: (feature: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.info.newFeature(feature) : technicalMessages.info.newFeature(feature),
  },
  warning: {
    unsavedChanges: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.warning.unsavedChanges() : technicalMessages.warning.unsavedChanges(),
    deleteConfirm: (item: string, useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.warning.deleteConfirm(item) : technicalMessages.warning.deleteConfirm(item),
    leavePage: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.warning.leavePage() : technicalMessages.warning.leavePage(),
    sessionExpiring: (useGalactic: boolean = false) => 
      useGalactic ? galacticMessages.warning.sessionExpiring() : technicalMessages.warning.sessionExpiring(),
  },
}
