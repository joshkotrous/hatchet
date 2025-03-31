/* eslint-disable no-console */
import { Logger, LogLevel, LogLevelEnum } from '@util/logger';

export const DEFAULT_LOGGER = (context: string, logLevel?: LogLevel) =>
  new HatchetLogger(context, logLevel);

export class HatchetLogger implements Logger {
  private logLevel: LogLevel;
  private context: string;

  constructor(context: string, logLevel: LogLevel = 'INFO') {
    this.logLevel = logLevel;
    this.context = context;
  }

  /**
   * Sanitizes input for safe log output
   * Prevents log injection by escaping special characters
   */
  private sanitizeForLog(input: unknown): string {
    if (input === undefined || input === null) {
      return '';
    }
    
    const str = String(input);
    // Replace newlines, carriage returns, and other control chars that could affect log structure
    return str.replace(/[\r\n\t\f\v\b\0\u001B]/g, (match) => {
      const replacements: Record<string, string> = {
        '\r': '\\r',
        '\n': '\\n',
        '\t': '\\t',
        '\f': '\\f',
        '\v': '\\v',
        '\b': '\\b',
        '\0': '\\0',
        '\u001B': '\\e', // Escape sequence used for ANSI color codes
      };
      return replacements[match] || '';
    });
  }

  private log(level: LogLevel, message: string, color: string = '37'): void {
    if (LogLevelEnum[level] >= LogLevelEnum[this.logLevel]) {
      const time = new Date().toLocaleString('en-US', {
        month: '2-digit',
        day: '2-digit',
        year: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });

      // eslint-disable-next-line prefer-destructuring
      let print = console.log;

      if (level === 'ERROR') {
        print = console.error;
      }

      if (level === 'WARN') {
        print = console.warn;
      }

      if (level === 'INFO') {
        print = console.info;
      }

      if (level === 'DEBUG') {
        print = console.debug;
      }

      // Sanitize message to prevent log injection
      const sanitizedMessage = this.sanitizeForLog(message);

      // eslint-disable-next-line no-console
      print(
        `🪓 ${process.pid} | ${time} ${color && `\x1b[${color || ''}m`} [${level}/${this.context}] ${sanitizedMessage}\x1b[0m`
      );
    }
  }

  debug(message: string): void {
    this.log('DEBUG', message, '35');
  }

  info(message: string): void {
    this.log('INFO', message);
  }

  green(message: string): void {
    this.log('INFO', message, '32');
  }

  warn(message: string, error?: Error): void {
    let logMessage = message;
    if (error) {
      // For Error objects, extract useful information safely
      const errorInfo = error.stack 
        ? `${error.name}: ${error.message} | ${error.stack.split('\n')[0]}`
        : `${error.name}: ${error.message}`;
      logMessage = `${message} ${errorInfo}`;
    }
    this.log('WARN', logMessage, '93');
  }

  error(message: string, error?: Error): void {
    let logMessage = message;
    if (error) {
      // For Error objects, extract useful information safely
      const errorInfo = error.stack 
        ? `${error.name}: ${error.message} | ${error.stack.split('\n')[0]}`
        : `${error.name}: ${error.message}`;
      logMessage = `${message} ${errorInfo}`;
    }
    this.log('ERROR', logMessage, '91');
  }
}