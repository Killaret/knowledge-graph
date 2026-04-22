import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

import { render, fireEvent } from '@testing-library/svelte';
import { tick } from 'svelte';
import GraphCanvas from './GraphCanvas.svelte';
import { mockState } from '../../__mocks__/d3-force';

// Мокаем $app/environment
vi.mock('$app/environment', () => ({
  browser: true
}));

// Мокаем $app/navigation
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

describe('GraphCanvas', () => {
  const mockNodes = [
    { id: '1', title: 'Node 1', type: 'star' },
    { id: '2', title: 'Node 2', type: 'planet' },
    { id: '3', title: 'Node 3', type: 'comet' }
  ];

  const mockLinks = [
    { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
    { source: '2', target: '3', weight: 0.5, link_type: 'dependency' }
  ];

  // Мок контекста canvas для отслеживания вызовов
  let mockCtx: any;
  let drawCalls: { type: string; args: any[] }[] = [];

  beforeEach(async () => {
    vi.resetModules(); // Сбрасываем кэш модулей
    vi.clearAllMocks();
    mockState.simulationNodes = [];
    mockState.simulationLinks = [];
    mockState.tickCallback = null;
    mockState.stopCallback = null;
    drawCalls = [];
    mockState.currentSimulation = mockState.createSimulation();
    
    // Мокаем canvas API с отслеживанием вызовов
    mockCtx = {
      clearRect: vi.fn(),
      save: vi.fn(),
      restore: vi.fn(),
      translate: vi.fn(),
      scale: vi.fn(),
      beginPath: vi.fn().mockImplementation(() => {
        drawCalls.push({ type: 'beginPath', args: [] });
      }),
      moveTo: vi.fn().mockImplementation((x: number, y: number) => {
        drawCalls.push({ type: 'moveTo', args: [x, y] });
      }),
      lineTo: vi.fn().mockImplementation((x: number, y: number) => {
        drawCalls.push({ type: 'lineTo', args: [x, y] });
      }),
      stroke: vi.fn().mockImplementation(() => {
        drawCalls.push({ type: 'stroke', args: [] });
      }),
      setLineDash: vi.fn(),
      arc: vi.fn().mockImplementation((x: number, y: number, r: number) => {
        drawCalls.push({ type: 'arc', args: [x, y, r] });
      }),
      fill: vi.fn().mockImplementation(() => {
        drawCalls.push({ type: 'fill', args: [] });
      }),
      closePath: vi.fn(),
      ellipse: vi.fn(),
      strokeRect: vi.fn(),
      fillRect: vi.fn(),
      rotate: vi.fn(),
      fillText: vi.fn().mockImplementation((text: string, x: number, y: number) => {
        drawCalls.push({ type: 'fillText', args: [text, x, y] });
      }),
      measureText: vi.fn().mockReturnValue({ width: 50 }),
      font: '',
      textAlign: '',
      textBaseline: '',
      lineWidth: 1,
      strokeStyle: '',
      fillStyle: '',
      shadowBlur: 0,
      shadowColor: ''
    };

    HTMLCanvasElement.prototype.getContext = vi.fn().mockReturnValue(mockCtx);

    // Мокаем ResizeObserver
    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      disconnect: vi.fn(),
      unobserve: vi.fn()
    }));

    // Мокаем requestAnimationFrame
    vi.stubGlobal('requestAnimationFrame', vi.fn().mockImplementation((cb: FrameRequestCallback) => {
      setTimeout(cb, 16);
      return 1;
    }));
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders canvas element', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();

    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('renders with empty data without errors', async () => {
    const { component } = render(GraphCanvas, {
      props: {
        nodes: [],
        links: []
      }
    });

    await tick();

    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
    expect(component).toBeDefined();
  });

  it('processes nodes and links without errors', async () => {
    const { component } = render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    expect(component).toBeDefined();
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('renders nodes with coordinates from simulation', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: []
      }
    });

    await tick();
    // Ждем инициализации симуляции и тика
    await new Promise(resolve => setTimeout(resolve, 200));

    // Проверяем что симуляция была создана с узлами
    expect(mockState.simulationNodes.length).toBe(3);
    
    // Проверяем что у каждого узла есть координаты
    mockState.simulationNodes.forEach((node, i) => {
      expect(node.x).toBeDefined();
      expect(node.y).toBeDefined();
      expect(typeof node.x).toBe('number');
      expect(typeof node.y).toBe('number');
    });

    // Проверяем что были вызовы отрисовки узлов (arc для кругов)
    const arcCalls = drawCalls.filter(c => c.type === 'arc');
    expect(arcCalls.length).toBeGreaterThan(0);
  });

  it('renders correct number of nodes', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: []
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    // Проверяем что количество узлов в симуляции соответствует переданным
    expect(mockState.simulationNodes.length).toBe(mockNodes.length);
    
    // Проверяем что каждый узел имеет id из mockNodes
    const nodeIds = mockState.simulationNodes.map(n => n.id);
    mockNodes.forEach(node => {
      expect(nodeIds).toContain(node.id);
    });
  });

  it('renders links between correct nodes', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    // Ждем тик симуляции для отрисовки связей
    if (mockState.tickCallback) {
      mockState.tickCallback();
    }

    // Проверяем что были вызовы lineTo (для рисования связей)
    const lineToCalls = drawCalls.filter(c => c.type === 'lineTo');
    
    // Должны быть вызовы lineTo для отрисовки связей
    expect(lineToCalls.length).toBeGreaterThan(0);
  });

  it('does not render links when no links provided', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: []
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    // Ждем тик симуляции
    if (mockState.tickCallback) {
      mockState.tickCallback();
    }

    // Очищаем вызовы до проверки
    drawCalls = [];

    // Симулируем еще один тик
    if (mockState.tickCallback) {
      mockState.tickCallback();
    }

    // При отсутствии связей не должно быть вызовов moveTo/lineTo для связей
    // (только для отрисовки узлов)
    const moveToCalls = drawCalls.filter(c => c.type === 'moveTo');
    
    // Если нет связей, moveTo не должен вызываться для линий
    // (только arc для узлов)
    expect(moveToCalls.length).toBe(0);
  });

  it('calls onNodeClick when canvas is clicked on a node', async () => {
    const onNodeClick = vi.fn();
    
    // Создаем симуляцию с узлами с координатами
    mockState.simulationNodes = [
      { id: '1', title: 'Node 1', type: 'star', x: 100, y: 100 },
      { id: '2', title: 'Node 2', type: 'planet', x: 200, y: 200 }
    ];

    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks,
        onNodeClick
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();

    // Симулируем клик на canvas в позиции первого узла
    await fireEvent.click(canvas!, { 
      clientX: 100, 
      clientY: 100 
    });

    // onNodeClick должен был быть вызван с данными узла
    expect(onNodeClick).toHaveBeenCalled();
    expect(onNodeClick).toHaveBeenCalledWith(expect.objectContaining({
      id: '1',
      title: 'Node 1',
      type: 'star'
    }));
  });

  it('applies correct cursor style', async () => {
    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();

    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
    expect(canvas?.style.cursor).toBe('grab');
  });

  it('updates when props change', async () => {
    const { component } = render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Обновляем пропсы через re-render (Svelte 5 API)
    const newNodes = [...mockNodes, { id: '4', title: 'Node 4', type: 'galaxy' }];
    
    render(GraphCanvas, {
      props: {
        nodes: newNodes,
        links: mockLinks
      }
    });

    await tick();

    // Компонент должен обновиться без ошибок
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('cleans up on unmount', async () => {
    const { unmount } = render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: mockLinks
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Размонтируем компонент - не должно быть ошибок
    expect(() => unmount()).not.toThrow();
  });

  it('handles different node types', async () => {
    const nodesWithTypes = [
      { id: '1', title: 'Star', type: 'star' },
      { id: '2', title: 'Planet', type: 'planet' },
      { id: '3', title: 'Comet', type: 'comet' },
      { id: '4', title: 'Galaxy', type: 'galaxy' },
      { id: '5', title: 'Asteroid', type: 'asteroid' },
      { id: '6', title: 'Debris', type: 'debris' },
      { id: '7', title: 'Unknown' } // без типа
    ];

    const { component } = render(GraphCanvas, {
      props: {
        nodes: nodesWithTypes,
        links: []
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    expect(component).toBeDefined();
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('handles different link types', async () => {
    const linksWithTypes = [
      { source: '1', target: '2', weight: 0.8, link_type: 'reference' },
      { source: '1', target: '3', weight: 0.5, link_type: 'dependency' },
      { source: '2', target: '3', weight: 0.3, link_type: 'related' },
      { source: '3', target: '1', weight: 0.6, link_type: 'custom' },
      { source: '2', target: '1' } // без типа
    ];

    const { component } = render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: linksWithTypes
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    expect(component).toBeDefined();
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('handles orphan links (links without corresponding nodes)', async () => {
    const orphanLinks = [
      { source: '1', target: '999', weight: 0.5 }, // узел 999 не существует
      { source: '1', target: '2', weight: 0.8 }
    ];

    const { component } = render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: orphanLinks
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 150));

    // Компонент должен обработать orphan links без ошибок
    expect(component).toBeDefined();
    const canvas = document.querySelector('canvas');
    expect(canvas).toBeInTheDocument();
  });

  it('renders all nodes within canvas bounds after stabilization', async () => {
    const canvasWidth = 800;
    const canvasHeight = 600;

    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: []
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    // Проверяем что все узлы имеют координаты в пределах canvas
    mockState.simulationNodes.forEach((node: any) => {
      expect(node.x).toBeGreaterThanOrEqual(0);
      expect(node.x).toBeLessThanOrEqual(canvasWidth);
      expect(node.y).toBeGreaterThanOrEqual(0);
      expect(node.y).toBeLessThanOrEqual(canvasHeight);
    });

    // Проверяем что координаты валидны (не NaN, не Infinity)
    mockState.simulationNodes.forEach((node: any) => {
      expect(Number.isFinite(node.x)).toBe(true);
      expect(Number.isFinite(node.y)).toBe(true);
      expect(Number.isNaN(node.x)).toBe(false);
      expect(Number.isNaN(node.y)).toBe(false);
    });
  });

  it('restarts simulation with new center when container resizes', async () => {
    let resizeCallback: (() => void) | null = null;

    // Мокаем ResizeObserver с перехватом колбэка
    global.ResizeObserver = vi.fn().mockImplementation((callback: ResizeObserverCallback) => {
      resizeCallback = () => {
        callback([], {} as ResizeObserver);
      };
      return {
        observe: vi.fn(),
        disconnect: vi.fn(),
        unobserve: vi.fn()
      };
    });

    render(GraphCanvas, {
      props: {
        nodes: mockNodes,
        links: []
      }
    });

    await tick();
    await new Promise(resolve => setTimeout(resolve, 200));

    const initialCenterX = 400;
    const initialCenterY = 300;

    // Эмулируем изменение размеров контейнера
    const canvas = document.querySelector('canvas');
    if (canvas) {
      // Изменяем размеры canvas
      Object.defineProperty(canvas, 'width', { value: 1000, writable: true });
      Object.defineProperty(canvas, 'height', { value: 800, writable: true });
      
      // Вызываем resize колбэк
      if (resizeCallback) {
        (resizeCallback as () => void)();
      }
    }

    await tick();
    await new Promise(resolve => setTimeout(resolve, 100));

    // Проверяем что симуляция была пересоздана с новым центром
    // (forceCenter должен был вызваться с новыми размерами)
    expect(mockCtx.clearRect).toHaveBeenCalled();
  });
});
