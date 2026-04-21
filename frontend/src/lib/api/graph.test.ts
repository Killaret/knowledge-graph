import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  getGraphData,
  getFullGraphData,
  type GraphNode,
  type GraphLink,
  type GraphData
} from './graph';

// Мокаем ky модуль
vi.mock('ky', () => {
  const mockResponse = (data: any) => Promise.resolve({
    json: () => Promise.resolve(data)
  });

  const mockKy = {
    get: vi.fn(() => mockResponse({ nodes: [], links: [] })),
    post: vi.fn(() => mockResponse({})),
    put: vi.fn(() => mockResponse({})),
    delete: vi.fn(() => Promise.resolve()),
    extend: vi.fn(() => mockKy)
  };

  return {
    default: mockKy,
    create: vi.fn(() => mockKy)
  };
});

describe('graph API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getGraphData should return graph data for note', async () => {
    const mockGraphData: GraphData = {
      nodes: [
        { id: '1', title: 'Center Node', type: 'star', x: 0, y: 0, z: 0, size: 10 },
        { id: '2', title: 'Related Node 1', type: 'planet', x: 10, y: 10, z: 0, size: 5 },
        { id: '3', title: 'Related Node 2', type: 'moon', x: -10, y: 10, z: 0, size: 3 }
      ],
      links: [
        { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
        { source: '1', target: '3', weight: 0.6, link_type: 'related' }
      ]
    };

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve(mockGraphData)
    });

    const result = await getGraphData('1', 2);

    expect(result.nodes).toHaveLength(3);
    expect(result.links).toHaveLength(2);
    expect(result.nodes[0].title).toBe('Center Node');
  });

  it('getFullGraphData should return full graph data', async () => {
    const mockGraphData: GraphData = {
      nodes: [
        { id: '1', title: 'Node 1', type: 'star' },
        { id: '2', title: 'Node 2', type: 'planet' }
      ],
      links: [
        { source: '1', target: '2', weight: 1.0, link_type: 'reference' }
      ]
    };

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve(mockGraphData)
    });

    const result = await getFullGraphData(50);

    expect(result.nodes).toHaveLength(2);
  });

  it('getGraphData should handle empty graph', async () => {
    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve({ nodes: [], links: [] })
    });

    const result = await getGraphData('999');

    expect(result.nodes).toHaveLength(0);
    expect(result.links).toHaveLength(0);
  });

  it('should handle network errors', async () => {
    const ky = await import('ky');
    ky.default.get.mockRejectedValueOnce(new Error('Network error'));

    await expect(getGraphData('1')).rejects.toThrow('Network error');
  });

  it('getFullGraphData should handle large graphs', async () => {
    const manyNodes: GraphNode[] = Array.from({ length: 100 }, (_, i) => ({
      id: String(i),
      title: `Node ${i}`,
      type: i % 3 === 0 ? 'star' : 'planet'
    }));

    const manyLinks: GraphLink[] = Array.from({ length: 99 }, (_, i) => ({
      source: String(i),
      target: String(i + 1),
      weight: 0.5,
      link_type: 'reference'
    }));

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve({ nodes: manyNodes, links: manyLinks })
    });

    const result = await getFullGraphData(100);

    expect(result.nodes).toHaveLength(100);
    expect(result.links).toHaveLength(99);
  });
});
