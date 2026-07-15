<script lang="ts">
  import { browser } from '$app/environment';

  interface Star {
    x: number;
    y: number;
    size: number;
    opacity: number;
    speed: number;
    twinklePhase: number;
  }

  interface Nebula {
    x: number;
    y: number;
    radius: number;
    color: string;
    opacity: number;
  }

  // Configuration
  const STAR_COUNT = 150;
  const NEBULA_COUNT = 3;
  
  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D | null = null;
  let animationId: number;
  let stars: Star[] = [];
  let nebulas: Nebula[] = [];
  let mouseX = 0;
  let mouseY = 0;
  let isActive = true;

  function generateStars(): Star[] {
    const stars: Star[] = [];
    for (let i = 0; i < STAR_COUNT; i++) {
      stars.push({
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        size: Math.random() * 2 + 0.5,
        opacity: Math.random() * 0.8 + 0.2,
        speed: Math.random() * 0.02 + 0.005,
        twinklePhase: Math.random() * Math.PI * 2
      });
    }
    return stars;
  }

  function generateNebulas(): Nebula[] {
    const colors = [
      'rgba(64, 169, 255, 0.15)',   // Blue
      'rgba(255, 204, 0, 0.1)',     // Gold
      'rgba(139, 0, 0, 0.12)',      // Dark red
      'rgba(167, 139, 250, 0.1)'    // Purple
    ];
    
    const nebulas: Nebula[] = [];
    for (let i = 0; i < NEBULA_COUNT; i++) {
      nebulas.push({
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        radius: Math.random() * 300 + 200,
        color: colors[Math.floor(Math.random() * colors.length)],
        opacity: Math.random() * 0.5 + 0.3
      });
    }
    return nebulas;
  }

  function drawNebula(nebula: Nebula) {
    if (!ctx) return;
    
    const gradient = ctx.createRadialGradient(
      nebula.x, nebula.y, 0,
      nebula.x, nebula.y, nebula.radius
    );
    
    gradient.addColorStop(0, nebula.color);
    gradient.addColorStop(1, 'transparent');
    
    ctx.fillStyle = gradient;
    ctx.globalAlpha = nebula.opacity;
    ctx.beginPath();
    ctx.arc(nebula.x, nebula.y, nebula.radius, 0, Math.PI * 2);
    ctx.fill();
    ctx.globalAlpha = 1;
  }

  function drawStar(star: Star, time: number) {
    if (!ctx) return;

    // Twinkle effect
    const twinkle = Math.sin(time * star.speed + star.twinklePhase);
    const currentOpacity = star.opacity * (0.7 + twinkle * 0.3);
    
    // Parallax effect based on mouse position
    const parallaxX = (mouseX - window.innerWidth / 2) * star.speed * 0.5;
    const parallaxY = (mouseY - window.innerHeight / 2) * star.speed * 0.5;
    
    let x = star.x + parallaxX;
    let y = star.y + parallaxY;
    
    // Wrap around screen
    if (x < 0) x += window.innerWidth;
    if (x > window.innerWidth) x -= window.innerWidth;
    if (y < 0) y += window.innerHeight;
    if (y > window.innerHeight) y -= window.innerHeight;

    // Draw star glow
    const gradient = ctx.createRadialGradient(x, y, 0, x, y, star.size * 3);
    gradient.addColorStop(0, `rgba(255, 255, 255, ${currentOpacity})`);
    gradient.addColorStop(0.5, `rgba(255, 255, 255, ${currentOpacity * 0.3})`);
    gradient.addColorStop(1, 'transparent');
    
    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.arc(x, y, star.size * 3, 0, Math.PI * 2);
    ctx.fill();
    
    // Draw star core
    ctx.fillStyle = `rgba(255, 255, 255, ${currentOpacity})`;
    ctx.beginPath();
    ctx.arc(x, y, star.size, 0, Math.PI * 2);
    ctx.fill();
  }

  function drawShootingStar(time: number) {
    if (!ctx) return;
    
    // Occasional shooting star
    const shootingStarPhase = (time * 0.0005) % 10;
    if (shootingStarPhase < 1) {
      const startX = window.innerWidth * 0.8;
      const startY = window.innerHeight * 0.2;
      const progress = shootingStarPhase;
      
      const x = startX - progress * 300;
      const y = startY + progress * 150;
      
      // Draw trail
      const gradient = ctx.createLinearGradient(x, y, x + 50, y - 25);
      gradient.addColorStop(0, 'rgba(255, 255, 255, 0)');
      gradient.addColorStop(0.5, 'rgba(255, 255, 255, 0.8)');
      gradient.addColorStop(1, 'rgba(255, 255, 255, 0)');
      
      ctx.strokeStyle = gradient;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x, y);
      ctx.lineTo(x + 50, y - 25);
      ctx.stroke();
      
      // Draw head
      ctx.fillStyle = 'rgba(255, 255, 255, 1)';
      ctx.beginPath();
      ctx.arc(x, y, 2, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  function animate(time: number) {
    if (!isActive || !ctx || !canvas) return;
    
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    // Draw nebulas first (background layer)
    nebulas.forEach(nebula => drawNebula(nebula));
    
    // Draw stars
    stars.forEach(star => drawStar(star, time));
    
    // Draw occasional shooting star
    drawShootingStar(time);
    
    animationId = requestAnimationFrame(animate);
  }

  function handleResize() {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    stars = generateStars();
    nebulas = generateNebulas();
  }

  $effect(() => {
    if (!browser || !canvas) return;

    ctx = canvas.getContext('2d');
    if (!ctx) return;

    handleResize();

    const onMove = (e: MouseEvent) => {
      mouseX = e.clientX;
      mouseY = e.clientY;
    };
    const onResize = () => handleResize();
    const onVisibility = () => {
      if (document.hidden) {
        isActive = false;
        cancelAnimationFrame(animationId);
      } else {
        isActive = true;
        animationId = requestAnimationFrame(animate);
      }
    };

    window.addEventListener('mousemove', onMove);
    window.addEventListener('resize', onResize);
    document.addEventListener('visibilitychange', onVisibility);

    isActive = true;
    animationId = requestAnimationFrame(animate);

    return () => {
      isActive = false;
      cancelAnimationFrame(animationId);
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('resize', onResize);
      document.removeEventListener('visibilitychange', onVisibility);
    };
  });
</script>

<canvas
  bind:this={canvas}
  class="cosmic-background"
  aria-hidden="true"
></canvas>

<style>
  .cosmic-background {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 0;
    pointer-events: none;
    background: radial-gradient(ellipse at 50% 100%, #0a0a1a 0%, #000 80%);
  }
</style>
