/**
 * Achievement — Entity representing a user achievement.
 *
 * Normalizes the several API shapes currently used across the frontend
 * (v1/achievements and v1/users/me/achievements) and centralizes status
 * helpers such as unlock detection and "new" notifications.
 */

export interface AchievementApiData {
  id?: string;
  code?: string;
  title?: string;
  name_en?: string;
  name_ru?: string;
  description?: string;
  description_en?: string;
  description_ru?: string;
  icon?: string;
  icon_emoji?: string;
  points?: number;
  earned?: boolean;
  is_hidden?: boolean;
  unlocked_at?: string | null;
  obtained_at?: string | null;
  notification_seen?: boolean;
}

export interface AchievementProps {
  id: string;
  code: string;
  title: string;
  titleRu: string;
  description: string;
  descriptionRu: string;
  icon: string;
  points: number;
  earned: boolean;
  hidden: boolean;
  obtainedAt: string | null;
  notificationSeen: boolean;
}

export class Achievement {
  constructor(public readonly props: Readonly<AchievementProps>) {}

  get id(): string {
    return this.props.id;
  }
  get code(): string {
    return this.props.code;
  }
  get title(): string {
    return this.props.title;
  }
  get titleRu(): string {
    return this.props.titleRu;
  }
  get description(): string {
    return this.props.description;
  }
  get descriptionRu(): string {
    return this.props.descriptionRu;
  }
  get icon(): string {
    return this.props.icon;
  }
  get points(): number {
    return this.props.points;
  }
  get earned(): boolean {
    return this.props.earned;
  }
  get hidden(): boolean {
    return this.props.hidden;
  }
  get obtainedAt(): string | null {
    return this.props.obtainedAt;
  }
  get notificationSeen(): boolean {
    return this.props.notificationSeen;
  }

  isUnlocked(): boolean {
    return this.props.earned;
  }

  isNew(): boolean {
    return this.props.earned && !this.props.notificationSeen;
  }

  getTitle(locale: "en" | "ru" = "en"): string {
    if (locale === "ru" && this.props.titleRu) return this.props.titleRu;
    return this.props.title;
  }

  getDescription(locale: "en" | "ru" = "en"): string {
    if (locale === "ru" && this.props.descriptionRu) return this.props.descriptionRu;
    return this.props.description;
  }

  markSeen(): Achievement {
    return new Achievement({ ...this.props, notificationSeen: true });
  }

  static fromApi(data: AchievementApiData): Achievement {
    const earned = data.earned ?? (!!data.unlocked_at || !!data.obtained_at);
    const title = data.title ?? data.name_en ?? data.code ?? "Unknown achievement";
    const titleRu = data.name_ru ?? title;
    const description = data.description ?? data.description_en ?? "";
    const descriptionRu = data.description_ru ?? description;
    const icon = data.icon ?? data.icon_emoji ?? "🏆";

    return new Achievement({
      id: data.id ?? data.code ?? "",
      code: data.code ?? "",
      title,
      titleRu,
      description,
      descriptionRu,
      icon,
      points: data.points ?? 0,
      earned,
      hidden: data.is_hidden ?? false,
      obtainedAt: data.unlocked_at ?? data.obtained_at ?? null,
      notificationSeen: data.notification_seen ?? false,
    });
  }

  static empty(): Achievement {
    return new Achievement({
      id: "",
      code: "",
      title: "",
      titleRu: "",
      description: "",
      descriptionRu: "",
      icon: "🏆",
      points: 0,
      earned: false,
      hidden: false,
      obtainedAt: null,
      notificationSeen: false,
    });
  }
}
