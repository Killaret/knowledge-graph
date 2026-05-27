import { chromium } from 'playwright';

function stringHash(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
}

function getVariation(nodeId, type) {
  const hash = stringHash(nodeId);
  const hash1 = hash % 1000;
  const hash2 = (hash >> 10) % 1000;
  const hash3 = (hash >> 20) % 1000;
  const isCompactType = ['star','planet','satellite','moon'].includes(type);
  const sizeMin = isCompactType ? 0.8 : 0.7;
  const sizeMax = isCompactType ? 1.2 : 1.3;
  const sizeMultiplier = sizeMin + (hash1/1000)*(sizeMax-sizeMin);
  const hueShift = (hash2/1000)*20 - 10;
  const phaseShift = (hash3/1000)*2*Math.PI;
  return { sizeMultiplier, hueShift, phaseShift, hash };
}

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const base = process.env.FRONTEND_URL || 'http://localhost:5173';
  const url = `${base.replace(/\/$/, '')}/test/isolated-node?type=planet&stableRender=true`;
  console.log('Opening', url);
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  await page.goto(url, { waitUntil: 'networkidle' });
  await page.waitForSelector('canvas', { timeout: 10000 });
  // compute variation locally
  const v = getVariation('test-node', 'planet');
  console.log('Local variation for test-node:', v);

  // find canvas central pixel grid colors
  const colors = await page.evaluate(() => {
    const canvas = document.querySelector('canvas');
    const rect = canvas.getBoundingClientRect();
    const cx = Math.round(rect.width/2);
    const cy = Math.round(rect.height/2);
    const ctx = canvas.getContext('2d');
    const out = {};
    for (let dy = -8; dy <= 8; dy++) {
      for (let dx = -8; dx <= 8; dx++) {
        const px = cx + dx;
        const py = cy + dy;
        const d = ctx.getImageData(px, py, 1, 1).data;
        const key = d.join(',');
        out[key] = (out[key] || 0) + 1;
      }
    }
    return out;
  });

  console.log('Central pixel color counts:', colors);

  await browser.close();
})();
