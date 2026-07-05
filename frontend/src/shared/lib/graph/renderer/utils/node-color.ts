/**
 * Node color utility
 */

/**
 * Get color for a node type
 */
export function getNodeColor(type: string | undefined): string {
  const colors: Record<string, string> = {
    star: '#ffcc00',
    planet: '#d6aa5d',
    comet: '#e879f9',
    galaxy: '#8b5cf6',
    asteroid: '#94a3b8',
    blackhole: '#000000',
    moon: '#cccccc',
    nebula: '#2dd4bf',
    dust: '#a1a1aa',
    inbox: '#fbbf24',
    unknown: '#94a3b8'
  };
  return colors[type || 'unknown'] || colors.unknown;
}
