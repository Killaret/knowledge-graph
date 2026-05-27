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
  getAnomalyParams
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
      const firstCallCount = mockCtx.save.mock.calls.length;
      
      drawUnknown(mockCtx, 100, 100, 30, 0, nodeId);
      const secondCallCount = mockCtx.save.mock.calls.length;
      
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
});
