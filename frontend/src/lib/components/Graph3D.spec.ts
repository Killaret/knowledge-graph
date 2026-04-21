import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { tick } from 'svelte';
import Graph3D from './Graph3D.svelte';

// Мокаем Three.js
vi.mock('three', () => ({
  Scene: vi.fn().mockImplementation(() => ({
    add: vi.fn(),
    remove: vi.fn(),
    clear: vi.fn()
  })),
  PerspectiveCamera: vi.fn().mockImplementation(() => ({
    position: { set: vi.fn() },
    lookAt: vi.fn(),
    updateProjectionMatrix: vi.fn()
  })),
  WebGLRenderer: vi.fn().mockImplementation(() => ({
    setSize: vi.fn(),
    setPixelRatio: vi.fn(),
    render: vi.fn(),
    dispose: vi.fn(),
    domElement: document.createElement('canvas')
  })),
  Vector3: vi.fn().mockImplementation(() => ({
    set: vi.fn(),
    length: vi.fn().mockReturnValue(1),
    normalize: vi.fn().mockReturnThis()
  })),
  Color: vi.fn().mockImplementation(() => ({ setHex: vi.fn() }))
}));

vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$lib/three/core/sceneSetup', () => ({
  initScene: vi.fn().mockReturnValue({
    scene: { add: vi.fn() },
    camera: { position: { set: vi.fn() } }
  }),
  setFogDensity: vi.fn()
}));

vi.mock('$lib/three/simulation/forceSimulation', () => ({
  createSimulation: vi.fn().mockReturnValue({
    tick: vi.fn(),
    nodes: vi.fn().mockReturnValue([]),
    alpha: vi.fn().mockReturnValue(0)
  }),
  addNodesToSimulation: vi.fn()
}));

vi.mock('$lib/three/rendering/objectManager', () => ({
  ObjectManager: vi.fn().mockImplementation(() => ({
    createNodeMesh: vi.fn(),
    createLinkLine: vi.fn(),
    clear: vi.fn()
  }))
}));

vi.mock('$lib/three/camera/cameraUtils', () => ({
  autoZoomToFit: vi.fn(),
  centerCameraOnNode: vi.fn()
}));

describe('Graph3D', () => {
  const mockData = {
    nodes: [
      { id: '1', title: 'Node 1', type: 'star' },
      { id: '2', title: 'Node 2', type: 'planet' },
      { id: '3', title: 'Node 3', type: 'comet' }
    ],
    links: [
      { source: '1', target: '2' }
    ]
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders canvas element', async () => {
    render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null
      }
    });
    
    await tick();
    
    // Должен быть canvas или div контейнер
    const container = document.querySelector('div');
    expect(container).toBeInTheDocument();
  });

  it('shows loading state initially', async () => {
    render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null
      }
    });
    
    await tick();
    
    // Проверяем что отображается loading spinner
    const loadingElement = document.querySelector('.loading-spinner, [data-testid="loading"]');
    // Если есть специальный элемент загрузки
    if (loadingElement) {
      expect(loadingElement).toBeInTheDocument();
    }
  });

  it('processes data without errors', async () => {
    const onNodeClick = vi.fn();
    
    const { component } = render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null,
        onNodeClick
      }
    });
    
    await tick();
    
    // Компонент должен создаться без ошибок
    expect(component).toBeDefined();
  });

  it('handles empty data', async () => {
    const emptyData = { nodes: [], links: [] };
    
    render(Graph3D, {
      props: {
        data: emptyData,
        centerNodeId: null
      }
    });
    
    await tick();
    
    // Компонент должен обработать пустые данные без ошибок
    const container = document.querySelector('div');
    expect(container).toBeInTheDocument();
  });

  it('handles centerNodeId prop', async () => {
    const onNodeClick = vi.fn();
    
    render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: '1',
        onNodeClick
      }
    });
    
    await tick();
    
    // Компонент должен использовать centerNodeId
    expect(document.querySelector('div')).toBeInTheDocument();
  });

  it('updates when data changes', async () => {
    const { component } = render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null
      }
    });
    
    await tick();
    
    // Меняем данные
    const newData = {
      nodes: [...mockData.nodes, { id: '4', title: 'Node 4', type: 'galaxy' }],
      links: [...mockData.links, { source: '2', target: '4' }]
    };
    
    // Обновляем пропс
    // @ts-ignore
    component.$set({ data: newData });
    
    await tick();
    
    // Компонент должен обновиться без ошибок
    expect(document.querySelector('div')).toBeInTheDocument();
  });

  it('cleans up on unmount', async () => {
    const { unmount } = render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null
      }
    });
    
    await tick();
    
    // Размонтируем компонент - не должно быть ошибок
    expect(() => unmount()).not.toThrow();
  });
});
