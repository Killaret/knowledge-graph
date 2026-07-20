export function filterValidLinks<
  Node extends { id: string },
  Link extends {
    source: string | number | { id?: string };
    target: string | number | { id?: string };
  },
>(nodes: Node[], links: Link[]): Link[] {
  const nodeIds = new Set(nodes.map((n) => n.id));

  function endpointId(
    value: string | number | { id?: string },
  ): string | undefined {
    if (typeof value === "string") return value;
    if (typeof value === "number") return nodes[value]?.id;
    return value.id;
  }

  return links.filter((l) => {
    const sourceId = endpointId(l.source);
    const targetId = endpointId(l.target);
    if (!sourceId || !targetId) return false;
    return nodeIds.has(sourceId) && nodeIds.has(targetId);
  });
}
