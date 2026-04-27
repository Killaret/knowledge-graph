import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { tick } from 'svelte';
import SmartGraph from './SmartGraph.svelte';

// Mock device capabilities
vi.mock('$lib/utils/deviceCapabilities', () => ({
  detectDeviceCapabilities: vi.fn().mockReturnValue({
    isMobile: false,
    isLowPower: false,
    hasWebGL: true,
    screenSize: { width: 1920, height: 1080 }
  }),
  shouldUse3D: vi.fn().mockReturnValue(false)
}));

// Mock GraphCanvas
vi.mock('./GraphCanvas.svelte', () => ({
  default: vi.fn().mockImplementation(() => {
    const div = document.createElement('div');
    div.setAttribute('data-testid', 'graph-canvas');
    return div;
  })
}));

describe('SmartGraph', () => {
  const mockData = {
    nodes: [
      { id: '1', title: 'Node 1', type: 'star' },
      { id: '2', title: 'Node 2', type: 'planet' }
    ],
    links: [
      { source: '1', target: '2', weight: 1 }
    ]
  };

  it('renders without errors', async () => {
    const { container } = render(SmartGraph, {
      props: mockData
    });

    await tick();

    expect(container).toBeTruthy();
  });

  it('accepts nodes and links props', async () => {
    const { container } = render(SmartGraph, {
      props: {
        nodes: mockData.nodes,
        links: mockData.links
      }
    });

    await tick();

    expect(container).toBeTruthy();
  });

  it('handles empty data', async () => {
    const { container } = render(SmartGraph, {
      props: {
        nodes: [],
        links: []
      }
    });

    await tick();

    expect(container).toBeTruthy();
  });
});
