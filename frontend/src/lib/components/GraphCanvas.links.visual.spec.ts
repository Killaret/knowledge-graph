/**
 * Visual correctness tests for GraphCanvas links
 * These tests verify that links are rendered correctly between nodes
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';
import * as renderer from './GraphCanvas/renderer';
import type { SimulationNode, SimulationLink } from './GraphCanvas/types';

describe('GraphCanvas - Link Visual Correctness', () => {
  describe('Link drawing functions', () => {
    it('should draw link from source to target coordinates', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: '',
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = { id: '1', title: 'Source', x: 100, y: 100, type: 'star' };
      const targetNode: SimulationNode = { id: '2', title: 'Target', x: 200, y: 200, type: 'planet' };
      const link: SimulationLink = { source: '1', target: '2', weight: 0.5 };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      // Verify line starts at source and ends at target
      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      expect(ctx.lineTo).toHaveBeenCalledWith(200, 200);
      expect(ctx.stroke).toHaveBeenCalled();
    });

    it('should use correct colors for reference links', () => {
      const color = renderer.getLinkColor(0.5, 'reference');
      expect(color).toContain('51, 102, 255'); // Blue reference color
    });

    it('should use correct colors for dependency links', () => {
      const color = renderer.getLinkColor(0.5, 'dependency');
      expect(color).toContain('255, 102, 0'); // Orange dependency color
    });

    it('should use correct colors for related links', () => {
      const color = renderer.getLinkColor(0.5, 'related');
      expect(color).toContain('153, 153, 153'); // Gray related color
    });

    it('should use correct colors for custom links', () => {
      const color = renderer.getLinkColor(0.5, 'custom');
      expect(color).toContain('255, 102, 255'); // Pink custom color
    });

    it('should apply opacity based on weight', () => {
      const lowWeightColor = renderer.getLinkColor(0.2, 'reference');
      const highWeightColor = renderer.getLinkColor(0.8, 'reference');

      expect(lowWeightColor).toContain('0.48'); // 0.4 + 0.2*0.4
      expect(highWeightColor).toContain('0.72'); // 0.4 + 0.8*0.4
    });
  });

  describe('Line dash patterns', () => {
    it('should use solid line for reference links', () => {
      const dash = renderer.getLineDash('reference', 0.5);
      expect(dash).toEqual([]);
    });

    it('should use dash-dot pattern for dependency links', () => {
      const dash = renderer.getLineDash('dependency', 0.5);
      expect(dash).toEqual([10, 3]);
    });

    it('should use dashed line for weak related links', () => {
      const dash = renderer.getLineDash('related', 0.2);
      expect(dash).toEqual([6, 4]);
    });

    it('should use solid line for strong related links', () => {
      const dash = renderer.getLineDash('related', 0.5);
      expect(dash).toEqual([]);
    });

    it('should use dotted pattern for custom links', () => {
      const dash = renderer.getLineDash('custom', 0.5);
      expect(dash).toEqual([2, 6]);
    });
  });

  describe('Link width calculation', () => {
    it('should apply thicker line for dependency links', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: '',
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = { id: '1', title: 'Source', x: 100, y: 100, type: 'star' };
      const targetNode: SimulationNode = { id: '2', title: 'Target', x: 200, y: 200, type: 'planet' };
      const link: SimulationLink = { source: '1', target: '2', weight: 0.5, link_type: 'dependency' };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      // Dependency uses 1.5x thickness: max(1, 0.5*4) * 1.5 = 3
      expect(ctx.lineWidth).toBe(3);
    });

    it('should apply thinner line for reference links', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: '',
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = { id: '1', title: 'Source', x: 100, y: 100, type: 'star' };
      const targetNode: SimulationNode = { id: '2', title: 'Target', x: 200, y: 200, type: 'planet' };
      const link: SimulationLink = { source: '1', target: '2', weight: 0.5, link_type: 'reference' };

      renderer.drawLink(ctx, link, sourceNode, targetNode);

      // Reference uses 0.8x thickness: max(1, 0.5*4) * 0.8 = 1.6
      expect(ctx.lineWidth).toBeCloseTo(1.6, 1);
    });
  });

  describe('drawAllLinks', () => {
    it('should skip drawing when source node not found', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
      } as unknown as CanvasRenderingContext2D;

      const nodes: SimulationNode[] = [
        { id: '2', title: 'Target', x: 200, y: 200, type: 'planet' }
      ];
      const links: SimulationLink[] = [
        { source: '1', target: '2', weight: 0.5 } // Source '1' not in nodes
      ];

      renderer.drawAllLinks(ctx, links, nodes);

      expect(ctx.stroke).not.toHaveBeenCalled();
    });

    it('should skip drawing when target node not found', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
      } as unknown as CanvasRenderingContext2D;

      const nodes: SimulationNode[] = [
        { id: '1', title: 'Source', x: 100, y: 100, type: 'star' }
      ];
      const links: SimulationLink[] = [
        { source: '1', target: '2', weight: 0.5 } // Target '2' not in nodes
      ];

      renderer.drawAllLinks(ctx, links, nodes);

      expect(ctx.stroke).not.toHaveBeenCalled();
    });

    it('should draw links when source/target are node objects', () => {
      const ctx = {
        beginPath: vi.fn(),
        moveTo: vi.fn(),
        lineTo: vi.fn(),
        stroke: vi.fn(),
        setLineDash: vi.fn(),
        lineWidth: 1,
        strokeStyle: '',
      } as unknown as CanvasRenderingContext2D;

      const sourceNode: SimulationNode = { id: '1', title: 'Source', x: 100, y: 100, type: 'star' };
      const targetNode: SimulationNode = { id: '2', title: 'Target', x: 200, y: 200, type: 'planet' };
      const links: SimulationLink[] = [
        { source: sourceNode, target: targetNode, weight: 0.5 }
      ];

      renderer.drawAllLinks(ctx, links, []);

      expect(ctx.moveTo).toHaveBeenCalledWith(100, 100);
      expect(ctx.lineTo).toHaveBeenCalledWith(200, 200);
      expect(ctx.stroke).toHaveBeenCalled();
    });
  });
});
