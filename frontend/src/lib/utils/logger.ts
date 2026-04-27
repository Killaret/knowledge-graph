// Утилита для централизованного логирования в проекте
// Логи сохраняются в console и могут быть отправлены на сервер при необходимости

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogEntry {
  timestamp: string;
  level: LogLevel;
  message: string;
  context?: string;
  data?: unknown;
}

class Logger {
  private context: string;
  private logBuffer: LogEntry[] = [];
  private readonly maxBufferSize = 100;
  private readonly enableFileLogging = import.meta.env.DEV; // Только в dev режиме

  constructor(context: string = 'app') {
    this.context = context;
  }

  private createLogEntry(level: LogLevel, message: string, data?: unknown): LogEntry {
    return {
      timestamp: new Date().toISOString(),
      level,
      message,
      context: this.context,
      data
    };
  }

  private logToConsole(entry: LogEntry): void {
    const prefix = `[${entry.timestamp}] [${entry.level.toUpperCase()}] [${entry.context}]`;
    
    switch (entry.level) {
      case 'debug':
        console.debug(prefix, entry.message, entry.data || '');
        break;
      case 'info':
        console.info(prefix, entry.message, entry.data || '');
        break;
      case 'warn':
        console.warn(prefix, entry.message, entry.data || '');
        break;
      case 'error':
        console.error(prefix, entry.message, entry.data || '');
        break;
    }
  }

  private addToBuffer(entry: LogEntry): void {
    this.logBuffer.push(entry);
    if (this.logBuffer.length > this.maxBufferSize) {
      this.logBuffer.shift();
    }
  }

  debug(message: string, data?: unknown): void {
    const entry = this.createLogEntry('debug', message, data);
    this.logToConsole(entry);
    this.addToBuffer(entry);
  }

  info(message: string, data?: unknown): void {
    const entry = this.createLogEntry('info', message, data);
    this.logToConsole(entry);
    this.addToBuffer(entry);
  }

  warn(message: string, data?: unknown): void {
    const entry = this.createLogEntry('warn', message, data);
    this.logToConsole(entry);
    this.addToBuffer(entry);
  }

  error(message: string, error?: unknown): void {
    const entry = this.createLogEntry('error', message, error);
    this.logToConsole(entry);
    this.addToBuffer(entry);
  }

  // Получить все логи из буфера
  getLogs(): LogEntry[] {
    return [...this.logBuffer];
  }

  // Очистить буфер
  clearLogs(): void {
    this.logBuffer = [];
  }
}

// Фабрика для создания логеров с разными контекстами
export function createLogger(context: string): Logger {
  return new Logger(context);
}

// Глобальный логер
export const logger = new Logger('app');

export default logger;
