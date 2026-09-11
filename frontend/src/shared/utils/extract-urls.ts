export interface ExtractedImportItem {
  title: string;
  url: string;
}

const MAX_TITLE_RUNES = 200;
const MAX_URL_RUNES = 16384;

// Splits a string at every http/https scheme boundary so that concatenated
// URLs such as "https://a.com/https://b.com" are treated as two items.
const URL_START_RE = /(?=https?:\/\/)/;
const URL_PREFIX_RE = /^https?:\/\/\S+/;

// Decodes common HTML entities in bookmark titles and URLs.
function decodeHtmlEntities(raw: string): string {
  return raw
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&#(\d+);/g, (_, code) => String.fromCodePoint(Number(code)))
    .replace(/&#x([0-9a-fA-F]+);/g, (_, hex) => String.fromCodePoint(parseInt(hex, 16)))
    .replace(/&nbsp;/g, " ");
}

function truncateToRunes(value: string, max: number): string {
  const runes = Array.from(value);
  if (runes.length <= max) return value;
  return runes.slice(0, max).join("");
}

function cleanUrl(raw: string): string {
  // Strip common trailing punctuation that may be glued to the URL.
  let url = raw.replace(/[,;.:<>"')\]}]+$/, "");
  // Netscape bookmark HTML may contain encoded ampersands.
  url = decodeHtmlEntities(url);
  return url;
}

function stripHtmlTags(raw: string): string {
  return raw.replace(/<[^>]+>/g, "").replace(/\s+/g, " ").trim();
}

function extractUrlsFrom(text: string): string[] {
  const parts = text.split(URL_START_RE);
  const urls: string[] = [];

  for (const part of parts) {
    const trimmed = part.trim();
    if (!trimmed) continue;

    const match = trimmed.match(URL_PREFIX_RE);
    if (!match) continue;

    urls.push(cleanUrl(match[0]));
  }

  return urls;
}

function lineToCandidates(line: string): ExtractedImportItem[] {
  const trimmed = line.trim();
  if (!trimmed || trimmed.startsWith("#")) {
    return [];
  }

  const pipeIdx = trimmed.indexOf("|");
  let title = "";
  let rest = trimmed;

  if (pipeIdx > 0) {
    title = trimmed.slice(0, pipeIdx).trim();
    rest = trimmed.slice(pipeIdx + 1).trim();
  }

  const urls = extractUrlsFrom(rest);
  if (urls.length === 0) {
    return [];
  }

  return urls.map((url, index) => ({
    title: truncateToRunes(index === 0 && title ? title : url, MAX_TITLE_RUNES),
    url: truncateToRunes(url, MAX_URL_RUNES),
  }));
}

export function extractURLs(input: string): ExtractedImportItem[] {
  const items: ExtractedImportItem[] = [];
  const seen = new Set<string>();

  for (const line of input.split("\n")) {
    for (const it of lineToCandidates(line)) {
      if (seen.has(it.url)) continue;
      seen.add(it.url);
      items.push(it);
    }
  }

  return items;
}

// Matches <A HREF="..." ...>title</A> in a case-insensitive way. The title may
// contain HTML entities or nested tags; we strip tags and collapse whitespace.
const BOOKMARK_ANCHOR_RE = /<A\b[^>]*href=(["'])([^"']+)\1[^>]*>([\s\S]*?)<\/A\s*>/gi;

export function extractURLsFromHTML(html: string): ExtractedImportItem[] {
  const items: ExtractedImportItem[] = [];
  const seen = new Set<string>();

  for (const match of html.matchAll(BOOKMARK_ANCHOR_RE)) {
    const rawUrl = match[2].trim();
    const title = stripHtmlTags(decodeHtmlEntities(match[3])) || rawUrl;

    if (!rawUrl) continue;
    const url = truncateToRunes(cleanUrl(rawUrl), MAX_URL_RUNES);
    if (seen.has(url)) continue;
    seen.add(url);

    items.push({ title: truncateToRunes(title, MAX_TITLE_RUNES), url });
  }

  return items;
}

export function chunk<T>(items: T[], size: number): T[][] {
  if (size <= 0) return [];
  const chunks: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    chunks.push(items.slice(i, i + size));
  }
  return chunks;
}
