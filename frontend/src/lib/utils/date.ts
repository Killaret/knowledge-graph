/**
 * Утилиты для форматирования дат
 */

const DEFAULT_OPTIONS: Intl.DateTimeFormatOptions = {
  day: 'numeric',
  month: 'long',
  year: 'numeric'
};

const DEFAULT_LOCALE = 'ru-RU';

/**
 * Форматирует дату в локализованную строку
 * @param dateString - ISO строка даты или Date объект
 * @param options - Опции форматирования Intl.DateTimeFormat
 * @param locale - Локаль (по умолчанию 'ru-RU')
 * @returns Отформатированная строка даты
 */
export function formatDate(
  dateString: string | Date,
  options?: Intl.DateTimeFormatOptions,
  locale: string = DEFAULT_LOCALE
): string {
  try {
    const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
    
    if (isNaN(date.getTime())) {
      return 'Некорректная дата';
    }
    
    const formatter = new Intl.DateTimeFormat(locale, options ?? DEFAULT_OPTIONS);
    return formatter.format(date);
  } catch (error) {
    console.error('Error formatting date:', error);
    return String(dateString);
  }
}

/**
 * Форматирует дату с временем
 * @param dateString - ISO строка даты или Date объект
 * @param locale - Локаль (по умолчанию 'ru-RU')
 * @returns Отформатированная строка даты и времени
 */
export function formatDateTime(
  dateString: string | Date,
  locale: string = DEFAULT_LOCALE
): string {
  const options: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  };
  
  return formatDate(dateString, options, locale);
}

/**
 * Форматирует относительную дату (сегодня, вчера, и т.д.)
 * @param dateString - ISO строка даты или Date объект
 * @returns Относительная строка даты
 */
export function formatRelativeDate(dateString: string | Date): string {
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) {
    return 'Сегодня';
  } else if (diffDays === 1) {
    return 'Вчера';
  } else if (diffDays < 7) {
    return `${diffDays} дня назад`;
  } else {
    return formatDate(dateString);
  }
}
