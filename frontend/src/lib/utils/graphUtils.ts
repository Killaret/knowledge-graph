export function filterValidLinks(nodes: any[], links: any[]): any[] {
  const nodeIds = new Set(nodes.map(n => n.id));
  console.log('[filterValidLinks] nodes count:', nodes.length, 'sample ids:', nodes.slice(0, 3).map(n => n.id));
  console.log('[filterValidLinks] links count:', links.length, 'sample link:', links[0]);

  const valid = links.filter(l => {
    const sourceId = typeof l.source === 'object' ? l.source?.id : l.source;
    const targetId = typeof l.target === 'object' ? l.target?.id : l.target;
    const valid = nodeIds.has(sourceId) && nodeIds.has(targetId);
    if (!valid) {
      console.log('[filterValidLinks] orphan link:', { source: sourceId, target: targetId, hasSource: nodeIds.has(sourceId), hasTarget: nodeIds.has(targetId) });
    }
    return valid;
  });

  console.log('[filterValidLinks] valid links:', valid.length);
  return valid;
}
