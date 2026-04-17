import { setWorldConstructor, setDefaultTimeout, World, type IWorldOptions } from '@cucumber/cucumber';
import { chromium, type Browser, type BrowserContext, type Page, type APIRequestContext } from '@playwright/test';

export interface ITestWorld extends World {
  browser?: Browser;
  context?: BrowserContext;
  page: Page;
  request: APIRequestContext;
  testNotes: Array<{ id: string; title: string; type: string }>;
  centerNoteId?: string;
  currentNoteId?: string;
  initialFogDensity?: number;
  initialCameraPos?: { x: number; y: number; z: number } | null;
  clickedNodeId?: string;
}

class CustomWorld extends World implements ITestWorld {
  browser?: Browser;
  context?: BrowserContext;
  page!: Page;
  request!: APIRequestContext;
  testNotes: Array<{ id: string; title: string; type: string }> = [];
  centerNoteId?: string;
  currentNoteId?: string;
  initialFogDensity?: number;
  initialCameraPos?: { x: number; y: number; z: number } | null;
  clickedNodeId?: string;

  constructor(options: any) {
    super(options);
  }
}

setWorldConstructor(CustomWorld);
setDefaultTimeout(15000);
