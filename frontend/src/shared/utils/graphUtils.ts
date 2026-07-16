export function filterValidLinks(nodes: any[], links: any[]): any[] {
  const nodeIds = new Set(nodes.map(n => n.id));

  return links.filter(l => {
    const sourceId = typeof l.source === 'object' ? l.source?.id : l.source;
    const targetId = typeof l.target === 'object' ? l.target?.id : l.target;
    return nodeIds.has(sourceId) && nodeIds.has(targetId);
  });
}
