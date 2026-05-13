// Unit тесты для интеграции тумана в Graph3D.svelte
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as THREE from 'three';
import type { GraphData } from '$lib/api/graph';

// Мокаем зависимости Three.js
vi.mock('three', () => ({
  Scene: vi.fn().mockImplementation(() => ({
    fog: { density: 0.08, near: 0.3, far: 10 },
    add: vi.fn(),
    remove: vi.fn(),
    children: []
  })),
  PerspectiveCamera: vi.fn().mockImplementation(() => ({
    position: { set: vi.fn() },
    lookAt: vi.fn(),
    aspect: 1,
    updateProjectionMatrix: vi.fn()
  })),
  WebGLRenderer: vi.fn().mockImplementation(() => ({
    setSize: vi.fn(),
    setPixelRatio: vi.fn(),
    render: vi.fn(),
    domElement: document.createElement('canvas')
  })),
  FogExp2: vi.fn().mockImplementation((color, density) => ({ color, density })),
  Color: vi.fn().mockImplementation(() => ({})),
  AmbientLight: vi.fn().mockImplementation(() => ({})),
  DirectionalLight: vi.fn().mockImplementation(() => ({
    position: { set: vi.fn() },
    castShadow: true,
    shadow: { mapSize: { width: 1024, height: 1024 } }
  })),
  PointLight: vi.fn().mockImplementation(() => ({
    position: { set: vi.fn() }
  })),
  BufferGeometry: vi.fn().mockImplementation(() => ({
    setAttribute: vi.fn()
  })),
  BufferAttribute: vi.fn().mockImplementation(() => ({})),
  PointsMaterial: vi.fn().mockImplementation(() => ({})),
  Points: vi.fn().mockImplementation(() => ({})),
  Raycaster: vi.fn().mockImplementation(() => ({
    setFromCamera: vi.fn(),
    intersectObjects: vi.fn(() => [])
  })),
  Vector2: vi.fn().mockImplementation(() => ({ x: 0, y: 0 })),
  AdditiveBlending: 0
}));

// Мокаем OrbitControls
vi.mock('three/examples/jsm/controls/OrbitControls.js', () => ({
  OrbitControls: vi.fn().mockImplementation(() => ({
    enableDamping: true,
    dampingFactor: 0.05,
    autoRotate: true,
    autoRotateSpeed: 0.8,
    enableZoom: true,
    enablePan: true,
    maxPolarAngle: Math.PI / 1.8,
    update: vi.fn()
  }))
}));

// Мокаем CSS2DRenderer
vi.mock('three/examples/jsm/renderers/CSS2DRenderer.js', () => ({
  CSS2DRenderer: vi.fn().mockImplementation(() => ({
    setSize: vi.fn(),
    domElement: document.createElement('div')
  }))
}));

// Мокаем PreloadService
vi.mock('$lib/services/PreloadService', () => ({
  hasPreloadedData: vi.fn(() => false)
}));

// Мокаем Graph3D модули
const mockSetFogDensity = vi.fn();
const mockCreateSimulation = vi.fn().mockImplementation(() => ({
  on: vi.fn(),
  stop: vi.fn(),
  nodes: vi.fn(() => [
    { id: '1', x: 0, y: 0, z: 0 },
    { id: '2', x: 10, y: 0, z: 0 },
    { id: '3', x: 5, y: 10, z: 0 }
  ]),
  force: vi.fn(() => ({ links: vi.fn(() => []) }))
}));

const mockAnimateFogDensity = vi.fn().mockReturnValue({ stop: vi.fn() });

vi.mock('./Graph3D/sceneSetup', () => ({
  initScene: vi.fn().mockReturnValue({
    scene: new THREE.Scene(),
    camera: new THREE.PerspectiveCamera(),
    renderer: new THREE.WebGLRenderer(),
    labelRenderer: { render: vi.fn(), domElement: document.createElement('div') },
    controls: { update: vi.fn() }
  }),
  setFogDensity: mockSetFogDensity,
  resizeScene: vi.fn(),
  disposeScene: vi.fn()
}));

vi.mock('./Graph3D/simulation', () => ({
  createSimulation: mockCreateSimulation,
  autoZoomToFit: vi.fn(),
  centerCameraOnNode: vi.fn(),
  toggleAutoRotate: vi.fn(() => true)
}));

// Мокаем fogManager
vi.mock('./Graph3D/fogManager', () => ({
  animateFogDensity: mockAnimateFogDensity
}));

// Мокаем ObjectManager
vi.mock('$lib/three/rendering/objectManager', () => ({
  ObjectManager: vi.fn().mockImplementation(() => ({
    clear: vi.fn(),
    updatePositions: vi.fn(),
    updateLinks: vi.fn()
  }))
}));

// Мокаем browser environment
vi.mock('$app/environment', () => ({
  browser: true
}));

// Мокаем graphUtils
vi.mock('$lib/utils/graphUtils', () => ({
  filterValidLinks: vi.fn((nodes, links) => links)
}));

// Мокаем config
vi.mock('$lib/config', () => ({
  graphConfig3D: { max_nodes: 1000 }
}));

describe('Graph3D Fog Integration', () => {
  let mockData: GraphData;
  let mockSetFogDensity: ReturnType<typeof vi.fn>;
  let mockAnimateFogDensity: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.clearAllMocks();
    
    mockData = {
      nodes: [
        { id: '1', title: 'Node 1', type: 'star' },
        { id: '2', title: 'Node 2', type: 'planet' },
        { id: '3', title: 'Node 3', type: 'star' }
      ],
      links: [
        { source: '1', target: '2', link_type: 'reference' },
        { source: '2', target: '3', link_type: 'reference' }
      ]
    };

    // Получаем моки из импортированных модулей
    const Graph3DMocks = vi.mocked(import('./Graph3D'));
    mockSetFogDensity = Graph3DMocks.setFogDensity as ReturnType<typeof vi.fn>;
    
    const fogManagerMocks = vi.mocked(import('./Graph3D/fogManager'));
    mockAnimateFogDensity = fogManagerMocks.animateFogDensity as ReturnType<typeof vi.fn>;
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.restoreAllMocks();
  });

  describe('Fog Density Management', () => {
    it('should set initial fog density when simulation starts', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Проверяем, что setFogDensity был вызван с начальной плотностью
      expect(mockSetFogDensity).toHaveBeenCalledWith(
        expect.any(Object), // scene
        0.08 // initial density
      );
    });

    it('should use lower initial density when data is preloaded', async () => {
      // Мокаем предзагруженные данные
      const { hasPreloadedData } = vi.mocked(import('$lib/services/PreloadService'));
      hasPreloadedData.mockReturnValue(true);

      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Проверяем, что используется меньшая начальная плотность
      expect(mockSetFogDensity).toHaveBeenCalledWith(
        expect.any(Object),
        0.04 // lower density for preloaded data
      );
    });

    it('should progressively decrease fog density during simulation', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Симулируем несколько тиков симуляции
      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      const simulation = Graph3DMocks.createSimulation();
      
      // Получаем callback для 'tick' события
      const tickCallback = (simulation.on as ReturnType<typeof vi.fn>).mock.calls
        .find(call => call[0] === 'tick')?.[1];

      if (tickCallback) {
        // Вызываем tick несколько раз для проверки прогрессивного изменения
        for (let i = 0; i < 20; i++) {
          tickCallback();
        }

        // Проверяем, что setFogDensity вызывался несколько раз с разными значениями
        expect(mockSetFogDensity).toHaveBeenCalledTimes(expect.any(Number));
        
        // Последний вызов должен быть с плотностью меньше начальной
        const lastCall = mockSetFogDensity.mock.calls[mockSetFogDensity.mock.calls.length - 1];
        const lastDensity = lastCall[1];
        expect(lastDensity).toBeLessThan(0.08);
        expect(lastDensity).toBeGreaterThanOrEqual(0.005);
      }
    });

    it('should call animateFogDensity when simulation ends', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Получаем callback для 'end' события
      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      const simulation = Graph3DMocks.createSimulation();
      
      const endCallback = (simulation.on as ReturnType<typeof vi.fn>).mock.calls
        .find(call => call[0] === 'end')?.[1];

      if (endCallback) {
        endCallback();

        // Проверяем, что animateFogDensity был вызван
        expect(mockAnimateFogDensity).toHaveBeenCalledWith(
          expect.any(Object), // scene
          0.005, // target density
          800 // duration
        );
      }
    });

    it('should handle empty graph data correctly', async () => {
      const emptyData: GraphData = { nodes: [], links: [] };
      
      const { component } = render(Graph3D, { 
        props: { data: emptyData }
      });

      await tick();

      // Для пустого графа туман должен быть сразу установлен на минимальную плотность
      expect(mockSetFogDensity).toHaveBeenCalledWith(
        expect.any(Object),
        0.005 // minimal density for empty graph
      );
    });

    it('should handle single node graph correctly', async () => {
      const singleNodeData: GraphData = { 
        nodes: [{ id: '1', title: 'Single', type: 'star' }], 
        links: [] 
      };
      
      const { component } = render(Graph3D, { 
        props: { data: singleNodeData }
      });

      await tick();

      // Для графа с одним узлом туман должен быть сразу установлен на минимальную плотность
      expect(mockSetFogDensity).toHaveBeenCalledWith(
        expect.any(Object),
        0.005 // minimal density for single node
      );
    });
  });

  describe('Progressive Fog Logic', () => {
    it('should calculate fog density based on node positioning progress', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      const simulation = Graph3DMocks.createSimulation();
      
      // Мокаем узлы с разным прогрессом позиционирования
      simulation.nodes.mockReturnValue([
        { id: '1', x: 0, y: 0, z: 0 }, // позиционирован
        { id: '2', x: undefined, y: 0, z: 0 }, // не позиционирован
        { id: '3', x: 10, y: 10, z: 10 } // позиционирован
      ]);

      const tickCallback = (simulation.on as ReturnType<typeof vi.fn>).mock.calls
        .find(call => call[0] === 'tick')?.[1];

      if (tickCallback) {
        // Вызываем tick для обновления тумана
        tickCallback();

        // Проверяем, что плотность тумана рассчитывается правильно
        // 2 из 3 узлов позиционированы = 66.7% прогресс
        // Ожидаемая плотность: 0.08 - (0.08 - 0.005) * 0.667 ≈ 0.03
        const lastCall = mockSetFogDensity.mock.calls[mockSetFogDensity.mock.calls.length - 1];
        const calculatedDensity = lastCall[1];
        expect(calculatedDensity).toBeCloseTo(0.03, 1);
      }
    });

    it('should handle NaN coordinates gracefully', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      const simulation = Graph3DMocks.createSimulation();
      
      // Мокаем узлы с NaN координатами
      simulation.nodes.mockReturnValue([
        { id: '1', x: NaN, y: 0, z: 0 },
        { id: '2', x: 0, y: NaN, z: 0 },
        { id: '3', x: 0, y: 0, z: NaN }
      ]);

      const tickCallback = (simulation.on as ReturnType<typeof vi.fn>).mock.calls
        .find(call => call[0] === 'tick')?.[1];

      if (tickCallback) {
        // Не должно быть ошибок
        expect(() => tickCallback()).not.toThrow();
        
        // Плотность должна остаться высокой (низкий прогресс)
        const lastCall = mockSetFogDensity.mock.calls[mockSetFogDensity.mock.calls.length - 1];
        expect(lastCall[1]).toBeCloseTo(0.08, 2);
      }
    });

    it('should update fog density every 5 ticks', async () => {
      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      const simulation = Graph3DMocks.createSimulation();
      
      const tickCallback = (simulation.on as ReturnType<typeof vi.fn>).mock.calls
        .find(call => call[0] === 'tick')?.[1];

      if (tickCallback) {
        const initialCallCount = mockSetFogDensity.mock.calls.length;

        // Вызываем tick 4 раза - не должно быть обновлений тумана
        for (let i = 0; i < 4; i++) {
          tickCallback();
        }
        expect(mockSetFogDensity.mock.calls.length).toBe(initialCallCount);

        // Вызываем 5-й раз - должно быть обновление тумана
        tickCallback();
        expect(mockSetFogDensity.mock.calls.length).toBeGreaterThan(initialCallCount);
      }
    });
  });

  describe('Integration with PreloadService', () => {
    it('should log preloaded data status in development mode', async () => {
      // Мокаем development mode
      const originalImportMeta = import.meta;
      Object.defineProperty(import.meta, 'env', {
        value: { DEV: true },
        writable: true
      });

      // Мокаем console.log
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      const { hasPreloadedData } = vi.mocked(import('$lib/services/PreloadService'));
      hasPreloadedData.mockReturnValue(true);

      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Проверяем, что был вызван лог с информацией о предзагруженных данных
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('[Graph3D] Starting simulation with preloaded data: true'),
        expect.stringContaining('initial fog density: 0.04')
      );

      consoleSpy.mockRestore();
    });
  });

  describe('Error Handling', () => {
    it('should handle missing scene gracefully', async () => {
      // Мокаем initScene чтобы вернуть null scene
      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      Graph3DMocks.initScene.mockReturnValue({
        scene: null,
        camera: new THREE.PerspectiveCamera(),
        renderer: new THREE.WebGLRenderer(),
        labelRenderer: { render: vi.fn() },
        controls: { update: vi.fn() }
      });

      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Не должно быть ошибок и setFogDensity не должен вызываться
      expect(mockSetFogDensity).not.toHaveBeenCalled();
    });

    it('should handle simulation errors gracefully', async () => {
      // Мокаем createSimulation чтобы выбросить ошибку
      const Graph3DMocks = vi.mocked(import('./Graph3D'));
      Graph3DMocks.createSimulation.mockImplementation(() => {
        throw new Error('Simulation error');
      });

      const { component } = render(Graph3D, { 
        props: { data: mockData }
      });

      await tick();

      // Начальная плотность должна быть установлена
      expect(mockSetFogDensity).toHaveBeenCalledWith(
        expect.any(Object),
        0.08
      );
    });
  });
});
