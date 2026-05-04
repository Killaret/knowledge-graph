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
 * Global galactic lexicon instance
 */
export const GalacticLexicon = {
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
