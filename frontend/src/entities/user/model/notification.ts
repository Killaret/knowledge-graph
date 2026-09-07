/**
 * Notification — Value Object for a toast/overlay notification.
 *
 * Centralizes icon selection, CSS class derivation, and lifecycle metadata
 * (duration, auto-close) so that the UI component only handles rendering
 * and timers.
 */

export type NotificationType = "success" | "error" | "info" | "warning";

export interface NotificationProps {
  message: string;
  type?: NotificationType;
  duration?: number;
  useGalacticMode?: boolean;
}

export class Notification {
  private static readonly ICONS: Record<NotificationType, string> = {
    success: "✅",
    error: "❌",
    info: "ℹ️",
    warning: "⚠️",
  };

  private static readonly GALACTIC_ICONS: Record<NotificationType, string> = {
    success: "⭐",
    error: "💥",
    info: "🔭",
    warning: "🚨",
  };

  private static readonly TYPE_CLASSES: Record<NotificationType, string> = {
    success: "toast-success",
    error: "toast-error",
    info: "toast-info",
    warning: "toast-warning",
  };

  constructor(public readonly props: Readonly<NotificationProps>) {}

  get message(): string {
    return this.props.message;
  }

  get type(): NotificationType {
    return this.props.type ?? "info";
  }

  get duration(): number {
    return this.props.duration ?? 5000;
  }

  get useGalacticMode(): boolean {
    return this.props.useGalacticMode ?? false;
  }

  get icon(): string {
    return this.useGalacticMode
      ? Notification.GALACTIC_ICONS[this.type]
      : Notification.ICONS[this.type];
  }

  get typeClass(): string {
    return Notification.TYPE_CLASSES[this.type];
  }

  cssClass(visible: boolean = true): string {
    return `toast-notification ${this.typeClass}${visible ? " visible" : ""}`;
  }

  static achievement(title: string, useGalacticMode: boolean = false): Notification {
    return new Notification({
      message: `Achievement unlocked: ${title}`,
      type: "success",
      useGalacticMode,
    });
  }
}
