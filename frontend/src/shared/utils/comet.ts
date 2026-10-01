// COMET-1: чистые преобразования полей кометы — изолированно от компонентов,
// чтобы тестировать без рендера.

export const REMIND_CHOICES = ["none", "15m", "1h", "1d", "custom"] as const;
export type RemindChoice = (typeof REMIND_CHOICES)[number];

const REMIND_SECONDS: Record<Exclude<RemindChoice, "none" | "custom">, number> = {
  "15m": 15 * 60,
  "1h": 60 * 60,
  "1d": 24 * 60 * 60,
};

/** datetime-local value ("YYYY-MM-DDTHH:MM") → RFC 3339 UTC. "" → null. */
export function localInputToIso(value: string): string | null {
  if (!value) return null;
  const t = new Date(value);
  if (Number.isNaN(t.getTime())) return null;
  return t.toISOString();
}

/** RFC 3339 → datetime-local value in the browser timezone. null → "". */
export function isoToLocalInput(iso: string | null | undefined): string {
  if (!iso) return "";
  const t = new Date(iso);
  if (Number.isNaN(t.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${t.getFullYear()}-${pad(t.getMonth() + 1)}-${pad(t.getDate())}T${pad(t.getHours())}:${pad(t.getMinutes())}`;
}

/** Выбор напоминания → секунды до даты. none → null, custom → минуты*60. */
export function remindChoiceToSeconds(choice: RemindChoice, customMinutes: number): number | null {
  if (choice === "none") return null;
  if (choice === "custom") {
    const minutes = Math.max(0, Math.floor(customMinutes));
    return minutes > 0 ? minutes * 60 : null;
  }
  return REMIND_SECONDS[choice];
}

/** Секунды → выбор селекта (для подстановки сохранённого значения). */
export function secondsToRemindChoice(seconds: number | null | undefined): {
  choice: RemindChoice;
  customMinutes: number;
} {
  if (seconds == null || seconds <= 0) return { choice: "none", customMinutes: 60 };
  for (const [key, sec] of Object.entries(REMIND_SECONDS)) {
    if (sec === seconds) {
      return { choice: key as RemindChoice, customMinutes: 60 };
    }
  }
  return { choice: "custom", customMinutes: Math.round(seconds / 60) };
}

/**
 * COMET-1 stage E: визуальная метафора «приближается».
 * urgency — 0..1: чем ближе срок, тем выше. done/просроченная — тусклая (0).
 * Без даты — null (рисуем как обычную комету).
 * Полное приближение считаем за 7 дней до срока.
 */
export function cometUrgency(
  dueAt: string | null | undefined,
  doneAt: string | null | undefined,
  now: number = Date.now()
): number | null {
  if (doneAt) return 0;
  if (!dueAt) return null;
  const due = new Date(dueAt).getTime();
  if (Number.isNaN(due)) return null;
  if (due <= now) return 0; // прошла — тусклая
  const week = 7 * 24 * 60 * 60 * 1000;
  const left = due - now;
  return Math.min(1, Math.max(0, 1 - left / week));
}

/** Комета «прошла» или «сделано» — можно предлагать архив (debris). */
export function isCometSettled(
  dueAt: string | null | undefined,
  doneAt: string | null | undefined
): boolean {
  if (doneAt) return true;
  if (!dueAt) return false;
  const due = new Date(dueAt).getTime();
  return !Number.isNaN(due) && due <= Date.now();
}
