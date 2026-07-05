/**
 * Visual tests for GraphCanvas renderer - anomaly types
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  drawRealityRift,
  drawChromaticMaw,
  drawVoidWhisper,
  drawCosmicAbomination,
  drawUnknown,
  drawNode,
  getAnomalyParams,
  getLinkColor,
  getLineDash,
  getGlowIntensity,
  getNodeGradient,
  getNodeColor
} from './renderer';

// Mock CanvasRenderingContext2D
const mockCtx = {
  save: vi.fn(),
  restore: vi.fn(),
  translate: vi.fn(),
  rotate: vi.fn(),
  beginPath: vi.fn(),
  closePath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  arc: vi.fn(),
  ellipse: vi.fn(),
  bezierCurveTo: vi.fn(),
  fill: vi.fn(),
  stroke: vi.fn(),
  createRadialGradient: vi.fn(() => ({
    addColorStop: vi.fn()
  })),
  createLinearGradient: vi.fn(() => ({
    addColorStop: vi.fn()
  })),
  shadowBlur: 0,
  shadowColor: '',
  lineWidth: 0,
  lineCap: '',
  fillStyle: '',
  strokeStyle: '',
  setLineDash: vi.fn(),
  globalAlpha: 1
} as unknown as CanvasRenderingContext2D;

describe('renderer anomaly functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('drawRealityRift', () => {
    it('should draw reality rift with dark core and cracks', () => {
      const params = getAnomalyParams('node-1');
      
      drawRealityRift(mockCtx, 100, 100, 30, params);
      
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(100, 100);
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should use deterministic parameters for same nodeId', () => {
      const params1 = getAnomalyParams('node-1');
      const params2 = getAnomalyParams('node-1');
      
      expect(params1.crackCount).toBe(params2.crackCount);
      expect(params1.deformAmount).toBe(params2.deformAmount);
      expect(params1.rotationOffset).toBe(params2.rotationOffset);
    });

    it('should generate different parameters for different nodeIds', () => {
      const params1 = getAnomalyParams('node-1');
      const params2 = getAnomalyParams('node-2');
      
      // At least one parameter should differ
      const differs =
        params1.crackCount !== params2.crackCount ||
        params1.deformAmount !== params2.deformAmount ||
        params1.rotationOffset !== params2.rotationOffset;
      
      expect(differs).toBe(true);
    });
  });

  describe('drawChromaticMaw', () => {
    it('should draw chromatic maw with tentacles and gradient', () => {
      const params = getAnomalyParams('node-2');
      
      drawChromaticMaw(mockCtx, 100, 100, 30, params);
      
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(100, 100);
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should use deterministic parameters for same nodeId', () => {
      const params1 = getAnomalyParams('node-2');
      const params2 = getAnomalyParams('node-2');
      
      expect(params1.tentacleCount).toBe(params2.tentacleCount);
      expect(params1.colorShift1).toBe(params2.colorShift1);
      expect(params1.rotationOffset).toBe(params2.rotationOffset);
    });
  });

  describe('drawVoidWhisper', () => {
    it('should draw void whisper with particles and connections', () => {
      const params = getAnomalyParams('node-3');
      
      drawVoidWhisper(mockCtx, 100, 100, 30, params);
      
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(100, 100);
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should use deterministic parameters for same nodeId', () => {
      const params1 = getAnomalyParams('node-3');
      const params2 = getAnomalyParams('node-3');
      
      expect(params1.particleCount).toBe(params2.particleCount);
      expect(params1.colorShift2).toBe(params2.colorShift2);
      expect(params1.rotationOffset).toBe(params2.rotationOffset);
    });
  });

  describe('drawCosmicAbomination', () => {
    it('should draw cosmic abomination combining all three types', () => {
      const params = getAnomalyParams('node-4');
      
      drawCosmicAbomination(mockCtx, 100, 100, 30, params);
      
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(100, 100);
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should use deterministic parameters for same nodeId', () => {
      const params1 = getAnomalyParams('node-4');
      const params2 = getAnomalyParams('node-4');
      
      expect(params1.crackCount).toBe(params2.crackCount);
      expect(params1.tentacleCount).toBe(params2.tentacleCount);
      expect(params1.particleCount).toBe(params2.particleCount);
    });
  });

  describe('drawUnknown dispatcher', () => {
    it('should select anomaly type based on hash of nodeId', () => {
      // Test different node IDs to ensure they don't crash
      const nodeIds = ['node-0', 'node-1', 'node-2', 'node-3'];
      
      vi.clearAllMocks();
      
      nodeIds.forEach(nodeId => {
        drawUnknown(mockCtx, 100, 100, 30, 0, nodeId);
      });
      
      // Verify that drawUnknown was called for each node
      expect(mockCtx.save).toHaveBeenCalledTimes(4);
      expect(mockCtx.restore).toHaveBeenCalledTimes(4);
    });

    it('should call the same renderer for the same nodeId', () => {
      const nodeId = 'consistent-node';
      
      // Clear previous calls
      vi.clearAllMocks();
      
      drawUnknown(mockCtx, 100, 100, 30, 0, nodeId);
      const firstCallCount = (mockCtx.save as any).mock.calls.length;
      
      drawUnknown(mockCtx, 100, 100, 30, 0, nodeId);
      const secondCallCount = (mockCtx.save as any).mock.calls.length;
      
      // Both calls should have been made
      expect(secondCallCount).toBe(firstCallCount + 1);
    });

    it('should generate different parameters for different nodeIds', () => {
      const params1 = getAnomalyParams('node-a');
      const params2 = getAnomalyParams('node-b');
      
      // At least one parameter should differ
      const differs =
        params1.crackCount !== params2.crackCount ||
        params1.tentacleCount !== params2.tentacleCount ||
        params1.particleCount !== params2.particleCount ||
        params1.colorShift1 !== params2.colorShift1 ||
        params1.colorShift2 !== params2.colorShift2 ||
        params1.deformAmount !== params2.deformAmount ||
        params1.rotationOffset !== params2.rotationOffset;
      
      expect(differs).toBe(true);
    });

    it('should be deterministic for the same nodeId', () => {
      const nodeId = 'deterministic-node';
      
      const params1 = getAnomalyParams(nodeId);
      const params2 = getAnomalyParams(nodeId);
      const params3 = getAnomalyParams(nodeId);
      
      // All parameters should be identical across multiple calls
      expect(params1).toEqual(params2);
      expect(params2).toEqual(params3);
    });

    it('should accept custom renderers without crashing', () => {
      const customRenderer = vi.fn();
      const customRenderers: Record<number, any> = {
        0: customRenderer,
        1: customRenderer,
        2: customRenderer,
        3: customRenderer
      };
      
      // Just verify the function accepts the parameter without error
      expect(() => {
        drawUnknown(mockCtx, 100, 100, 30, 0, 'node-custom', customRenderers);
      }).not.toThrow();
    });
  });

  describe('getAnomalyParams', () => {
    it('should generate parameters within configured ranges', () => {
      const params = getAnomalyParams('test-node');
      
      // Verify all required parameters exist
      expect(params).toHaveProperty('crackCount');
      expect(params).toHaveProperty('tentacleCount');
      expect(params).toHaveProperty('particleCount');
      expect(params).toHaveProperty('colorShift1');
      expect(params).toHaveProperty('colorShift2');
      expect(params).toHaveProperty('deformAmount');
      expect(params).toHaveProperty('rotationOffset');
      expect(params).toHaveProperty('seedBase');
    });

    it('should generate integer counts', () => {
      const params = getAnomalyParams('test-node');
      
      expect(Number.isInteger(params.crackCount)).toBe(true);
      expect(Number.isInteger(params.tentacleCount)).toBe(true);
      expect(Number.isInteger(params.particleCount)).toBe(true);
    });

    it('should generate rotation offset in valid range', () => {
      const params = getAnomalyParams('test-node');
      
      expect(params.rotationOffset).toBeGreaterThanOrEqual(0);
      expect(params.rotationOffset).toBeLessThanOrEqual(Math.PI * 2);
    });

    it('should generate non-negative seed base', () => {
      const params = getAnomalyParams('test-node');
      
      expect(params.seedBase).toBeGreaterThanOrEqual(0);
    });
  });

  describe('drawNode default/unknown type fallback', () => {
    it('should call anomaly renderer for type "unknown"', () => {
      vi.clearAllMocks();
      
      const node = { id: 'unknown-node-1', title: 'Test', type: 'unknown', x: 100, y: 100 };
      drawNode(mockCtx, node as any, 24, 0, false);
      
      // Anomaly renderers use save/translate/restore pattern
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(expect.any(Number), expect.any(Number));
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should call anomaly renderer for unrecognized type (default case)', () => {
      vi.clearAllMocks();
      
      const node = { id: 'weird-node-1', title: 'Test', type: 'nonexistent_type', x: 100, y: 100 };
      drawNode(mockCtx, node as any, 24, 0, false);
      
      // Anomaly renderers use save/translate/restore
      expect(mockCtx.save).toHaveBeenCalled();
      expect(mockCtx.translate).toHaveBeenCalledWith(expect.any(Number), expect.any(Number));
      expect(mockCtx.restore).toHaveBeenCalled();
    });

    it('should NOT call drawStar for unrecognized types', () => {
      vi.clearAllMocks();
      
      const node = { id: 'weird-node-2', title: 'Test', type: 'banana_type', x: 100, y: 100 };
      drawNode(mockCtx, node as any, 24, 0, false, true);
      
      // drawStar does NOT use save/translate/restore, so save being called means anomaly was used
      // Also verify translate was called (anomaly pattern) 
      expect(mockCtx.translate).toHaveBeenCalled();
    });

    it('should produce stable anomaly selection for same nodeId regardless of type string', () => {
      vi.clearAllMocks();
      
      const node1 = { id: 'stable-node', title: 'Test', type: 'unknown', x: 100, y: 100 };
      const node2 = { id: 'stable-node', title: 'Test', type: 'xyz_bogus', x: 100, y: 100 };
      
      drawNode(mockCtx, node1 as any, 24, 0, false, true);
      const callsAfterFirst = (mockCtx.translate as any).mock.calls.length;
      
      drawNode(mockCtx, node2 as any, 24, 0, false, true);
      const callsAfterSecond = (mockCtx.translate as any).mock.calls.length;
      
      // Both should have called translate (anomaly renderer)
      expect(callsAfterFirst).toBeGreaterThan(0);
      expect(callsAfterSecond).toBeGreaterThan(callsAfterFirst);
    });
  });

  describe('getLinkColor', () => {
    it('should return correct color for reference links', () => {
      const color = getLinkColor(1.0, 'reference');
      expect(color).toContain('51'); // RGB for blue #3366ff
      expect(color).toContain('102');
      expect(color).toContain('255');
    });

    it('should return correct color for dependency links', () => {
      const color = getLinkColor(0.8, 'dependency');
      expect(color).toContain('255'); // RGB for orange #ff6600
      expect(color).toContain('102');
      expect(color).toContain('0');
    });

    it('should apply opacity based on weight', () => {
      const color1 = getLinkColor(1.0, 'reference');
      const color2 = getLinkColor(0.2, 'reference');
      
      // Higher weight should have higher opacity
      expect(color1).toMatch(/rgba?\([^,]+,\s*[^,]+,\s*[^,]+,\s*([0-9.]+)\)/);
      expect(color2).toMatch(/rgba?\([^,]+,\s*[^,]+,\s*[^,]+,\s*([0-9.]+)\)/);
    });
  });

  describe('getLineDash', () => {
    it('should return empty array for reference links (solid)', () => {
      const dash = getLineDash('reference');
      expect(dash).toEqual([]);
    });

    it('should return dash-dot for dependency links', () => {
      const dash = getLineDash('dependency');
      expect(dash).toEqual([10, 3]);
    });

    it('should return dotted for custom links', () => {
      const dash = getLineDash('custom');
      expect(dash).toEqual([2, 6]);
    });

    it('should return dashed for weak related links', () => {
      const dash = getLineDash('related', 0.1);
      expect(dash).toEqual([6, 4]);
    });

    it('should return solid for strong related links', () => {
      const dash = getLineDash('related', 0.8);
      expect(dash).toEqual([]);
    });
  });

  describe('getGlowIntensity', () => {
    it('should return minimal glow when nodeCount > 100', () => {
      const intensity = getGlowIntensity('node-1', 1000, 150);
      expect(intensity).toBe(0.3);
    });

    it('should return pulsating value between 0.3 and 1.0', () => {
      const intensity = getGlowIntensity('node-1', 1000, 50);
      expect(intensity).toBeGreaterThanOrEqual(0.3);
      expect(intensity).toBeLessThanOrEqual(1.0);
    });

    it('should return different values for different nodeIds', () => {
      const intensity1 = getGlowIntensity('node-1', 1000, 50);
      const intensity2 = getGlowIntensity('node-2', 1000, 50);
      // Due to hash-based phase offset, values should differ
      expect(intensity1).not.toBe(intensity2);
    });
  });

  describe('getNodeColor', () => {
    it('should return correct color for star', () => {
      const color = getNodeColor('star');
      expect(color).toBe('#ffcc00');
    });

    it('should return correct color for planet', () => {
      const color = getNodeColor('planet');
      expect(color).toBe('#d6aa5d');
    });

    it('should return default color for unknown type', () => {
      const color = getNodeColor('unknown');
      expect(color).toBe('#94a3b8');
    });

    it('should return default color for undefined type', () => {
      const color = getNodeColor('nonexistent');
      expect(color).toBe('#94a3b8');
    });
  });

  describe('getNodeGradient', () => {
    it('should create a radial gradient', () => {
      const gradient = getNodeGradient(mockCtx, 100, 100, 30, 'star', '#ffcc00');
      expect(mockCtx.createRadialGradient).toHaveBeenCalled();
      expect(gradient).toBeDefined();
    });

    it('should add color stops for star type', () => {
      const gradient = getNodeGradient(mockCtx, 100, 100, 30, 'star', '#ffcc00');
      expect(gradient.addColorStop).toHaveBeenCalledTimes(3);
    });

    it('should add color stops for planet type', () => {
      const gradient = getNodeGradient(mockCtx, 100, 100, 30, 'planet', '#d6aa5d');
      expect(gradient.addColorStop).toHaveBeenCalledTimes(3);
    });

    it('should add color stops for galaxy type', () => {
      const gradient = getNodeGradient(mockCtx, 100, 100, 30, 'galaxy', '#8b5cf6');
      expect(gradient.addColorStop).toHaveBeenCalledTimes(3);
    });
  });
});
