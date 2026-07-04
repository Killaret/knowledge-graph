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

    const particleCount = 5 + Math.floor(Math.random() * 5); // 5-10 particles
    const particles: Particle[] = [];

    for (let i = 0; i < particleCount; i++) {
      const angle = (Math.PI * 2 * i) / particleCount;
      const distance = 20 + Math.random() * 15;
      const speed = 0.005 + Math.random() * 0.01;

      particles.push({
        x,
        y,
        angle,
        distance,
        speed,
        size: 1 + Math.random(),
        color,
        alpha: 0.3 + Math.random() * 0.4
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
      particle.angle += particle.speed;
      particle.x = centerX + Math.cos(particle.angle) * particle.distance;
      particle.y = centerY + Math.sin(particle.angle) * particle.distance;
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
      ctx.fillStyle = particle.color.replace(')', `, ${particle.alpha})`).replace('rgb', 'rgba');
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
