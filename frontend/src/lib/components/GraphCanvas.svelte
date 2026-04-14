<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';

  const { 
    nodes, 
    links,
    onNodeClick
  }: { 
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  } = $props();

  // Цвета для разных типов связей
  const linkTypeColors: Record<string, string> = {
    'reference': '#3366ff',   // Синий - ссылочная связь
    'dependency': '#ff6600',  // Оранжевый - зависимость
    'related': '#999999',     // Серый - связанная тема (по умолчанию)
    'custom': '#ff66ff',      // Розовый - пользовательская
  };

  // Функция для получения цвета связи по типу
  function getLinkColor(weight: number, linkType?: string): string {
    const effectiveType = linkType || 'related';
    const color = linkTypeColors[effectiveType] || linkTypeColors['related'];
    const opacity = 0.4 + (weight ?? 0.5) * 0.4;
    
    // Конвертируем hex в rgba
    const r = parseInt(color.slice(1, 3), 16);
    const g = parseInt(color.slice(3, 5), 16);
    const b = parseInt(color.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${opacity})`;
  }

  // Функция для получения стиля линии по типу связи и весу
  function getLineDash(linkType?: string, weight?: number): number[] {
    const effectiveType = linkType || 'related';
    
    switch (effectiveType) {
      case 'reference': 
        return []; // Сплошная
      case 'dependency': 
        return [10, 3]; // Штрихпунктир
      case 'related': 
        // Пунктир только при слабом весе
        return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
      case 'custom': 
        return [2, 6]; // Точечный
      default: 
        return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    }
  }

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let width = 800;
  let height = 600;
  let animationId: number;
  const angles: Map<string, number> = new Map();
  const speeds: Map<string, number> = new Map();
  let d3Force: typeof import('d3-force') | null = null;

  const transform = $state({ x: 0, y: 0, k: 1 });
  let dragging = $state(false);
  let dragStart = $state({ x: 0, y: 0 });
  let simulation: any = null;

  onMount(() => {
    if (!browser) return;
    
    // Dynamic import for SSR safety
    let cleanup = () => {};
    
    import('d3-force').then(d3 => {
      d3Force = d3;
      ctx = canvas.getContext('2d')!;
      resize();
      window.addEventListener('resize', resize);
      startSimulation();
      startAnimation();
      
      cleanup = () => {
        window.removeEventListener('resize', resize);
        if (simulation) simulation.stop();
        cancelAnimationFrame(animationId);
      };
    });
    
    return () => cleanup();
  });

  // Реактивно перезапускаем симуляцию при изменении данных
  $effect(() => {
    // Отслеживаем изменения в данных
    const nodesKey = nodes.map(n => n.id).join(',');
    const linksKey = links.map(l => `${l.source}-${l.target}`).join(',');
    const dataKey = `${nodesKey}|${linksKey}`;
    
    console.log('[GraphCanvas] $effect triggered:', { nodes: nodes.length, links: links.length, dataKey: dataKey.slice(0, 50) });
    
    if (d3Force) {
      if (nodes.length === 0) {
        console.log('[GraphCanvas] No nodes to simulate, stopping simulation');
        if (simulation) {
          simulation.stop();
          simulation = null;
        }
        // Clear canvas
        if (ctx) {
          ctx.clearRect(0, 0, width, height);
        }
        return;
      }
      
      console.log('[GraphCanvas] Restarting simulation with', nodes.length, 'nodes and', links.length, 'links');
      
      // Останавливаем старую симуляцию
      if (simulation) {
        console.log('[GraphCanvas] Stopping old simulation');
        simulation.stop();
      }
      
      // Очищаем углы и скорости для новых данных
      angles.clear();
      speeds.clear();
      console.log('[GraphCanvas] Cleared angles and speeds maps');
      
      // Запускаем новую с обновленными данными
      startSimulation();
      console.log('[GraphCanvas] New simulation started');
    } else {
      console.log('[GraphCanvas] d3Force not loaded yet, skipping simulation restart');
    }
  });

  function startAnimation() {
    function animate() {
      for (const node of simulation?.nodes() || []) {
        const id = node.id;
        const type = node.type || 'star';
        let baseSpeed = 0.005; // звезда
        if (type === 'planet') baseSpeed = 0.02;
        else if (type === 'comet') baseSpeed = 0.03;
        else if (type === 'galaxy') baseSpeed = 0.01;
        let speed = speeds.get(id);
        if (speed === undefined) {
          speed = baseSpeed * (0.7 + Math.random() * 0.6);
          speeds.set(id, speed);
        }
        const current = angles.get(id) || 0;
        angles.set(id, current + speed);
      }
      draw();
      animationId = requestAnimationFrame(animate);
    }
    animate();
  }

  function resize() {
    const rect = canvas.parentElement?.getBoundingClientRect();
    if (rect) {
      width = rect.width;
      height = rect.height;
      canvas.width = width;
      canvas.height = height;
    }
  }

  function startSimulation() {
    if (!d3Force) {
      console.log('[GraphCanvas] Cannot start simulation: d3Force not loaded');
      return;
    }
    
    console.log('[GraphCanvas] startSimulation: creating simulation with', nodes.length, 'nodes');
    
    const simulationNodes = nodes.map(n => ({ ...n, x: width/2, y: height/2 }));
    const edges = links.map(l => ({ source: l.source, target: l.target, weight: l.weight ?? 1, link_type: l.link_type }));

    simulation = d3Force.forceSimulation(simulationNodes as any)
      .force('link', d3Force.forceLink(edges).id((d: any) => d.id).distance(150).strength(0.5))
      .force('charge', d3Force.forceManyBody().strength(-300))
      .force('center', d3Force.forceCenter(width/2, height/2))
      .force('collision', d3Force.forceCollide().radius(40))
      .alphaDecay(0.02)
      .on('tick', () => draw());

    // Run initial ticks to stabilize
    simulation.tick(100);
    console.log('[GraphCanvas] Simulation initialized and ticked 100 times');
  }

  function drawStar(x: number, y: number, r: number, angle: number) {
    if (!ctx) return;
    const points = 5;
    const outerRadius = r;
    const innerRadius = r * 0.4;
    let rot = angle;
    const step = Math.PI / points;

    ctx.beginPath();
    for (let i = 0; i < points; i++) {
      const x1 = x + Math.cos(rot) * outerRadius;
      const y1 = y + Math.sin(rot) * outerRadius;
      ctx.lineTo(x1, y1);
      rot += step;
      const x2 = x + Math.cos(rot) * innerRadius;
      const y2 = y + Math.sin(rot) * innerRadius;
      ctx.lineTo(x2, y2);
      rot += step;
    }
    ctx.closePath();
  }

  function drawPlanet(x: number, y: number, r: number, angle: number) {
    if (!ctx) return;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, 2 * Math.PI);
    ctx.fillStyle = '#c9b37c';
    ctx.fill();
    for (let i = -r/2; i <= r/2; i += r/4) {
      ctx.beginPath();
      ctx.ellipse(x, y + i, r * 0.8, r * 0.15, angle, 0, 2 * Math.PI);
      ctx.fillStyle = '#a57c2c';
      ctx.fill();
    }
  }

  function drawComet(x: number, y: number, r: number, angle: number) {
    if (!ctx) return;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, 2 * Math.PI);
    ctx.fillStyle = '#aaffdd';
    ctx.fill();
    const tailLength = 40;
    const tailAngle = angle;
    const tipX = x + Math.cos(tailAngle) * tailLength;
    const tipY = y + Math.sin(tailAngle) * tailLength;
    ctx.beginPath();
    ctx.moveTo(x, y);
    ctx.lineTo(tipX, tipY);
    ctx.lineWidth = 4;
    ctx.strokeStyle = 'rgba(170, 255, 221, 0.6)';
    ctx.stroke();
  }

  function drawGalaxy(x: number, y: number, r: number, angle: number) {
    if (!ctx) return;
    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(angle);
    for (let i = 0; i < 3; i++) {
      ctx.beginPath();
      ctx.ellipse(0, 0, r * (1 - i*0.2), r * 0.4, 0, 0, 2 * Math.PI);
      ctx.fillStyle = `rgba(200, 180, 255, ${0.3 - i*0.1})`;
      ctx.fill();
    }
    ctx.restore();
  }

  function drawAsteroid(x: number, y: number, r: number, _angle: number) {
    if (!ctx) return;
    // Irregular rocky shape
    ctx.beginPath();
    const points = 7;
    for (let i = 0; i < points; i++) {
      const theta = (i / points) * 2 * Math.PI;
      const radiusVariation = 0.7 + Math.random() * 0.3;
      const px = x + Math.cos(theta) * r * radiusVariation;
      const py = y + Math.sin(theta) * r * radiusVariation;
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
    ctx.fillStyle = '#8b7355';
    ctx.fill();
    ctx.strokeStyle = '#5c4a3a';
    ctx.lineWidth = 1;
    ctx.stroke();
  }

  function drawDebris(x: number, y: number, r: number, _angle: number) {
    if (!ctx) return;
    // Scattered small particles
    ctx.fillStyle = 'rgba(150, 150, 150, 0.6)';
    for (let i = 0; i < 5; i++) {
      const offsetX = (Math.random() - 0.5) * r * 2;
      const offsetY = (Math.random() - 0.5) * r * 2;
      ctx.beginPath();
      ctx.arc(x + offsetX, y + offsetY, r * 0.3, 0, 2 * Math.PI);
      ctx.fill();
    }
  }

  function draw() {
    if (!ctx) return;
    ctx.clearRect(0, 0, width, height);
    ctx.save();
    ctx.translate(transform.x, transform.y);
    ctx.scale(transform.k, transform.k);

    // Рисуем связи
    links.forEach(link => {
      const sourceNode = simulation.nodes().find((n: any) => n.id === link.source);
      const targetNode = simulation.nodes().find((n: any) => n.id === link.target);
      if (!sourceNode || !targetNode) return;
      
      ctx.beginPath();
      ctx.moveTo(sourceNode.x, sourceNode.y);
      ctx.lineTo(targetNode.x, targetNode.y);
      
      const weight = link.weight ?? 0.5;
      const linkType = link.link_type;
      
      // Толщина линии зависит от типа и веса
      let lineWidth = Math.max(1, weight * 4);
      if (linkType === 'dependency') lineWidth *= 1.5;
      if (linkType === 'reference') lineWidth *= 0.8;
      
      ctx.lineWidth = lineWidth;
      ctx.strokeStyle = getLinkColor(weight, linkType);
      
      // Устанавливаем dash pattern для пунктирных линий
      const dash = getLineDash(linkType, weight);
      if (dash.length > 0) {
        ctx.setLineDash(dash);
      } else {
        ctx.setLineDash([]);
      }
      
      ctx.stroke();
      ctx.setLineDash([]); // Сброс dash pattern
    });

    const r = 24; // радиус увеличен для лучшей читаемости
    simulation.nodes().forEach((node: any) => {
      const type = node.type || 'star';
      const angle = angles.get(node.id) || 0;

      switch (type) {
        case 'star':
          drawStar(node.x, node.y, r, angle);
          ctx.fillStyle = '#ffdd88';
          ctx.shadowBlur = 10;
          ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
          ctx.fill();
          break;
        case 'planet':
          drawPlanet(node.x, node.y, r, angle);
          break;
        case 'comet':
          drawComet(node.x, node.y, r, angle);
          break;
        case 'galaxy':
          drawGalaxy(node.x, node.y, r, angle);
          break;
        case 'asteroid':
          drawAsteroid(node.x, node.y, r, angle);
          break;
        case 'debris':
          drawDebris(node.x, node.y, r, angle);
          break;
        default:
          drawStar(node.x, node.y, r, angle);
          ctx.fillStyle = '#cccccc';
          ctx.fill();
      }
      ctx.shadowBlur = 0;

      // Текст под звездой
      ctx.font = `bold ${Math.min(14, r * 0.65)}px sans-serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      let title = node.title;
      if (title.length > 14) title = title.slice(0, 12) + '…';
      ctx.shadowBlur = 2;
      ctx.shadowColor = 'rgba(0,0,0,0.5)';
      ctx.fillStyle = '#ffffff';
      ctx.fillText(title, node.x, node.y + r + 6);
      ctx.shadowBlur = 0;
    });

    ctx.restore();
  }

  function handleZoom(e: WheelEvent) {
    e.preventDefault();
    const delta = e.deltaY > 0 ? 0.95 : 1.05;
    const newK = transform.k * delta;
    if (newK < 0.2 || newK > 5) return;
    transform.k = newK;
    draw();
  }

  function handlePanStart(e: MouseEvent) {
    dragging = true;
    dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y };
    canvas.style.cursor = 'grabbing';
  }

  function handlePanMove(e: MouseEvent) {
    if (!dragging) return;
    transform.x = e.clientX - dragStart.x;
    transform.y = e.clientY - dragStart.y;
    draw();
  }

  function handlePanEnd() {
    dragging = false;
    canvas.style.cursor = 'grab';
  }

  function handleClick(e: MouseEvent) {
    const rect = canvas.getBoundingClientRect();
    const clickX = (e.clientX - rect.left - transform.x) / transform.k;
    const clickY = (e.clientY - rect.top - transform.y) / transform.k;
    const node = simulation.nodes().find((n: any) => {
      const dx = n.x - clickX;
      const dy = n.y - clickY;
      return Math.hypot(dx, dy) < 24;
    });
    if (node) {
      if (onNodeClick) {
        onNodeClick({ id: node.id, title: node.title, type: node.type });
      } else {
        goto(`/notes/${node.id}`);
      }
    }
  }
</script>

<canvas
  bind:this={canvas}
  onmousedown={handlePanStart}
  onmousemove={handlePanMove}
  onmouseup={handlePanEnd}
  onclick={handleClick}
  onwheel={handleZoom}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
></canvas>