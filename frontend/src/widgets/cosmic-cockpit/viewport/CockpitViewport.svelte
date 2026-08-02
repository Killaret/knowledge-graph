<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { cockpitStore } from "$features/cosmic-cockpit";

  interface Props {
    /** Optional external nodes to render as distant stars. */
    nodes?: Array<{ id: string; x?: number; y?: number; type?: string }>;
  }

  const { nodes = [] }: Props = $props();

  let canvas: HTMLCanvasElement | null = $state(null);
  let ctx: CanvasRenderingContext2D | null = $state(null);
  let width = $state(0);
  let height = $state(0);
  let mouseX = $state(0.5);
  let mouseY = $state(0.5);
  let rafId: number | null = $state(null);

  interface Star {
    x: number;
    y: number;
    z: number;
    size: number;
    speed: number;
    color: string;
  }

  const STAR_COLORS = ["#ffffff", "#2dd4bf", "#c026d3", "#facc15", "#60a5fa"];

  function createStars(count: number): Star[] {
    const stars: Star[] = [];
    for (let i = 0; i < count; i++) {
      stars.push({
        x: Math.random() * width,
        y: Math.random() * height,
        z: Math.random() * 2 + 0.2,
        size: Math.random() * 1.5 + 0.5,
        speed: Math.random() * 0.3 + 0.05,
        color: STAR_COLORS[Math.floor(Math.random() * STAR_COLORS.length)],
      });
    }
    return stars;
  }

  let stars = $state<Star[]>([]);

  function resize() {
    if (!canvas || !ctx) return;
    const rect = canvas.getBoundingClientRect();
    width = rect.width;
    height = rect.height;
    canvas.width = width * window.devicePixelRatio;
    canvas.height = height * window.devicePixelRatio;
    ctx.scale(window.devicePixelRatio, window.devicePixelRatio);

    if (stars.length === 0) {
      stars = createStars(Math.min(120, Math.floor((width * height) / 4000)) || 40);
    }
  }

  function drawGrid() {
    if (!ctx) return;
    ctx.strokeStyle = "rgba(45, 212, 191, 0.05)";
    ctx.lineWidth = 1;
    const gridSize = 40;
    const offsetX = (mouseX - 0.5) * 8;
    const offsetY = (mouseY - 0.5) * 8;

    for (let x = offsetX % gridSize; x < width; x += gridSize) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    }
    for (let y = offsetY % gridSize; y < height; y += gridSize) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }
  }

  function drawStars() {
    if (!ctx) return;
    const c = ctx;
    const parallaxX = (mouseX - 0.5) * 2;
    const parallaxY = (mouseY - 0.5) * 2;

    stars.forEach((star) => {
      star.x -= star.speed;
      if (star.x < 0) {
        star.x = width;
        star.y = Math.random() * height;
      }

      const x = star.x + parallaxX * star.z;
      const y = star.y + parallaxY * star.z;

      c.fillStyle = star.color;
      c.globalAlpha = Math.min(1, star.z / 1.5);
      c.beginPath();
      c.arc(x, y, star.size, 0, Math.PI * 2);
      c.fill();
    });
    c.globalAlpha = 1;
  }

  function drawNodes() {
    if (!ctx || nodes.length === 0) return;
    const c = ctx;
    // Render distant note nodes as faint constellation points.
    nodes.forEach((node) => {
      const x = (node.x ?? 0.5) * width;
      const y = (node.y ?? 0.5) * height;
      c.fillStyle = "rgba(45, 212, 191, 0.6)";
      c.beginPath();
      c.arc(x, y, 2, 0, Math.PI * 2);
      c.fill();
    });
  }

  function drawCrosshair() {
    if (!ctx) return;
    const cx = width / 2 + (mouseX - 0.5) * 12;
    const cy = height / 2 + (mouseY - 0.5) * 12;
    ctx.strokeStyle = "rgba(45, 212, 191, 0.25)";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.arc(cx, cy, 20, 0, Math.PI * 2);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(cx - 30, cy);
    ctx.lineTo(cx + 30, cy);
    ctx.moveTo(cx, cy - 30);
    ctx.lineTo(cx, cy + 30);
    ctx.stroke();
  }

  function loop() {
    if (!ctx || !canvas) return;
    ctx.clearRect(0, 0, width, height);

    // deep space background
    const gradient = ctx.createRadialGradient(width / 2, height / 2, 0, width / 2, height / 2, Math.max(width, height));
    gradient.addColorStop(0, "rgba(10, 10, 15, 0)");
    gradient.addColorStop(1, "rgba(20, 20, 43, 0.4)");
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, width, height);

    drawGrid();
    drawStars();
    drawNodes();
    drawCrosshair();

    rafId = requestAnimationFrame(loop);
  }

  function handleMouseMove(e: MouseEvent) {
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    mouseX = (e.clientX - rect.left) / rect.width;
    mouseY = (e.clientY - rect.top) / rect.height;
  }

  onMount(() => {
    if (!browser || !canvas) return;
    ctx = canvas.getContext("2d");
    resize();
    loop();

    const resizeObserver = new ResizeObserver(() => resize());
    resizeObserver.observe(canvas.parentElement ?? canvas);

    const move = (e: MouseEvent) => handleMouseMove(e);
    canvas.addEventListener("mousemove", move);

    return () => {
      if (rafId) cancelAnimationFrame(rafId);
      resizeObserver.disconnect();
      canvas?.removeEventListener("mousemove", move);
    };
  });
</script>

<div class="cockpit-viewport" data-testid="cockpit-viewport">
  <canvas bind:this={canvas} class="viewport-canvas"></canvas>
  <div class="viewport-overlay"></div>
</div>

<style>
  .cockpit-viewport {
    position: relative;
    width: 100%;
    height: 100%;
    min-height: 0;
    overflow: hidden;
    border-radius: 10px;
    border: 1px solid rgba(45, 212, 191, 0.2);
    background: radial-gradient(ellipse at 50% 100%, #14142b 0%, #0a0a0f 80%);
  }

  .viewport-canvas {
    display: block;
    width: 100%;
    height: 100%;
  }

  .viewport-overlay {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background: linear-gradient(
      to bottom,
      rgba(45, 212, 191, 0.06) 0%,
      transparent 30%,
      transparent 70%,
      rgba(192, 38, 211, 0.06) 100%
    );
    box-shadow: inset 0 0 30px rgba(0, 0, 0, 0.5);
  }
</style>
