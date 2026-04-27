/**
 * Canvas rendering functions for GraphCanvas
 */
import { graphConfig2D } from '$lib/config';
import type { SimulationNode, SimulationLink, TransformState } from './types';

export type { SimulationNode, SimulationLink };

// Colors for different link types
const linkTypeColors: Record<string, string> = {
  reference: '#3366ff', // Blue - reference link
  dependency: '#ff6600', // Orange - dependency
  related: '#999999', // Gray - related topic (default)
  custom: '#ff66ff' // Pink - custom
};

/**
 * Get link color based on weight and type
 */
export function getLinkColor(weight: number, linkType?: string): string {
  const effectiveType = linkType || 'related';
  const color = linkTypeColors[effectiveType] || linkTypeColors['related'];
  const opacity = 0.4 + (weight ?? 0.5) * 0.4;

  // Convert hex to rgba
  const r = parseInt(color.slice(1, 3), 16);
  const g = parseInt(color.slice(3, 5), 16);
  const b = parseInt(color.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
}

/**
 * Get line dash pattern based on link type and weight
 */
export function getLineDash(linkType?: string, weight?: number): number[] {
  const effectiveType = linkType || 'related';

  switch (effectiveType) {
    case 'reference':
      return []; // Solid
    case 'dependency':
      return [10, 3]; // Dash-dot
    case 'related':
      // Dash only for weak weight
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    case 'custom':
      return [2, 6]; // Dotted
    default:
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
  }
}

/**
 * Draw a star node
 */
export function drawStar(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number
): void {
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
  ctx.fillStyle = '#ffdd88';
  ctx.strokeStyle = '#cc9900';
  ctx.lineWidth = 2;
  ctx.fill();
  ctx.stroke();
}

/**
 * Draw a planet node
 */
export function drawPlanet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  color?: string
): void {
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = color || '#c9b37c';
  ctx.fill();
  for (let i = -r / 2; i <= r / 2; i += r / 4) {
    ctx.beginPath();
    ctx.ellipse(x, y + i, r * 0.8, r * 0.15, angle, 0, 2 * Math.PI);
    ctx.fillStyle = color ? 'rgba(100,100,100,0.3)' : '#a57c2c';
    ctx.fill();
  }
}

/**
 * Draw a comet node
 */
export function drawComet(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number
): void {
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

/**
 * Draw a galaxy node
 */
export function drawGalaxy(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number
): void {
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);
  for (let i = 0; i < 3; i++) {
    ctx.beginPath();
    ctx.ellipse(0, 0, r * (1 - i * 0.2), r * 0.4, 0, 0, 2 * Math.PI);
    ctx.fillStyle = `rgba(200, 180, 255, ${0.3 - i * 0.1})`;
    ctx.fill();
  }
  ctx.restore();
}

/**
 * Draw a nebula node
 */
export function drawNebula(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number
): void {
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(angle);
  // Nebula - more blurred and turquoise
  for (let i = 0; i < 4; i++) {
    ctx.beginPath();
    ctx.ellipse(0, 0, r * (1.2 - i * 0.2), r * 0.5, i * 0.3, 0, 2 * Math.PI);
    ctx.fillStyle = `rgba(100, 220, 220, ${0.25 - i * 0.05})`;
    ctx.fill();
  }
  ctx.restore();
}

/**
 * Draw an asteroid node
 */
export function drawAsteroid(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
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

/**
 * Draw a debris node
 */
export function drawDebris(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
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

/**
 * Draw a black hole node
 */
export function drawBlackhole(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
  // Event horizon (black circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = '#000000';
  ctx.fill();
  // Accretion disk (glowing ring)
  ctx.beginPath();
  ctx.arc(x, y, r * 1.3, 0, 2 * Math.PI);
  ctx.strokeStyle = '#ff6600';
  ctx.lineWidth = 3;
  ctx.stroke();
  // Inner glow
  ctx.beginPath();
  ctx.arc(x, y, r * 0.8, 0, 2 * Math.PI);
  ctx.strokeStyle = '#ff3300';
  ctx.lineWidth = 2;
  ctx.stroke();
}

/**
 * Draw a moon node
 */
export function drawMoon(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
  // Moon body (grey circle)
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = '#cccccc';
  ctx.fill();
  ctx.strokeStyle = '#999999';
  ctx.lineWidth = 1;
  ctx.stroke();
  // Crater
  ctx.beginPath();
  ctx.arc(x - r * 0.3, y - r * 0.2, r * 0.25, 0, 2 * Math.PI);
  ctx.fillStyle = '#aaaaaa';
  ctx.fill();
}

/**
 * Draw an unknown type node
 */
export function drawUnknown(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  _angle: number
): void {
  // Unknown type - question mark in a dashed circle
  ctx.beginPath();
  ctx.arc(x, y, r, 0, 2 * Math.PI);
  ctx.fillStyle = 'rgba(150, 150, 150, 0.3)';
  ctx.fill();
  ctx.strokeStyle = '#888888';
  ctx.lineWidth = 2;
  ctx.setLineDash([5, 3]);
  ctx.stroke();
  ctx.setLineDash([]);

  // Question mark
  ctx.font = `bold ${Math.floor(r * 1.2)}px sans-serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillStyle = '#aaaaaa';
  ctx.fillText('?', x, y);
}

/**
 * Draw a link between two nodes
 */
export function drawLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  sourceNode: SimulationNode,
  targetNode: SimulationNode
): void {
  ctx.beginPath();
  ctx.moveTo(sourceNode.x!, sourceNode.y!);
  ctx.lineTo(targetNode.x!, targetNode.y!);

  const weight = link.weight ?? 0.5;
  const linkType = link.link_type;

  // Line thickness depends on type and weight
  let lineWidth = Math.max(1, weight * 4);
  if (linkType === 'dependency') lineWidth *= 1.5;
  if (linkType === 'reference') lineWidth *= 0.8;

  ctx.lineWidth = lineWidth;
  ctx.strokeStyle = getLinkColor(weight, linkType);

  // Set dash pattern for dashed lines
  const dash = getLineDash(linkType, weight);
  if (dash.length > 0) {
    ctx.setLineDash(dash);
  } else {
    ctx.setLineDash([]);
  }

  ctx.stroke();
  ctx.setLineDash([]);
}

/**
 * Draw all links
 */
export function drawAllLinks(
  ctx: CanvasRenderingContext2D,
  simLinks: SimulationLink[],
  nodes: SimulationNode[]
): void {
  simLinks.forEach((link) => {
    // After simulation source/target become node objects
    const sourceNode =
      typeof link.source === 'object'
        ? (link.source as SimulationNode)
        : nodes.find((n) => String(n.id) === String(link.source));
    const targetNode =
      typeof link.target === 'object'
        ? (link.target as SimulationNode)
        : nodes.find((n) => String(n.id) === String(link.target));

    if (!sourceNode || !targetNode) {
      return;
    }

    drawLink(ctx, link, sourceNode, targetNode);
  });
}

/**
 * Draw a single node based on its type
 */
export function drawNode(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  r: number,
  angle: number,
  enableShadows: boolean
): void {
  const type = node.type || 'unknown';

  switch (type) {
    case 'star':
      if (enableShadows) {
        ctx.shadowBlur = 10;
        ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
      }
      drawStar(ctx, node.x!, node.y!, r, angle);
      break;
    case 'planet':
      drawPlanet(ctx, node.x!, node.y!, r, angle);
      break;
    case 'satellite':
      drawPlanet(ctx, node.x!, node.y!, r * 0.6, angle, '#aaaaaa');
      break;
    case 'comet':
      drawComet(ctx, node.x!, node.y!, r, angle);
      break;
    case 'galaxy':
      drawGalaxy(ctx, node.x!, node.y!, r, angle);
      break;
    case 'nebula':
      drawNebula(ctx, node.x!, node.y!, r * 1.5, angle);
      break;
    case 'asteroid':
      drawAsteroid(ctx, node.x!, node.y!, r, angle);
      break;
    case 'debris':
      drawDebris(ctx, node.x!, node.y!, r, angle);
      break;
    case 'blackhole':
      drawBlackhole(ctx, node.x!, node.y!, r, angle);
      break;
    case 'moon':
      drawMoon(ctx, node.x!, node.y!, r, angle);
      break;
    case 'unknown':
      drawUnknown(ctx, node.x!, node.y!, r, angle);
      break;
    default:
      if (enableShadows) {
        ctx.shadowBlur = 10;
        ctx.shadowColor = 'rgba(255, 200, 100, 0.8)';
      }
      drawStar(ctx, node.x!, node.y!, r, angle);
      break;
  }
  ctx.shadowBlur = 0;
}

/**
 * Draw node title
 */
export function drawNodeTitle(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  r: number
): void {
  ctx.font = `bold ${Math.min(14, r * 0.65)}px sans-serif`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  let title = node.title || 'Untitled';
  if (title.length > 14) title = title.slice(0, 12) + '…';
  ctx.shadowBlur = 2;
  ctx.shadowColor = 'rgba(0,0,0,0.5)';
  ctx.fillStyle = '#ffffff';
  ctx.fillText(title, node.x!, node.y! + r + 6);
  ctx.shadowBlur = 0;
}

/**
 * Draw all nodes
 */
export function drawAllNodes(
  ctx: CanvasRenderingContext2D,
  nodes: SimulationNode[],
  angles: Map<string, number>,
  enableShadows: boolean
): void {
  const r = 24; // radius increased for better readability

  nodes.forEach((node) => {
    const angle = angles.get(node.id) || 0;
    drawNode(ctx, node, r, angle, enableShadows);
    drawNodeTitle(ctx, node, r);
  });
}

/**
 * Main draw function
 */
export function draw(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  simLinks: SimulationLink[],
  nodes: SimulationNode[],
  angles: Map<string, number>,
  transform: { x: number; y: number; k: number }
): void {
  ctx.clearRect(0, 0, width, height);

  ctx.save();
  ctx.translate(transform.x, transform.y);
  ctx.scale(transform.k, transform.k);

  // Draw links
  drawAllLinks(ctx, simLinks, nodes);

  // Draw nodes
  const enableShadows = nodes.length < graphConfig2D.shadows_threshold;
  drawAllNodes(ctx, nodes, angles, enableShadows);

  ctx.restore();
}

/**
 * Reset view to center the graph
 */
export function resetView(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  transform: { x: number; y: number; k: number }
): void {
  if (nodes.length === 0) return;

  // Find graph bounds
  let minX = Infinity,
    minY = Infinity,
    maxX = -Infinity,
    maxY = -Infinity;
  for (const node of nodes) {
    if (node.x! < minX) minX = node.x!;
    if (node.x! > maxX) maxX = node.x!;
    if (node.y! < minY) minY = node.y!;
    if (node.y! > maxY) maxY = node.y!;
  }

  // Add padding
  const padding = 50;
  minX -= padding;
  minY -= padding;
  maxX += padding;
  maxY += padding;

  const graphWidth = maxX - minX;
  const graphHeight = maxY - minY;

  // Compute scale to fit entire graph
  const scaleX = width / graphWidth;
  const scaleY = height / graphHeight;
  transform.k = Math.min(scaleX, scaleY, 1); // Don't zoom beyond 1:1

  // Center
  transform.x = (width - graphWidth * transform.k) / 2 - minX * transform.k;
  transform.y = (height - graphHeight * transform.k) / 2 - minY * transform.k;
}
