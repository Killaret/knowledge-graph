import { setWorldConstructor, World } from '@cucumber/cucumber';
import type { Browser, Page, APIRequestContext } from '@playwright/test';
import { chromium, firefox, webkit } from '@playwright/test';

export interface ITestWorld extends World {
  browser: Browser;
  page: Page;
  context: any;
  request: APIRequestContext;
  step: (text: string) => Promise<any>;
  setup: () => Promise<void>;
  teardown: () => Promise<void>;
}

class CustomWorld extends World implements ITestWorld {
  browser!: Browser;
  page!: Page;
  context: any;
  request!: APIRequestContext;

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
    
    // Create API request context
    this.request = await this.context.request;

    // Set default timeout
    this.page.setDefaultTimeout(parseInt(process.env.DEFAULT_TIMEOUT || '10000'));
  }
  
  // Helper method to call other steps - stores current step text for reference
  stepText: string = '';
  async step(text: string): Promise<any> {
    this.stepText = text;
    // Steps are executed by Cucumber runtime, this is just a placeholder
    return Promise.resolve();
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
