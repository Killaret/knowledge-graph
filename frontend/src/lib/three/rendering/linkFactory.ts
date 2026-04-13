import * as THREE from 'three';

export function createLinkLine(
  sourcePos: THREE.Vector3,
  targetPos: THREE.Vector3,
  weight: number,
  linkType?: string
): THREE.Line {
  const points = [sourcePos.clone(), targetPos.clone()];
  const geometry = new THREE.BufferGeometry().setFromPoints(points);

  // Цвета для разных типов связей
  const typeColors: Record<string, number> = {
    'reference': 0x3366ff,   // Синий - ссылочная связь
    'dependency': 0xff6600,  // Оранжевый - зависимость
    'related': 0x66ff66,     // Зеленый - связанная тема
    'custom': 0xff66ff,      // Розовый - пользовательская
    'default': 0x88aaff,    // Голубой по умолчанию
  };

  // Выбор цвета на основе типа связи или градиент по весу
  let color: THREE.Color;
  if (linkType && typeColors[linkType]) {
    color = new THREE.Color(typeColors[linkType]);
  } else {
    // Градиент от холодного к тёплому по весу
    const coldColor = new THREE.Color(0x3366ff);
    const warmColor = new THREE.Color(0xffaa00);
    color = coldColor.clone().lerp(warmColor, weight);
  }

  const lineWidth = 1 + (weight ?? 0.5) * 4; // 1..5

  // Стиль линии в зависимости от типа связи
  let material: THREE.LineBasicMaterial | THREE.LineDashedMaterial;
  
  switch (linkType) {
    case 'reference':
      // Сплошная тонкая линия для ссылок
      material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth * 0.8 });
      break;
    case 'dependency':
      // Жирная сплошная для зависимостей
      material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth * 1.5 });
      break;
    case 'related':
      // Пунктирная для связанных тем
      material = new THREE.LineDashedMaterial({
        color,
        dashSize: 0.3,
        gapSize: 0.2,
        linewidth: lineWidth,
      });
      break;
    case 'custom':
      // Точечная линия для пользовательских
      material = new THREE.LineDashedMaterial({
        color,
        dashSize: 0.1,
        gapSize: 0.3,
        linewidth: lineWidth,
      });
      break;
    default:
      // По умолчанию - по весу
      if ((weight ?? 0.5) < 0.3) {
        material = new THREE.LineDashedMaterial({
          color,
          dashSize: 0.2,
          gapSize: 0.1,
          linewidth: lineWidth,
        });
      } else {
        material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth });
      }
  }

  const line = new THREE.Line(geometry, material);
  if (material instanceof THREE.LineDashedMaterial) {
    line.computeLineDistances();
  }
  
  // Сохраняем тип связи в userData для отладки
  line.userData.linkType = linkType || 'default';
  
  return line;
}
