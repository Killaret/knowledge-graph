import type { Locale, Mode, LegacyCategory } from "$shared/utils/galactic-lexicon";
import { getLexiconMessage } from "$shared/utils/galactic-lexicon";
import { getCurrentLocale, setLocale as setUiLocale } from "$shared/utils/i18n";

let currentMode: Mode = "standard";
const subscribers = new Set<(value: Mode) => void>();

function notify(value: Mode) {
  subscribers.forEach((fn) => fn(value));
}

function setModeValue(value: Mode) {
  currentMode = value;
  notify(currentMode);
}

export const mode = {
  subscribe(fn: (value: Mode) => void) {
    fn(currentMode);
    subscribers.add(fn);
    return () => subscribers.delete(fn);
  },
  set: setModeValue,
  update(fn: (value: Mode) => Mode) {
    setModeValue(fn(currentMode));
  },
};

export function setLocale(l: Locale) {
  setUiLocale(l);
}

export function setMode(m: Mode) {
  mode.set(m);
}

export async function getMessage(category: string, key: string, ...params: unknown[]) {
  const locale = getCurrentLocale();
  let md: Mode = "standard";
  mode.subscribe((m) => (md = m))();
  return getLexiconMessage(locale, md, category as LegacyCategory, key, ...params);
}
