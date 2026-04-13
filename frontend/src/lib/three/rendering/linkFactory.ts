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
    'related': 0x999999,     // Серый - связанная тема (по умолчанию)
    'custom': 0xff66ff,      // Розовый - пользовательская
  };

  const lineWidth = 1 + (weight ?? 0.5) * 4; // 1..5

  // Определяем тип связи (по умолчанию 'related')
  const effectiveType = linkType || 'related';
  
  // Выбор цвета на основе типа связи
  const colorHex = typeColors[effectiveType] || typeColors['related'];
  const color = new THREE.Color(colorHex);

  // Стиль линии в зависимости от типа связи
  let material: THREE.LineBasicMaterial | THREE.LineDashedMaterial;
  
  switch (effectiveType) {
    case 'reference':
      // Сплошная линия для ссылок (синий)
      material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth });
      break;
    case 'dependency':
      // Штрихпунктир для зависимостей (оранжевый)
      material = new THREE.LineDashedMaterial({
        color,
        dashSize: 0.4,
        gapSize: 0.15,
        linewidth: lineWidth * 1.2,
      });
      break;
    case 'related':
      // Серый, пунктир при слабом весе
      if ((weight ?? 0.5) < 0.3) {
        material = new THREE.LineDashedMaterial({
          color,
          dashSize: 0.3,
          gapSize: 0.2,
          linewidth: lineWidth,
        });
      } else {
        material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth });
      }
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
      // По умолчанию - как related
      material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth });
  }

  const line = new THREE.Line(geometry, material);
  if (material instanceof THREE.LineDashedMaterial) {
    line.computeLineDistances();
  }
  
  // Сохраняем тип связи в userData для отладки
  line.userData.linkType = linkType || 'default';
  
  return line;
}
