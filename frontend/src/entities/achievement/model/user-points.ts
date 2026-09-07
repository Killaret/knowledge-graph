/**
 * UserPoints — Value Object that wraps a user's achievement progress.
 *
 * Centralizes total point computation, consistency checks and a simple level
 * calculation that was previously scattered across API responses and preload
 * service shape definitions.
 */

import { Achievement, type AchievementApiData } from "./achievement";

export interface UserPointsApiData {
  achievements?: AchievementApiData[];
  total_points?: number;
}

export interface UserPointsProps {
  achievements: Achievement[];
  reportedTotal?: number;
}

export class UserPoints {
  private static LEVEL_THRESHOLDS = [0, 50, 150, 300, 500, 1000];

  constructor(
    public readonly achievements: Achievement[],
    public readonly reportedTotal?: number
  ) {}

  get computedTotal(): number {
    return this.achievements.filter((a) => a.earned).reduce((sum, a) => sum + a.points, 0);
  }

  /** Total points to display. Prefers server-reported value when available. */
  get total(): number {
    return this.reportedTotal ?? this.computedTotal;
  }

  /** True when the reported total matches the sum of earned achievements. */
  get isConsistent(): boolean {
    return this.reportedTotal === undefined || this.reportedTotal === this.computedTotal;
  }

  get earnedCount(): number {
    return this.achievements.filter((a) => a.earned).length;
  }

  get totalCount(): number {
    return this.achievements.length;
  }

  get level(): number {
    const thresholds = UserPoints.LEVEL_THRESHOLDS;
    for (let i = thresholds.length - 1; i >= 0; i--) {
      if (this.total >= thresholds[i]) return i + 1;
    }
    return 1;
  }

  /** Points needed to reach the next level, or 0 at max level. */
  get pointsToNextLevel(): number {
    const thresholds = UserPoints.LEVEL_THRESHOLDS;
    const next = thresholds.find((t) => this.total < t);
    return next === undefined ? 0 : next - this.total;
  }

  static fromApi(data: UserPointsApiData): UserPoints {
    const achievements = (data.achievements ?? []).map((a) => Achievement.fromApi(a));
    return new UserPoints(achievements, data.total_points);
  }

  static empty(): UserPoints {
    return new UserPoints([], 0);
  }
}
