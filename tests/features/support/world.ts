import { setWorldConstructor, World } from '@cucumber/cucumber';
import { Browser, Page, chromium, firefox, webkit } from '@playwright/test';

export interface ITestWorld extends World {
  browser: Browser;
  page: Page;
  context: any;
  step: (text: string) => Promise<any>;
}

class CustomWorld extends World implements ITestWorld {
  browser!: Browser;
  page!: Page;
  context: any;

  async setup() {
    // Use chromium by default, can be configured via environment
    const browserType = process.env.BROWSER || 'chromium';
    const browserLauncher = browserType === 'firefox' ? firefox :
                           browserType === 'webkit' ? webkit : chromium;

    this.browser = await browserLauncher.launch({
      headless: process.env.HEADLESS !== 'false',
      slowMo: parseInt(process.env.SLOW_MO || '0'),
    });

    this.context = await this.browser.newContext({
      viewport: { width: 1280, height: 720 },
      deviceScaleFactor: 1,
    });

    this.page = await this.context.newPage();

    // Set default timeout
    this.page.setDefaultTimeout(parseInt(process.env.DEFAULT_TIMEOUT || '10000'));
  }

  async teardown() {
    if (this.context) {
      await this.context.close();
    }
    if (this.browser) {
      await this.browser.close();
    }
  }
}

setWorldConstructor(CustomWorld);

export default CustomWorld;
