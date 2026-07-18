/**
 * CelestialBody — Value Object for a graph node type.
 *
 * Encapsulates all visual parameters of a node (colors, emoji, radius bounds,
 * animation speed, gravity influence) and delegates rendering to a draw function
 * that is injected by the renderer layer. This keeps the domain object free of
 * Canvas/Three.js details while still making the rendering dispatch data-driven.
 */

export interface CelestialBodyDrawContext {
  x: number;
  y: number;
  /** Radius already scaled by CelestialBody.baseRadius. */
  r: number;
  angle: number;
  nodeId: string;
  nodeCount?: number;
  time?: number;
  variation?: {
    sizeMultiplier: number;
    hueShift: number;
    phaseShift?: number;
  };
  disableVariation?: boolean;
  /** Whether the global shadow setting is active for this frame. */
  enableShadows?: boolean;
  focusMode?: boolean;
}

export type CelestialBodyDrawFunction = (
  ctx: CanvasRenderingContext2D,
  context: CelestialBodyDrawContext,
) => void;

export interface CelestialBodyProps {
  type: string;
  label: string;
  emoji: string;
  /** Primary hex color used for the node body. */
  color: string;
  /** Glow color used for shadows / accretion disks. */
  glowColor: string;
  /**
   * Multiplier applied to the base radius (16 px) before drawing.
   * e.g. 0.6 for satellites, 1.5 for nebulas.
   */
  baseRadius: number;
  /**
   * Lower bound for the size multiplier produced by getVariation.
   * Kept as "minRadius" to match the requested domain vocabulary.
   */
  minRadius: number;
  /**
   * Upper bound for the size multiplier produced by getVariation.
   */
  maxRadius: number;
  /** Arbitrary mass used by the gravity distortion system. */
  gravityMass: number;
  /** Base rotation speed for the animation loop. */
  baseSpeed: number;
  /** Max pixel offset contributed to the gravity lens distortion. */
  gravityOffset: number;
  /** Optional CSS custom property name for components that use CSS variables. */
  cssVarName?: string;
  /** True for the four explicit anomaly renderers and the unknown dispatcher. */
  isAnomaly?: boolean;
  /** True for types that should appear in the note creation TypeSelector. */
  isUi?: boolean;
  anomalyType?:
    "reality_rift" | "chromatic_maw" | "void_whisper" | "cosmic_abomination";
  drawFunction?: CelestialBodyDrawFunction;
}

export class CelestialBody {
  /** Renderer adapter — set by the Canvas renderer at module load. */
  drawFunction?: CelestialBodyDrawFunction;

  constructor(public readonly props: Readonly<CelestialBodyProps>) {
    this.drawFunction = props.drawFunction;
  }

  get type(): string {
    return this.props.type;
  }
  get label(): string {
    return this.props.label;
  }
  get emoji(): string {
    return this.props.emoji;
  }
  get color(): string {
    return this.props.color;
  }
  get glowColor(): string {
    return this.props.glowColor;
  }
  get baseRadius(): number {
    return this.props.baseRadius;
  }
  get minRadius(): number {
    return this.props.minRadius;
  }
  get maxRadius(): number {
    return this.props.maxRadius;
  }
  get gravityMass(): number {
    return this.props.gravityMass;
  }
  get baseSpeed(): number {
    return this.props.baseSpeed;
  }
  get gravityOffset(): number {
    return this.props.gravityOffset;
  }
  get cssVarName(): string {
    return this.props.cssVarName ?? `--color-${this.type}`;
  }
  get isAnomaly(): boolean {
    return this.props.isAnomaly ?? false;
  }
  get isUi(): boolean {
    return this.props.isUi ?? false;
  }
  get anomalyType():
    | "reality_rift"
    | "chromatic_maw"
    | "void_whisper"
    | "cosmic_abomination"
    | undefined {
    return this.props.anomalyType;
  }

  draw(ctx: CanvasRenderingContext2D, context: CelestialBodyDrawContext): void {
    if (!this.drawFunction) {
      throw new Error(
        `Draw function not registered for celestial body type "${this.type}"`,
      );
    }
    this.drawFunction(ctx, context);
  }

  /** Returns a CSS var(...) expression with the hard-coded color as fallback. */
  toCSSColor(): string {
    return `var(${this.cssVarName}, ${this.color})`;
  }

  /** Case-insensitive lookup. Falls back to UNKNOWN. */
  static fromString(type: string | undefined): CelestialBody {
    const normalized = (type ?? "").toLowerCase().trim();
    return CelestialBody.MAP.get(normalized) ?? CelestialBody.UNKNOWN;
  }

  static readonly STAR = new CelestialBody({
    type: "star",
    label: "Star",
    emoji: "⭐",
    color: "#ffcc00",
    glowColor: "#ffcc00",
    baseRadius: 1,
    minRadius: 0.8,
    maxRadius: 1.2,
    gravityMass: 100,
    baseSpeed: 0.005,
    gravityOffset: 20,
    cssVarName: "--color-star",
    isUi: true,
  });

  static readonly PLANET = new CelestialBody({
    type: "planet",
    label: "Planet",
    emoji: "🪐",
    color: "#d6aa5d",
    glowColor: "#d6aa5d",
    baseRadius: 1,
    minRadius: 0.8,
    maxRadius: 1.2,
    gravityMass: 80,
    baseSpeed: 0.02,
    gravityOffset: 15,
    cssVarName: "--color-planet",
    isUi: true,
  });

  static readonly MOON = new CelestialBody({
    type: "moon",
    label: "Moon",
    emoji: "🌙",
    color: "#cccccc",
    glowColor: "#aaaaaa",
    baseRadius: 0.9,
    minRadius: 0.8,
    maxRadius: 1.2,
    gravityMass: 50,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-moon",
    isUi: false,
  });

  static readonly COMET = new CelestialBody({
    type: "comet",
    label: "Comet",
    emoji: "☄️",
    color: "#e879f9",
    glowColor: "#e879f9",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 30,
    baseSpeed: 0.03,
    gravityOffset: 10,
    cssVarName: "--color-comet",
    isUi: true,
  });

  static readonly GALAXY = new CelestialBody({
    type: "galaxy",
    label: "Galaxy",
    emoji: "🌀",
    color: "#8b5cf6",
    glowColor: "#8b5cf6",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 90,
    baseSpeed: 0.01,
    gravityOffset: 10,
    cssVarName: "--color-galaxy",
    isUi: true,
  });

  static readonly NEBULA = new CelestialBody({
    type: "nebula",
    label: "Nebula",
    emoji: "💫",
    color: "#2dd4bf",
    glowColor: "#2dd4bf",
    baseRadius: 1.5,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 20,
    baseSpeed: 0.008,
    gravityOffset: 10,
    cssVarName: "--color-nebula",
    isUi: true,
  });

  static readonly ASTEROID = new CelestialBody({
    type: "asteroid",
    label: "Asteroid",
    emoji: "🌑",
    color: "#94a3b8",
    glowColor: "#94a3b8",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 40,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-asteroid",
    isUi: true,
  });

  static readonly SATELLITE = new CelestialBody({
    type: "satellite",
    label: "Satellite",
    emoji: "🛰️",
    color: "#a1a1aa",
    glowColor: "#a1a1aa",
    baseRadius: 0.6,
    minRadius: 0.8,
    maxRadius: 1.2,
    gravityMass: 10,
    baseSpeed: 0.02,
    gravityOffset: 10,
    cssVarName: "--color-satellite",
    isUi: true,
  });

  static readonly BLACKHOLE = new CelestialBody({
    type: "blackhole",
    label: "Black Hole",
    emoji: "⚫",
    color: "#000000",
    glowColor: "#ff6600",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 200,
    baseSpeed: 0,
    gravityOffset: 25,
    cssVarName: "--color-blackhole",
    isUi: false,
  });

  static readonly DEBRIS = new CelestialBody({
    type: "debris",
    label: "Debris",
    emoji: "🌌",
    color: "#71717a",
    glowColor: "#71717a",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 5,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-debris",
    isUi: true,
  });

  static readonly DUST = new CelestialBody({
    type: "dust",
    label: "Cosmic Dust",
    emoji: "🌫️",
    color: "#a0a0a0",
    glowColor: "#a0a0a0",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 1,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-dust",
    isUi: true,
  });

  static readonly TECHNICAL = new CelestialBody({
    type: "technical",
    label: "Technical",
    emoji: "❓",
    color: "#8a2be2",
    glowColor: "#a78bfa",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0,
    gravityOffset: 10,
    cssVarName: "--color-technical",
    isUi: false,
  });

  static readonly UNKNOWN = new CelestialBody({
    type: "unknown",
    label: "Unknown",
    emoji: "❓",
    color: "#94a3b8",
    glowColor: "#94a3b8",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-unknown",
    isAnomaly: true,
    isUi: false,
  });

  static readonly REALITY_RIFT = new CelestialBody({
    type: "reality_rift",
    label: "Reality Rift",
    emoji: "🌑",
    color: "#6b21a8",
    glowColor: "#a855f7",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-reality-rift",
    isAnomaly: true,
    isUi: false,
    anomalyType: "reality_rift",
  });

  static readonly CHROMATIC_MAW = new CelestialBody({
    type: "chromatic_maw",
    label: "Chromatic Maw",
    emoji: "🐙",
    color: "#ff00aa",
    glowColor: "#ff66cc",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-chromatic-maw",
    isAnomaly: true,
    isUi: false,
    anomalyType: "chromatic_maw",
  });

  static readonly VOID_WHISPER = new CelestialBody({
    type: "void_whisper",
    label: "Void Whisper",
    emoji: "👁️",
    color: "#00ffff",
    glowColor: "#66ffff",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-void-whisper",
    isAnomaly: true,
    isUi: false,
    anomalyType: "void_whisper",
  });

  static readonly COSMIC_ABOMINATION = new CelestialBody({
    type: "cosmic_abomination",
    label: "Cosmic Abomination",
    emoji: "🌀",
    color: "#6600ff",
    glowColor: "#9966ff",
    baseRadius: 1,
    minRadius: 0.7,
    maxRadius: 1.3,
    gravityMass: 0,
    baseSpeed: 0.005,
    gravityOffset: 10,
    cssVarName: "--color-cosmic-abomination",
    isAnomaly: true,
    isUi: false,
    anomalyType: "cosmic_abomination",
  });

  private static readonly ALL = [
    CelestialBody.STAR,
    CelestialBody.PLANET,
    CelestialBody.MOON,
    CelestialBody.COMET,
    CelestialBody.GALAXY,
    CelestialBody.NEBULA,
    CelestialBody.ASTEROID,
    CelestialBody.SATELLITE,
    CelestialBody.BLACKHOLE,
    CelestialBody.DEBRIS,
    CelestialBody.DUST,
    CelestialBody.TECHNICAL,
    CelestialBody.UNKNOWN,
    CelestialBody.REALITY_RIFT,
    CelestialBody.CHROMATIC_MAW,
    CelestialBody.VOID_WHISPER,
    CelestialBody.COSMIC_ABOMINATION,
  ] as const;

  private static readonly MAP = new Map(
    CelestialBody.ALL.map((body) => [body.type, body]),
  );

  static readonly UI_TYPES = CelestialBody.ALL.filter((body) => body.isUi);
  static readonly ANOMALIES = CelestialBody.ALL.filter(
    (body) => body.isAnomaly,
  );
}
