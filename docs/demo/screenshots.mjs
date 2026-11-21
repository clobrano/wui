// Take the web GUI screenshots used in the README.
//
// Usage (needs Node and Playwright, e.g. `npm install playwright`):
//   docs/demo/seed.sh /tmp/wui-demo
//   HOME=/tmp/wui-demo wui gui --no-browser --port 7108 --api-port 7107 &
//   node docs/demo/screenshots.mjs http://localhost:7108 <uuid-of-task-1>
//
// Writes docs/demo/gui-*.png. Set CHROMIUM to use a specific browser binary.
import { chromium } from 'playwright';

const [base = 'http://localhost:7108', uuid] = process.argv.slice(2);
const out = new URL('.', import.meta.url).pathname;

const browser = await chromium.launch({ executablePath: process.env.CHROMIUM || undefined });

async function shot(name, path, viewport) {
  const page = await browser.newPage({ viewport, deviceScaleFactor: 2, colorScheme: 'dark' });
  await page.goto(base + path, { waitUntil: 'networkidle' });
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${out}gui-${name}.png` });
  await page.close();
}

const desktop = { width: 1200, height: 720 };
const phone = { width: 390, height: 780 };

await shot('desktop', '/', desktop);
if (uuid) await shot('detail', `/tasks/${uuid}`, desktop);
await shot('mobile', '/', phone);

await browser.close();
