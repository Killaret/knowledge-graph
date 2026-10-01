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
