import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../../../vitest-setup';
import { getGraphData, getFullGraphData } from './graph';
import type { GraphNode, GraphLink, GraphData } from './graph';

describe('graph API', () => {
  describe('getGraphData', () => {
    it('should return graph data for note', async () => {
      const mockGraphData: GraphData = {
        nodes: [
          { id: '1', title: 'Center Node', type: 'star', x: 0, y: 0, z: 0, size: 10 },
          { id: '2', title: 'Related Node 1', type: 'planet', x: 10, y: 10, z: 0, size: 5 },
          { id: '3', title: 'Related Node 2', type: 'moon', x: -10, y: 10, z: 0, size: 3 },
        ],
        links: [
          { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
          { source: '1', target: '3', weight: 0.6, link_type: 'related' },
        ],
      };

      server.use(
        http.get('http://localhost:8081/api/notes/1/graph', ({ request }) => {
          const url = new URL(request.url);
          if (url.searchParams.get('depth') === '2') {
            return HttpResponse.json(mockGraphData);
          }
          return HttpResponse.json(mockGraphData);
        })
      );

      const result = await getGraphData('1', 2);

      expect(result.nodes).toHaveLength(3);
      expect(result.links).toHaveLength(2);
      expect(result.nodes[0].title).toBe('Center Node');
    });
  });

  describe('getFullGraphData', () => {
    it('should return full graph data', async () => {
      const mockGraphData: GraphData = {
        nodes: [
          { id: '1', title: 'Node 1', type: 'star' },
          { id: '2', title: 'Node 2', type: 'planet' },
        ],
        links: [{ source: '1', target: '2', weight: 1.0, link_type: 'reference' }],
      };

      server.use(
        http.get('http://localhost:8081/api/graph/all', ({ request }) => {
          const url = new URL(request.url);
          const limit = url.searchParams.get('limit');
          return HttpResponse.json(mockGraphData);
        })
      );

      const result = await getFullGraphData(50);

      expect(result.nodes).toHaveLength(2);
    });

    it('should handle large graphs', async () => {
      const manyNodes: GraphNode[] = Array.from({ length: 100 }, (_, i) => ({
        id: String(i),
        title: `Node ${i}`,
        type: i % 3 === 0 ? 'star' : 'planet',
      }));

      const manyLinks: GraphLink[] = Array.from({ length: 99 }, (_, i) => ({
        source: String(i),
        target: String(i + 1),
        weight: 0.5,
        link_type: 'reference',
      }));

      server.use(
        http.get('http://localhost:8081/api/graph/all', () => HttpResponse.json({ nodes: manyNodes, links: manyLinks }))
      );

      const result = await getFullGraphData(100);

      expect(result.nodes).toHaveLength(100);
      expect(result.links).toHaveLength(99);
    });
  });

  describe('edge cases', () => {
    it('should handle empty graph', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/999/graph', () => HttpResponse.json({ nodes: [], links: [] }))
      );

      const result = await getGraphData('999', 1);

      expect(result.nodes).toHaveLength(0);
      expect(result.links).toHaveLength(0);
    });
  });

  describe('error handling', () => {
    it('should handle network errors', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/1/graph', () => HttpResponse.error())
      );

      await expect(getGraphData('1')).rejects.toThrow();
    });

    it('should handle HTTP 500 errors', async () => {
      server.use(
        http.get('http://localhost:8081/api/graph/all', () => HttpResponse.json({ error: 'Server error' }, { status: 500 }))
      );

      await expect(getFullGraphData(1000)).rejects.toThrow();
    });
  });
});
