/**
 * Isolated visual regression tests for graph rendering
 * Uses query params to test different node/link types
 * @visual @regression @isolated
 */

import { test, expect } from '@playwright/test';
import { createNote, createLink, deleteNote } from './helpers/testData';

test.describe.configure({ mode: 'serial' });

test.describe('GraphCanvas Visual - Isolated Node Types @visual @isolated', () => {
  const nodeTypes = [
    { type: 'star' },
    { type: 'planet' },
    { type: 'comet' },
    { type: 'galaxy' },
    { type: 'asteroid' }
  ];

  for (const { type } of nodeTypes) {
    test(`should render ${type} node correctly`, async ({ page, request }) => {
      // Create a single note via API
      const note = await createNote(request, {
        title: `Visual Test ${type}`,
        content: `Testing ${type} node rendering`,
        type: type
      });
      
      try {
        // Navigate to test page (public route, no auth required)
        await page.goto(`/test/isolated-node?type=${type}`);
        
        // Wait for canvas
        await page.waitForSelector('canvas', { timeout: 10000 });
        
        // Wait for simulation
        await page.waitForTimeout(1000);
        
        // Take screenshot
        const canvas = page.locator('canvas').first();
        await expect(canvas).toHaveScreenshot(`${type}-node-visual.png`, {
          maxDiffPixels: 1000,
          threshold: 0.6,
          animations: 'disabled'
        });
      } finally {
        // Cleanup: delete the test note
        await deleteNote(request, note.data.id).catch(() => {});
      }
    });
  }
});

test.describe('GraphCanvas Visual - Link Types @visual @links', () => {
  const linkTypes = ['reference', 'dependency', 'related', 'custom'];

  for (const linkType of linkTypes) {
    test(`should render ${linkType} link correctly`, async ({ page, request }) => {
      // Create source node
      const sourceNote = await createNote(request, {
        title: 'Source Node',
        content: 'Source node for link test',
        type: 'star'
      });
      const targetNote = await createNote(request, {
        title: 'Target',
        content: 'Target node',
        type: 'planet'
      });
      
      try {
        // Create link
        await createLink(request, sourceNote.data.id, targetNote.data.id, 0.8, linkType);
        
        // Navigate to test page (public route, no auth required)
        await page.goto(`/test/link-pair?linkType=${linkType}`);
        await page.waitForSelector('canvas', { timeout: 15000 });
        await page.waitForTimeout(2000);
        
        // Take screenshot
        const canvas = page.locator('canvas').first();
        await expect(canvas).toHaveScreenshot(`${linkType}-link-visual.png`, {
          maxDiffPixels: 600,
          threshold: 0.4,
          animations: 'disabled'
        });
      } finally {
        // Cleanup
        await deleteNote(request, sourceNote.data.id).catch(() => {});
        await deleteNote(request, targetNote.data.id).catch(() => {});
      }
    });
  }
});
