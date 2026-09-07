/* eslint-disable @typescript-eslint/no-explicit-any */
declare module "d3-force-3d" {
  export interface Simulation<
    NodeDatum extends SimulationNodeDatum,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _LinkDatum extends SimulationLinkDatum<NodeDatum>,
  > {
    alpha(): number;
    alpha(value: number): this;
    alphaMin(): number;
    alphaMin(value: number): this;
    alphaDecay(): number;
    alphaDecay(value: number): this;
    alphaTarget(): number;
    alphaTarget(value: number): this;
    velocityDecay(): number;
    velocityDecay(value: number): this;
    nodes(): NodeDatum[];
    nodes(nodes: NodeDatum[]): this;
    force(name: string): Force<NodeDatum> | undefined;
    force(name: string, force: Force<NodeDatum> | null): this;
    find(x: number, y: number, z?: number, radius?: number): NodeDatum | undefined;
    on(type: string): ((event: any) => void) | undefined;
    on(type: string, listener: ((event: any) => void) | null): this;
    restart(): this;
    stop(): this;
    tick(iterations?: number): this;
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
    x(): number;
    x(value: number): this;
    y(): number;
    y(value: number): this;
    z(): number;
    z(value: number): this;
  }

  export interface ForceLink<
    NodeDatum extends SimulationNodeDatum,
    LinkDatum extends SimulationLinkDatum<NodeDatum>,
  > extends Force<NodeDatum> {
    id(): (d: NodeDatum) => string | number;
    id(callback: (d: NodeDatum) => string | number): this;
    distance(): number;
    distance(distance: number | ((d: LinkDatum) => number)): this;
    strength(): number;
    strength(strength: number | ((d: LinkDatum) => number)): this;
    links(): LinkDatum[];
    links(links: LinkDatum[]): this;
  }

  export interface ForceManyBody<NodeDatum extends SimulationNodeDatum> extends Force<NodeDatum> {
    strength(): number;
    strength(value: number): this;
    distanceMin(): number;
    distanceMin(value: number): this;
    distanceMax(): number;
    distanceMax(value: number): this;
  }

  export function forceSimulation<NodeDatum extends SimulationNodeDatum>(
    nodes?: NodeDatum[]
  ): Simulation<NodeDatum, any>;
  export function forceCenter<NodeDatum extends SimulationNodeDatum>(
    x?: number,
    y?: number,
    z?: number
  ): ForceCenter<NodeDatum>;
  export function forceLink<
    NodeDatum extends SimulationNodeDatum,
    LinkDatum extends SimulationLinkDatum<NodeDatum>,
  >(links?: LinkDatum[]): ForceLink<NodeDatum, LinkDatum>;
  export function forceManyBody<NodeDatum extends SimulationNodeDatum>(): ForceManyBody<NodeDatum>;
}
