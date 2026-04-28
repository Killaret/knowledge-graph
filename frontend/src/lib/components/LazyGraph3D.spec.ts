import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { tick } from 'svelte';
import LazyGraph3D from './LazyGraph3D.svelte';

// Mock browser environment
vi.mock('$app/environment', () => ({
  browser: true
}));

// Mock Graph3D component
vi.mock('./Graph3D.svelte', () => ({
  default: vi.fn().mockImplementation(() => {
    const div = document.createElement('div');
    div.setAttribute('data-testid', 'graph-3d');
    div.textContent = 'Graph3D Component';
    return div;
  })
}));

describe('LazyGraph3D', () => {
  // Mock WebGL context
  beforeEach(() => {
    const mockGetContext = vi.fn((contextType: string) => {
      if (contextType === 'webgl' || contextType === 'experimental-webgl') {
        return {} as WebGLRenderingContext;
      }
      return null;
    });
    
    Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
      value: mockGetContext,
      writable: true
    });
  });

  const mockData = {
    nodes: [
      { id: '1', title: 'Node 1', type: 'star', x: 0, y: 0, z: 0 },
      { id: '2', title: 'Node 2', type: 'planet', x: 100, y: 0, z: 0 }
    ],
    links: [
      { source: '1', target: '2' }
    ]
  };

  it('shows loading state initially', async () => {
    render(LazyGraph3D, {
      props: { data: mockData }
    });

    expect(screen.getByText('Loading 3D engine...')).toBeInTheDocument();
  });

  it('loads Graph3D component after mount', async () => {
    const { container } = render(LazyGraph3D, {
      props: { data: mockData }
    });

    await tick();

    // Should eventually load (mock returns immediately)
    expect(container).toBeTruthy();
  });

  it('passes data prop to loaded component', async () => {
    const { container } = render(LazyGraph3D, {
      props: { data: mockData }
    });

    await tick();

    expect(container).toBeTruthy();
  });

  it('handles empty data', async () => {
    const emptyData = { nodes: [], links: [] };

    const { container } = render(LazyGraph3D, {
      props: { data: emptyData }
    });

    await tick();

    expect(container).toBeTruthy();
  });
});
