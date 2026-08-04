/**
 * Lightweight particle system for orbiting particles around nodes
 * Uses native Canvas API for maximum performance
 */

export interface Particle {
  x: number;
  y: number;
  angle: number;
  distance: number;
  speed: number;
  direction: number;
  tilt: number;
  size: number;
  color: string;
  alpha: number;
}

export class ParticleSystem {
  private particles: Map<string, Particle[]> = new Map();
  private enabled: boolean = true;

  constructor(private nodeCount: number) {
    this.enabled = nodeCount <= 100;
  }

  /**
   * Initialize particles for a node
   */
  initParticles(nodeId: string, x: number, y: number, color: string): void {
    if (!this.enabled) return;
    if (this.particles.has(nodeId)) return;

    const particleCount = 4 + Math.floor(Math.random() * 6); // 4-9 particles
    const particles: Particle[] = [];

    for (let i = 0; i < particleCount; i++) {
      const angle = (Math.PI * 2 * i) / particleCount + Math.random() * Math.PI;
      const distance = 18 + Math.random() * 28;
      const speed = 0.002 + Math.random() * 0.018;
      const direction = Math.random() > 0.5 ? 1 : -1;

      particles.push({
        x,
        y,
        angle,
        distance,
        speed,
        direction,
        tilt: Math.random() * Math.PI,
        size: 1 + Math.random() * 2,
        color,
        alpha: 0.25 + Math.random() * 0.55,
      });
    }

    this.particles.set(nodeId, particles);
  }

  /**
   * Update particle positions
   */
  update(nodeId: string, centerX: number, centerY: number): void {
    if (!this.enabled) return;

    const particles = this.particles.get(nodeId);
    if (!particles) return;

    for (const particle of particles) {
      particle.angle += particle.speed * particle.direction;

      // Slightly elliptical orbit per particle with its own tilt so orbits
      // around different nodes look distinct and not perfectly circular.
      const effectiveAngle = particle.angle + particle.tilt;
      particle.x = centerX + Math.cos(effectiveAngle) * particle.distance;
      particle.y = centerY + Math.sin(effectiveAngle) * particle.distance * 0.7;
    }
  }

  /**
   * Draw particles for a node
   */
  draw(ctx: CanvasRenderingContext2D, nodeId: string): void {
    if (!this.enabled) return;

    const particles = this.particles.get(nodeId);
    if (!particles) return;

    for (const particle of particles) {
      ctx.beginPath();
      ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
      ctx.fillStyle = particle.color.replace(")", `, ${particle.alpha})`).replace("rgb", "rgba");
      ctx.fill();
    }
  }

  /**
   * Remove particles for a node
   */
  removeParticles(nodeId: string): void {
    this.particles.delete(nodeId);
  }

  /**
   * Clear all particles
   */
  clear(): void {
    this.particles.clear();
  }

  /**
   * Check if particles are enabled
   */
  isEnabled(): boolean {
    return this.enabled;
  }
}
