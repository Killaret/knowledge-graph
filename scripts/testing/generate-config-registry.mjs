// Generated config registry for docs/operations/CONFIG_REGISTRY.md (CONFIG-AUDIT-1).
//
// For every key in config/*.json records:
//   - who reads it: the Go accessor `j.<FieldPath>` in backend
//     internal/config/config.go and services/graph-service/internal/config/config.go
//     (json tags resolved to config paths), or frontend/src usage for frontend.*/ci_cd.*
//   - env override found on the same resolveConfig statement
//   - compose files that set the env var
//   - presence in docs/operations/CONFIGURATION_EN.md / _RU.md
//   - status: «мёртвый» when no code reads the key
//
// Run: node scripts/testing/generate-config-registry.mjs .
// Stdout (for the drift checker): --check

import { readFileSync, readdirSync, writeFileSync, existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { execSync } from "node:child_process";

const repoRoot = resolve(process.argv[2] ?? ".");

const configDir = join(repoRoot, "config");
const outPath = join(repoRoot, "docs", "operations", "CONFIG_REGISTRY.md");

function flat(obj, prefix = "", out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    if (v && typeof v === "object" && !Array.isArray(v)) flat(v, prefix + k + ".", out);
    else out[prefix + k] = v;
  }
  return out;
}

const keys = [];
for (const f of readdirSync(configDir).filter((n) => n.endsWith(".json")).sort()) {
  const obj = JSON.parse(readFileSync(join(configDir, f), "utf8"));
  for (const [k, v] of Object.entries(flat(obj))) keys.push({ file: f, key: k, value: v });
}

// Parse `type X struct { ... }` into a tree; a struct's json tag sits on its
// closing line (`} `json:"y"``), so assign tags when nodes close.
function parseJSONConfigKeys(src, structName) {
  const start = src.indexOf(`type ${structName} struct {`);
  if (start < 0) return new Map();
  const body = src.slice(src.indexOf("{", start));
  const root = { goName: structName, tag: null, children: [] };
  const stack = [root];
  for (const raw of body.split("\n").slice(1)) {
    const line = raw.trim();
    const opens = (line.match(/\{/g) || []).length;
    const closes = (line.match(/\}/g) || []).length;
    const tagM = line.match(/json:"([^"]+)"/);
    if (/^\w+\s+struct\s*\{/.test(line)) {
      const node = { goName: line.match(/^(\w+)/)[1], tag: null, children: [] };
      stack[stack.length - 1].children.push(node);
      stack.push(node);
    } else if (closes > 0) {
      // leaf on the same line as a close is not possible in this file style
      if (stack.length === 1) break; // closing the JSONConfig struct itself
      for (let c = 0; c < closes && stack.length > 1; c++) {
        const n = stack.pop();
        n.tag = n.tag ?? tagM?.[1] ?? n.goName.toLowerCase();
      }
    } else if (tagM) {
      stack[stack.length - 1].children.push({
        goName: line.match(/^(\w+)/)[1],
        tag: tagM[1],
        children: [],
      });
    }
  }
  const map = new Map();
  const walk = (node, jsonPrefix, goPrefix) => {
    for (const ch of node.children) {
      const tag = ch.tag ?? ch.goName.toLowerCase();
      const jp = jsonPrefix ? `${jsonPrefix}.${tag}` : tag;
      const gp = goPrefix ? `${goPrefix}.${ch.goName}` : ch.goName;
      if (ch.children.length) walk(ch, jp, gp);
      else map.set(jp, gp);
    }
  };
  walk(root, "", "");
  return map;
}

// Consumers: `Field: get*Env("ENV", getJSON*(jsonCfg, func...{ return j.Path }`
// and JSON-only `Field: getJSON*(jsonCfg, ... j.Path` / `Field: resolve*(jsonCfg...)`.
const CONSUMER_RE = /(\w+):\s*.*?\b(?:getEnv|getBoolEnv|getIntEnv|getFloatEnv|getUint32Env|getUint8Env)\("([A-Z0-9_]+)".*?j\.([A-Za-z0-9_.]+)/g;
const CONSUMER_NOENV_RE = /(\w+)\s*[:=]\s*(?:getJSON\w+|resolve\w+)\(\s*jsonCfg.*?j\.([A-Za-z0-9_.]+)/g;
// any other `j.<Path>` occurrence (helper functions like resolveGammaLinkMinScore)
const ANY_ACCESSOR_RE = /j\.([A-Za-z0-9_.]+)/g;

function parseConsumers(src) {
  const out = new Map();
  let m;
  while ((m = CONSUMER_RE.exec(src))) out.set(m[3], { field: m[1], env: m[2] });
  while ((m = CONSUMER_NOENV_RE.exec(src))) {
    if (!out.has(m[2])) out.set(m[2], { field: m[1], env: null });
  }
  // fallback: accessor used inside a helper — report the enclosing func name
  while ((m = ANY_ACCESSOR_RE.exec(src))) {
    if (out.has(m[1])) continue;
    const before = src.slice(0, m.index);
    const fm = before.match(/func (?:\(\w+ \*?\w+\) )?(\w+)[^{]*\{(?:(?!func )[\s\S])*$/);
    const fname = fm ? fm[1] : "?";
    out.set(m[1], { field: fname === "loadJSONConfigParts" || fname === "loadJSONConfig" ? null : fname, env: null, helper: true });
  }
  return out;
}

const backendSrc = readFileSync(join(repoRoot, "backend/internal/config/config.go"), "utf8");
const gsSrc = readFileSync(join(repoRoot, "services/graph-service/internal/config/config.go"), "utf8");

const backendTags = parseJSONConfigKeys(backendSrc, "JSONConfig");
const gsTags = parseJSONConfigKeys(gsSrc, "JSONConfig");
const backendConsumers = parseConsumers(backendSrc);
const gsConsumers = parseConsumers(gsSrc);

// ── Frontend readers ───────────────────────────────────────────────
const feConfigSrc = readFileSync(join(repoRoot, "frontend/src/shared/config/config.ts"), "utf8");
const aliasRe = /export const (\w+)\s*=\s*config\.([A-Za-z0-9_.\[\]"]+)/g;
const aliases = [];
let am;
while ((am = aliasRe.exec(feConfigSrc))) {
  aliases.push({ name: am[1], prefix: am[2].replace(/\["([^"]+)"\]/g, ".$1") });
}

function rgFiles(pattern, dir) {
  try {
    return execSync(`rg -l --no-messages "${pattern}" "${dir}"`, { cwd: repoRoot, encoding: "utf8" })
      .split("\n").filter(Boolean).map((p) => p.replace(/\\/g, "/"));
  } catch { return []; }
}

function frontendReaders(key) {
  const readers = new Set();
  const skip = (f) => f.endsWith("shared/config/config.ts");
  for (const f of rgFiles(key.replace(/\./g, "\\."), "frontend/src")) if (!skip(f)) readers.add(f);
  for (const a of aliases) {
    if (!key.startsWith(a.prefix + ".") && key !== a.prefix) continue;
    const suffix = key === a.prefix ? "" : key.slice(a.prefix.length + 1);
    if (!suffix) {
      for (const f of rgFiles(`\\b${a.name}\\b`, "frontend/src")) if (!skip(f)) readers.add(f);
      continue;
    }
    for (const f of rgFiles(`${a.name}\\.${suffix.replace(/\./g, "\\.")}`, "frontend/src")) if (!skip(f)) readers.add(f);
    // first suffix segment — the intermediate object may be passed whole
    for (const f of rgFiles(`${a.name}\\.${suffix.split(".")[0]}\\b`, "frontend/src")) if (!skip(f)) readers.add(f);
    // the alias object may be rebound (this.performanceConfig.x): count a file
    // only if it mentions the alias AND the leaf token — otherwise a dead leaf
    // would be marked read just because its siblings are.
    const leaf = key.split(".").pop();
    for (const f of rgFiles(`\\b${a.name}\\b`, "frontend/src")) {
      if (skip(f)) continue;
      try {
        if (new RegExp(`\\b${leaf}\\b`).test(readFileSync(join(repoRoot, f), "utf8"))) readers.add(`${f} (через ${a.name})`);
      } catch { /* unreadable */ }
    }
  }
  // direct export of exactly this key: `export const X = config.<key>`
  const exportRe = new RegExp(`export const (\\w+)\\s*=\\s*config\\.${key.replace(/\./g, "\\.")}\\b`);
  const em = feConfigSrc.match(exportRe);
  if (em) for (const f of rgFiles(`\\b${em[1]}\\b`, "frontend/src")) if (!skip(f)) readers.add(f);
  return [...readers].sort();
}

// ── Compose / docs ─────────────────────────────────────────────────
const composeFiles = readdirSync(repoRoot).filter((f) => /^docker-compose.*\.yml$/.test(f));
const composeText = Object.fromEntries(composeFiles.map((f) => [f, readFileSync(join(repoRoot, f), "utf8")]));
const docsEn = readFileSync(join(repoRoot, "docs/operations/CONFIGURATION_EN.md"), "utf8");
const docsRu = existsSync(join(repoRoot, "docs/operations/CONFIGURATION_RU.md"))
  ? readFileSync(join(repoRoot, "docs/operations/CONFIGURATION_RU.md"), "utf8")
  : "";

const BAKE = "нет (вшивается при сборке)";
function composeHits(env) {
  if (!env) return "—";
  const hits = composeFiles.filter((f) => composeText[f].includes(env + ":") || composeText[f].includes(env + "="));
  return hits.length ? hits.map((f) => f.replace("docker-compose", "compose").replace(".yml", "")).join(", ") : "—";
}

// nlp.* keys consumed by the Python service: app/config.py seeds nlp.* into
// environment variables (env > file > default) before readers run.
const nlpSeedPath = join(repoRoot, "nlp-service/app/config.py");
const nlpSeedMap = new Map();
if (existsSync(nlpSeedPath)) {
  const src = readFileSync(nlpSeedPath, "utf8");
  for (const m of src.matchAll(/"([a-z_]+)":\s*"([A-Z_]+)"/g)) nlpSeedMap.set(m[1], m[2]);
}

const rows = [];
for (const { file, key, value } of keys) {
  let reader = "—";
  let env = "—";
  let status = "ok";
  if (key.startsWith("graph_service.")) {
    const goPath = gsTags.get(key);
    const c = goPath && gsConsumers.get(goPath);
    if (c) { reader = `graph-service config.go → ${c.field}`; env = c.env ?? "—"; }
    else { status = "мёртвый"; if (goPath) reader = "graph-service: поле есть, не читается"; }
  } else if (key.startsWith("frontend.") || key.startsWith("ci_cd.")) {
    const readers = frontendReaders(key);
    if (readers.length) reader = readers.map((r) => r.replace(/^frontend\//, "")).join(", ");
    else status = "мёртвый";
    env = BAKE;
  } else {
    const goPath = backendTags.get(key);
    let c = goPath && backendConsumers.get(goPath);
    let via = null;
    if (!c) {
      // leaf inside a consumed map/list: backend.rate_limit.endpoints.x lives
      // via the Endpoints accessor one level up — walk ancestors.
      let p = key;
      while (p.includes(".")) {
        p = p.slice(0, p.lastIndexOf("."));
        const anc = backendTags.get(p);
        if (anc && backendConsumers.has(anc)) { c = backendConsumers.get(anc); via = p; break; }
      }
    }
    if (c) { reader = `backend config.go → ${c.field}${via ? ` (через ${via})` : ""}`; env = c.env ?? "—"; }
    else if (key.startsWith("nlp.") && key.split(".").length === 2 && nlpSeedMap.has(key.slice(4))) {
      const envName = nlpSeedMap.get(key.slice(4));
      reader = `nlp-service app/config.py → ${envName}`;
      env = envName;
    }
    else { status = "мёртвый"; if (goPath) reader = "backend: поле есть, не читается"; }
  }
  const envForDocs = env === "—" || env === BAKE ? null : env;
  const docsHit = (envForDocs && (docsEn.includes(envForDocs) || docsRu.includes(envForDocs)))
    || docsEn.includes(key) || docsRu.includes(key);
  rows.push({
    file, key, value, reader, env,
    compose: envForDocs ? composeHits(envForDocs) : "—",
    docs: docsHit ? "+" : "—",
    status,
  });
}

const lines = [
  "# CONFIG_REGISTRY — реестр ключей config/*.json",
  "",
  "Сгенерировано `scripts/testing/generate-config-registry.mjs`; проверка дрейфа — `check-config-registry.mjs`.",
  "Порядок приоритета: **env > knowledge-graph.config.json (или config/*.json) > дефолт в коде** — backend,",
  "graph-service и NLP (nlp.* → env через `app/config.py` при старте); фронтенд: `config/*.json` вшиваются при сборке",
  "(`npm run build-config` + `npm run build`), переменных окружения в рантайме нет — смена `config/*.json`",
  "без пересборки фронтенда не действует.",
  "",
  "| Ключ | Файл | Читает | Env-переопределение | Compose | В CONFIGURATION | Статус |",
  "|---|---|---|---|---|---|---|",
];
for (const r of rows) {
  const envCell = r.env === "—" || r.env === BAKE ? r.env : "`" + r.env + "`";
  lines.push(`| \`${r.key}\` | ${r.file} | ${r.reader} | ${envCell} | ${r.compose} | ${r.docs} | ${r.status} |`);
}
lines.push("");

const output = lines.join("\n");
if (process.argv.includes("--check")) {
  process.stdout.write(output);
} else {
  writeFileSync(outPath, output);
  console.log(`Generated ${rows.length} registry entries at docs/operations/CONFIG_REGISTRY.md`);
  const dead = rows.filter((r) => r.status === "мёртвый");
  if (dead.length) console.log(`Dead keys (${dead.length}): ${dead.map((r) => r.key).join(", ")}`);
}
