declare module 'd3-force-3d' {
  export interface Simulation<NodeDatum extends SimulationNodeDatum, LinkDatum extends SimulationLinkDatum<NodeDatum>> {
    alpha(value?: number): number;
    alphaMin(value?: number): number;
    alphaDecay(value?: number): number;
    alphaTarget(value?: number): number;
    velocityDecay(value?: number): number;
    nodes(nodes: NodeDatum[]): this;
    force(name: string, force?: Force<NodeDatum>): this;
    find(x: number, y: number, z?: number, radius?: number): NodeDatum | undefined;
    on(type: string, listener?: (event: any) => void): this;
    restart(): this;
    stop(): this;
    tick(): this;
  }

  export interface SimulationNodeDatum {
    index?: number;
    x?: number;
    y?: number;
    z?: number;
    vx?: number;
    vy?: number;
    vz?: number;
    fx?: number | null;
    fy?: number | null;
    fz?: number | null;
  }

  export interface SimulationLinkDatum<NodeDatum extends SimulationNodeDatum> {
    source: NodeDatum | string | number;
    target: NodeDatum | string | number;
  }

  export interface Force<NodeDatum extends SimulationNodeDatum> {
    (alpha: number): void;
    initialize?(nodes: NodeDatum[]): void;
  }

  export interface ForceCenter<NodeDatum extends SimulationNodeDatum> extends Force<NodeDatum> {
    x(value?: number): number;
    y(value?: number): number;
    z(value?: number): number;
  }

  export interface ForceLink<NodeDatum extends SimulationNodeDatum, LinkDatum extends SimulationLinkDatum<NodeDatum>> extends Force<NodeDatum> {
    id(callback: (d: NodeDatum) => string | number): this;
    distance(distance: number | ((d: LinkDatum) => number)): this;
    strength(strength: number | ((d: LinkDatum) => number)): this;
    links(links?: LinkDatum[]): LinkDatum[];
  }

  export interface ForceManyBody<NodeDatum extends SimulationNodeDatum> extends Force<NodeDatum> {
    strength(value: number): this;
    distanceMin(value: number): this;
    distanceMax(value: number): this;
  }

  export function forceSimulation<NodeDatum extends SimulationNodeDatum>(nodes?: NodeDatum[]): Simulation<NodeDatum, any>;
  export function forceCenter<NodeDatum extends SimulationNodeDatum>(x?: number, y?: number, z?: number): ForceCenter<NodeDatum>;
  export function forceLink<NodeDatum extends SimulationNodeDatum, LinkDatum extends SimulationLinkDatum<NodeDatum>>(links?: LinkDatum[]): ForceLink<NodeDatum, LinkDatum>;
  export function forceManyBody<NodeDatum extends SimulationNodeDatum>(): ForceManyBody<NodeDatum>;
}
