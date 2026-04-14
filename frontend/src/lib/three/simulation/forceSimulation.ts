import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';
import type { GraphData } from '$lib/api/graph';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function createSimulation(
  data: GraphData,
  objectManager: ObjectManager
) {
  // Подготовка данных: добавляем координаты, если их нет
  const nodes = data.nodes.map((n) => ({
    ...n,
    x: (n as any).x ?? (Math.random() - 0.5) * 30,
    y: (n as any).y ?? (Math.random() - 0.5) * 30,
    z: (n as any).z ?? (Math.random() - 0.5) * 30,
  }));

  const links = data.links.map((l) => ({
    ...l,
    source: l.source,
    target: l.target,
    value: l.weight ?? 1,
  }));

  // Создаём визуальные объекты через objectManager
  objectManager.createAll(nodes, links);

  const simulation = forceSimulation(nodes)
    .force(
      'link',
      forceLink(links)
        .id((d: any) => d.id)
        .distance((l: any) => 20 / (l.value * 0.8))
        .strength(0.5)
    )
    .force('charge', forceManyBody().strength(-200).distanceMax(50))
    .force('center', forceCenter(0, 0, 0))
    .alphaDecay(0.02) // Добавляем затухание чтобы симуляция остановилась естественно
    .on('tick', () => {
      objectManager.updatePositions(nodes);
      objectManager.updateLinks(links);
    });

  return simulation;
}
