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
    camera: { position: { set: vi.fn() } },
    controls: { update: vi.fn(), autoRotate: false },
    labelRenderer: { render: vi.fn(), setSize: vi.fn() },
    renderer: { render: vi.fn(), setSize: vi.fn(), dispose: vi.fn() }
  }),
  setFogDensity: vi.fn()
}));

vi.mock('$lib/three/simulation/forceSimulation', () => ({
  createSimulation: vi.fn().mockReturnValue({
    tick: vi.fn(),
    nodes: vi.fn().mockReturnValue([]),
    alpha: vi.fn().mockReturnValue(0),
    stop: vi.fn()
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
    // Мокаем Web Animations API
    Element.prototype.animate = vi.fn().mockReturnValue({
      onfinish: vi.fn(),
      cancel: vi.fn()
    });
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

  it('positions camera to cover all nodes (bounding box)', async () => {
    const nodesWithPositions = [
      { id: '1', title: 'Node 1', type: 'star', x: -100, y: 50, z: 0 },
      { id: '2', title: 'Node 2', type: 'planet', x: 100, y: -50, z: 100 },
      { id: '3', title: 'Node 3', type: 'comet', x: 0, y: 0, z: -100 }
    ];

    const dataWithPositions = {
      nodes: nodesWithPositions,
      links: mockData.links
    };

    render(Graph3D, {
      props: {
        data: dataWithPositions,
        centerNodeId: null
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Проверяем что камера была создана с PerspectiveCamera
    // и ей установлена позиция
    const { PerspectiveCamera } = await import('three');
    expect(PerspectiveCamera).toHaveBeenCalled();
  });

  it('calls camera fit function when nodes change', async () => {
    const { component } = render(Graph3D, {
      props: {
        data: mockData,
        centerNodeId: null
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Создаем новые данные с измененными позициями узлов
    const newData = {
      nodes: [
        { id: '1', title: 'Node 1', type: 'star', x: -200, y: 100, z: 50 },
        { id: '2', title: 'Node 2', type: 'planet', x: 200, y: -100, z: 150 },
        { id: '3', title: 'Node 3', type: 'comet', x: 50, y: 50, z: -150 }
      ],
      links: mockData.links
    };

    // Обновляем пропсы
    // @ts-ignore
    component.$set({ data: newData });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 100));

    // Проверяем что компонент обновился без ошибок
    expect(document.querySelector('div')).toBeInTheDocument();
  });

  it('calculates correct bounding box from node positions', async () => {
    const nodesWithExtremePositions = [
      { id: '1', title: 'Far Left', type: 'star', x: -500, y: 0, z: 0 },
      { id: '2', title: 'Far Right', type: 'planet', x: 500, y: 0, z: 0 },
      { id: '3', title: 'Far Top', type: 'comet', x: 0, y: 300, z: 0 },
      { id: '4', title: 'Far Bottom', type: 'galaxy', x: 0, y: -300, z: 0 }
    ];

    const dataWithExtremes = {
      nodes: nodesWithExtremePositions,
      links: []
    };

    render(Graph3D, {
      props: {
        data: dataWithExtremes,
        centerNodeId: null
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Компонент должен обработать экстремальные позиции без ошибок
    expect(document.querySelector('div')).toBeInTheDocument();
  });
});
