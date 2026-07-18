export type Locale = "en" | "ru";
export type Mode = "standard" | "galactic";
/**
 * Galactic Lexicon - Cosmic-themed messaging system
 * Transforms technical messages into galactic metaphors
 */

export type MessageCategory = "success" | "error" | "info" | "warning";
export type MessageKey = string;

const messages: Record<
  Locale,
  Record<
    Mode,
    Record<MessageCategory, Record<string, (...args: any[]) => string>>
  >
> = {
  en: {
    standard: {
      success: {
        noteCreated: (title: string) => `Note "${title}" created successfully.`,
        noteUpdated: (title: string) => `Note "${title}" updated.`,
        noteDeleted: () => "Note deleted.",
        linkCreated: (source: string, target: string) =>
          `Link from "${source}" to "${target}" created.`,
        linkDeleted: () => "Link deleted.",
        settingsSaved: () => "Settings saved.",
        achievementUnlocked: (title: string) =>
          `Achievement unlocked: ${title}!`,
        shareCreated: () => "Share link created.",
        shareRevoked: () => "Access revoked.",
        loginSuccess: () => "Login successful.",
        logoutSuccess: () => "Logout successful.",
        accountDeleted: () => "Account deleted.",
        passwordChanged: () => "Password changed.",
      },
      error: {
        validation: (field: string) => `Invalid value in field "${field}".`,
        duplicateLink: () => "This link already exists.",
        noteNotFound: () => "Note not found.",
        linkNotFound: () => "Link not found.",
        unauthorized: () => "Authorization required.",
        forbidden: () => "Insufficient permissions.",
        serverError: () => "Server error. Please try again later.",
        networkError: () => "Connection error. Check your internet.",
        invalidCredentials: () => "Invalid username or password.",
        duplicateNote: () => "A note with this title already exists.",
        shareNotFound: () => "Share link not found or invalid.",
        maxSharesReached: () => "Maximum number of share links reached.",
      },
      info: {
        emptyGraph: () =>
          "The graph is empty. Create your first note to see the starry sky.",
        loading: () => "Loading data...",
        noResults: () => "Nothing found.",
        searchHint: () => "Enter a search query to explore the universe.",
        firstNoteHint: () =>
          "Create your first star in the universe of knowledge!",
        achievementProgress: (current: number, target: number) =>
          `Progress: ${current}/${target}`,
        streakActive: (days: number) =>
          `You have been active for ${days} days in a row!`,
        newFeature: (feature: string) => `New feature available: ${feature}`,
      },
      warning: {
        unsavedChanges: () => "There are unsaved changes.",
        deleteConfirm: (item: string) =>
          `Are you sure you want to delete "${item}"?`,
        leavePage: () => "Leaving the page will lose unsaved changes.",
        sessionExpiring: () =>
          "Your session is expiring soon. Please save your data.",
      },
    },
    galactic: {
      success: {
        noteCreated: (title: string) =>
          `Star "${title}" ignited in your galaxy. ✨`,
        noteUpdated: (title: string) =>
          `The glow of star "${title}" has changed. 🌟`,
        noteDeleted: () =>
          "A star has faded, but its light continues to travel. 💫",
        linkCreated: (source: string, target: string) =>
          `A gravity beam stretches from "${source}" to "${target}". 🌌`,
        linkDeleted: () => "The gravitational bridge has collapsed. 🌑",
        settingsSaved: () => "Galactic coordinates updated. 📍",
        achievementUnlocked: (title: string) =>
          `⭐ New star on the achievement map: ${title}!`,
        shareCreated: () => "A portal to another galaxy has opened. 🚪",
        shareRevoked: () => "The portal has closed forever. 🔒",
        loginSuccess: () => "Teleportation into the galaxy complete. 🚀",
        logoutSuccess: () => "Returning to your home planet. 🌍",
        accountDeleted: () =>
          "The galaxy has been consumed by a black hole. Everything created remains among the stars. 🌌",
        passwordChanged: () => "Access code for the starship updated. 🔐",
      },
      error: {
        validation: (field: string) =>
          `Sensors detected an anomaly in the field "${field}". Signal distorted. 📡`,
        duplicateLink: () =>
          "This gravitational bridge already exists in spacetime. 🌉",
        noteNotFound: () =>
          "Star not found in the constellation. Check the coordinates. 🌠",
        linkNotFound: () =>
          "Gravitational bridge not detected. It may have collapsed. ⚫",
        unauthorized: () =>
          "Access to the star system denied. Authorization required. 🛡️",
        forbidden: () =>
          "This region of the galaxy is shielded. Access forbidden. ⛔",
        serverError: () =>
          "Cosmic anomaly! Servers consumed by a supernova flare. Please try later. 🌋",
        networkError: () =>
          "Connection to the cosmic network lost. Check the hyperchannel. 📡",
        invalidCredentials: () =>
          "Invalid access code for the starship. The captain does not recognize you. 👨‍🚀",
        duplicateNote: () =>
          "A star with this name already shines in this galaxy. ☀️",
        shareNotFound: () => "Portal not found or vanished into a wormhole. 🌀",
        maxSharesReached: () =>
          "Limit of portals to other galaxies reached. 🚫",
      },
      info: {
        emptyGraph: () =>
          "Your starry sky is empty. Create your first star to begin building the universe! 🌌",
        loading: () => "Scanning the galaxy... Calibrating the telescope. 🔭",
        noResults: () => "Sensors detected nothing in this part of space. 🛰️",
        searchHint: () =>
          "Enter coordinates to search for stars in our universe. 🌟",
        firstNoteHint: () =>
          "Be a pioneer! Create your first star in the universe of knowledge! ⭐",
        achievementProgress: (current: number, target: number) =>
          `Stellar map filling: ${current} of ${target} objects discovered 🗺️`,
        streakActive: (days: number) =>
          `Continuous journey: ${days} days without losing contact with base! 🚀`,
        newFeature: (feature: string) =>
          `New technology available on board: ${feature} 🛸`,
      },
      warning: {
        unsavedChanges: () => "There are unsaved entries in the ship log. 📓",
        deleteConfirm: (item: string) =>
          `Are you sure you want to send "${item}" into a black hole? This is irreversible. 🕳️`,
        leavePage: () => "Exiting hyperspace will lose unsaved data. ⚠️",
        sessionExpiring: () =>
          "Teleportation fuel is running low. Save your coordinates! ⛽",
      },
    },
  },
  ru: {
    standard: {
      success: {
        noteCreated: (title: string) => `Заметка "${title}" успешно создана.`,
        noteUpdated: (title: string) => `Заметка "${title}" обновлена.`,
        noteDeleted: () => "Заметка удалена.",
        linkCreated: (source: string, target: string) =>
          `Связь от "${source}" к "${target}" создана.`,
        linkDeleted: () => "Связь удалена.",
        settingsSaved: () => "Настройки сохранены.",
        achievementUnlocked: (title: string) =>
          `Достижение получено: ${title}!`,
        shareCreated: () => "Ссылка для доступа создана.",
        shareRevoked: () => "Доступ отозван.",
        loginSuccess: () => "Вход выполнен успешно.",
        logoutSuccess: () => "Выход выполнен успешно.",
        accountDeleted: () => "Аккаунт удален.",
        passwordChanged: () => "Пароль изменен.",
      },
      error: {
        validation: (field: string) => `Неверное значение в поле "${field}".`,
        duplicateLink: () => "Такая связь уже существует.",
        noteNotFound: () => "Заметка не найдена.",
        linkNotFound: () => "Связь не найдена.",
        unauthorized: () => "Требуется авторизация.",
        forbidden: () => "Недостаточно прав.",
        serverError: () => "Ошибка сервера. Попробуйте позже.",
        networkError: () => "Ошибка соединения. Проверьте интернет.",
        invalidCredentials: () => "Неверный логин или пароль.",
        duplicateNote: () => "Заметка с таким названием уже существует.",
        shareNotFound: () => "Ссылка доступа не найдена или недействительна.",
        maxSharesReached: () => "Достигнут лимит количества ссылок доступа.",
      },
      info: {
        emptyGraph: () =>
          "Граф пуст. Создайте первую заметку, чтобы увидеть звёздное небо.",
        loading: () => "Загрузка данных...",
        noResults: () => "Ничего не найдено.",
        searchHint: () => "Введите поисковый запрос для поиска по вселенной.",
        firstNoteHint: () => "Создайте свою первую звезду во вселенной знаний!",
        achievementProgress: (current: number, target: number) =>
          `Прогресс: ${current}/${target}`,
        streakActive: (days: number) => `Вы активны уже ${days} дней подряд!`,
        newFeature: (feature: string) => `Новая функция доступна: ${feature}`,
      },
      warning: {
        unsavedChanges: () => "Есть несохраненные изменения.",
        deleteConfirm: (item: string) =>
          `Вы уверены, что хотите удалить "${item}"?`,
        leavePage: () => "При уходе со страницы изменения будут потеряны.",
        sessionExpiring: () => "Сессия скоро закончится. Сохраните данные.",
      },
    },
    galactic: {
      success: {
        noteCreated: (title: string) =>
          `Звезда «${title}» зажжена в вашей галактике. ✨`,
        noteUpdated: (title: string) =>
          `Свечение звезды «${title}» изменилось. 🌟`,
        noteDeleted: () =>
          "Звезда погасла, но её свет продолжает путешествовать. 💫",
        linkCreated: (source: string, target: string) =>
          `Луч гравитации протянут от «${source}» к «${target}». 🌌`,
        linkDeleted: () => "Гравитационный мост разрушен. 🌑",
        settingsSaved: () => "Координаты галактики обновлены. 📍",
        achievementUnlocked: (title: string) =>
          `⭐ Новая звезда на карте достижений: ${title}!`,
        shareCreated: () => "Портал в другую галактику открыт. 🚪",
        shareRevoked: () => "Портал закрыт навсегда. 🔒",
        loginSuccess: () => "Телепортация в галактику завершена. 🚀",
        logoutSuccess: () => "Возвращение на родную планету. 🌍",
        accountDeleted: () =>
          "Галактика поглощена чёрной дырой. Всё, что было создано, остаётся в звёздах. 🌌",
        passwordChanged: () =>
          "Код доступа к космическому кораблю обновлён. 🔐",
      },
      error: {
        validation: (field: string) =>
          `Сенсоры зафиксировали аномалию в поле «${field}». Сигнал искажён. 📡`,
        duplicateLink: () =>
          "Этот гравитационный мост уже существует в пространстве-времени. 🌉",
        noteNotFound: () =>
          "Звезда не найдена в созвездии. Проверьте координаты. 🌠",
        linkNotFound: () =>
          "Гравитационный мост не обнаружен. Возможно, он коллапсировал. ⚫",
        unauthorized: () =>
          "Отказано в доступе к звёздной системе. Требуется авторизация. 🛡️",
        forbidden: () =>
          "Эта область галактики под защитой щита. Доступ запрещён. ⛔",
        serverError: () =>
          "Космическая аномалия! Сервера поглощены вспышкой сверхновой. Попробуйте позже. 🌋",
        networkError: () =>
          "Потеряно соединение с космической сетью. Проверьте гиперканал. 📡",
        invalidCredentials: () =>
          "Неверный код доступа к звездолёту. Капитан не узнаёт вас. 👨‍🚀",
        duplicateNote: () =>
          "Звезда с таким именем уже светит в этой галактике. ☀️",
        shareNotFound: () => "Портал не найден или исчез в червоточине. 🌀",
        maxSharesReached: () =>
          "Достигнут лимит порталов в другие галактики. 🚫",
      },
      info: {
        emptyGraph: () =>
          "Ваше звёздное небо пусто. Создайте первую звезду, чтобы начать создавать вселенную! 🌌",
        loading: () => "Сканируем галактику... Телескоп настраивается. 🔭",
        noResults: () =>
          "Сенсоры ничего не обнаружили в этой части космоса. 🛰️",
        searchHint: () =>
          "Введите координаты для поиска звёзд в нашей вселенной. 🌟",
        firstNoteHint: () =>
          "Будьте первооткрывателем! Создайте свою первую звезду во вселенной знаний! ⭐",
        achievementProgress: (current: number, target: number) =>
          `Звёздная карта заполняется: ${current} из ${target} объектов открыто 🗺️`,
        streakActive: (days: number) =>
          `Непрерывное путешествие: ${days} дней без потери связи с базой! 🚀`,
        newFeature: (feature: string) =>
          `Новая технология доступна на борту: ${feature} 🛸`,
      },
      warning: {
        unsavedChanges: () =>
          "В бортовом журнале есть несохранённые записи. 📓",
        deleteConfirm: (item: string) =>
          `Вы уверены, что хотите отправить «${item}» в чёрную дыру? Это необратимо. 🕳️`,
        leavePage: () =>
          "При выходе из гиперпространства несохранённые данные будут потеряны. ⚠️",
        sessionExpiring: () =>
          "Топливо для телепортации заканчивается. Сохраните координаты! ⛽",
      },
    },
  },
};

/**
 * Message formatter that supports locale and galactic mode
 */
export class MessageFormatter {
  private useGalacticMode: boolean;
  private locale: Locale;

  constructor(useGalacticMode: boolean = false, locale: Locale = "en") {
    this.useGalacticMode = useGalacticMode;
    this.locale = locale;
  }

  /**
   * Format a message with the current locale and mode
   */
  format(category: MessageCategory, key: MessageKey, ...args: any[]): string {
    const mode: Mode = this.useGalacticMode ? "galactic" : "standard";
    const localeMessages = messages[this.locale];

    if (!localeMessages) {
      return `[${this.locale}.${mode}.${category}.${key}]`;
    }

    const modeMessages = localeMessages[mode];
    const categoryMessages = modeMessages[category];

    if (!categoryMessages || !(key in categoryMessages)) {
      return `[${this.locale}.${mode}.${category}.${key}]`;
    }

    return categoryMessages[key](...args);
  }

  /**
   * Get success message
   */
  success(key: MessageKey, ...args: any[]): string {
    return this.format("success", key, ...args);
  }

  /**
   * Get error message
   */
  error(key: MessageKey, ...args: any[]): string {
    return this.format("error", key, ...args);
  }

  /**
   * Get info message
   */
  info(key: MessageKey, ...args: any[]): string {
    return this.format("info", key, ...args);
  }

  /**
   * Get warning message
   */
  warning(key: MessageKey, ...args: any[]): string {
    return this.format("warning", key, ...args);
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

  /**
   * Set locale
   */
  setLocale(locale: Locale): void {
    this.locale = locale;
  }

  /**
   * Get current locale
   */
  getLocale(): Locale {
    return this.locale;
  }
}

/**
 * Create a formatter with galactic mode and locale settings
 */
export function createFormatter(
  galacticMode: boolean,
  locale: Locale = "en",
): MessageFormatter {
  return new MessageFormatter(galacticMode, locale);
}

/**
 * Get all available message keys from the English standard dictionary
 */
export function getMessageKeys(): Record<MessageCategory, string[]> {
  return {
    success: Object.keys(messages.en.standard.success),
    error: Object.keys(messages.en.standard.error),
    info: Object.keys(messages.en.standard.info),
    warning: Object.keys(messages.en.standard.warning),
  };
}

/**
 * Legacy compatibility function for lexicon-settings.ts
 * Maps old API to new MessageFormatter system
 */
type LegacyCategory = "success" | "error" | "info" | "warning" | "achievement";

const keyToMethodMap: Record<LegacyCategory, Record<string, string>> = {
  success: { unlocked: "achievementUnlocked" },
  error: { connectionExists: "duplicateLink", generic: "serverError" },
  info: { saved: "settingsSaved" },
  warning: { unsavedChanges: "unsavedChanges" },
  achievement: { unlocked: "achievementUnlocked" },
};

export function getLexiconMessage(
  locale: Locale,
  mode: Mode,
  category: LegacyCategory,
  key: string,
  ...params: any[]
): string {
  const methodKey = keyToMethodMap[category]?.[key] || key;
  const targetCategory: MessageCategory =
    category === "achievement" ? "success" : category;
  return createFormatter(mode === "galactic", locale).format(
    targetCategory,
    methodKey,
    ...params,
  );
}

/**
 * Legacy compatibility object for tests and old call sites
 */
export const GalacticLexicon = {
  success: {
    noteCreated: (
      title: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).success("noteCreated", title),
    noteUpdated: (
      title: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).success("noteUpdated", title),
    noteDeleted: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("noteDeleted"),
    linkCreated: (
      source: string,
      target: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) =>
      createFormatter(useGalactic, locale).success(
        "linkCreated",
        source,
        target,
      ),
    linkDeleted: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("linkDeleted"),
    settingsSaved: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("settingsSaved"),
    achievementUnlocked: (
      title: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) =>
      createFormatter(useGalactic, locale).success(
        "achievementUnlocked",
        title,
      ),
    shareCreated: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("shareCreated"),
    shareRevoked: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("shareRevoked"),
    loginSuccess: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("loginSuccess"),
    logoutSuccess: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("logoutSuccess"),
    accountDeleted: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("accountDeleted"),
    passwordChanged: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).success("passwordChanged"),
  },
  error: {
    validation: (
      field: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).error("validation", field),
    duplicateLink: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("duplicateLink"),
    noteNotFound: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("noteNotFound"),
    linkNotFound: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("linkNotFound"),
    unauthorized: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("unauthorized"),
    forbidden: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("forbidden"),
    serverError: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("serverError"),
    networkError: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("networkError"),
    invalidCredentials: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("invalidCredentials"),
    duplicateNote: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("duplicateNote"),
    shareNotFound: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("shareNotFound"),
    maxSharesReached: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).error("maxSharesReached"),
  },
  info: {
    emptyGraph: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).info("emptyGraph"),
    loading: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).info("loading"),
    noResults: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).info("noResults"),
    searchHint: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).info("searchHint"),
    firstNoteHint: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).info("firstNoteHint"),
    achievementProgress: (
      current: number,
      target: number,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) =>
      createFormatter(useGalactic, locale).info(
        "achievementProgress",
        current,
        target,
      ),
    streakActive: (
      days: number,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).info("streakActive", days),
    newFeature: (
      feature: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).info("newFeature", feature),
  },
  warning: {
    unsavedChanges: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).warning("unsavedChanges"),
    deleteConfirm: (
      item: string,
      useGalactic: boolean = false,
      locale: Locale = "en",
    ) => createFormatter(useGalactic, locale).warning("deleteConfirm", item),
    leavePage: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).warning("leavePage"),
    sessionExpiring: (useGalactic: boolean = false, locale: Locale = "en") =>
      createFormatter(useGalactic, locale).warning("sessionExpiring"),
  },
};
