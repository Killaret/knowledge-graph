import { writable } from 'svelte/store'
import type { Locale, Mode } from '$shared/utils/galactic-lexicon'
import { getLexiconMessage } from '$shared/utils/galactic-lexicon'

export const locale = writable<Locale>('ru')
export const mode = writable<Mode>('standard')

export function setLocale(l: Locale) { locale.set(l) }
export function setMode(m: Mode) { mode.set(m) }

export async function getMessage(category: string, key: string, ...params: any[]) {
  let loc: Locale = 'ru'
  let md: Mode = 'standard'
  locale.subscribe(v => loc = v)()
  mode.subscribe(v => md = v)()
  // category must match types used in lexicon
  return getLexiconMessage(loc, md, category as any, key, ...params)
}
