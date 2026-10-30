const fs = require('fs');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');
const configDir = path.join(rootDir, 'config');
const outputPath = path.join(rootDir, 'knowledge-graph.config.json');

function deepMerge(target, source) {
  for (const key of Object.keys(source)) {
    if (Object.prototype.hasOwnProperty.call(target, key)) {
      if (isObject(target[key]) && isObject(source[key])) {
        target[key] = deepMerge({ ...target[key] }, source[key]);
      } else {
        target[key] = source[key];
      }
    } else {
      target[key] = source[key];
    }
  }
  return target;
}

function isObject(value) {
  return value && typeof value === 'object' && !Array.isArray(value);
}

function loadConfigParts(dir) {
  const files = fs.readdirSync(dir).filter((name) => name.endsWith('.json')).sort();
  if (files.length === 0) {
    throw new Error(`No JSON config files found in ${dir}`);
  }

  let result = {};
  for (const file of files) {
    const filePath = path.join(dir, file);
    const raw = fs.readFileSync(filePath, 'utf8');
    let parsed;
    try {
      parsed = JSON.parse(raw);
    } catch (error) {
      throw new Error(`Failed to parse ${filePath}: ${error.message}`);
    }
    if (!isObject(parsed)) {
      throw new Error(`${filePath} must contain a JSON object at the top level`);
    }
    result = deepMerge(result, parsed);
  }
  return result;
}

function main() {
  if (!fs.existsSync(configDir)) {
    throw new Error(`Config folder not found: ${configDir}`);
  }

  const merged = loadConfigParts(configDir);
  const content = JSON.stringify(merged, null, 2) + '\n';
  fs.writeFileSync(outputPath, content, 'utf8');
  console.log(`Generated ${path.relative(rootDir, outputPath)} from ${configDir}`);
}

main();
