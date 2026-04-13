import * as THREE from 'three';

export function createLinkLine(
  sourcePos: THREE.Vector3,
  targetPos: THREE.Vector3,
  weight: number
): THREE.Line {
  const points = [sourcePos.clone(), targetPos.clone()];
  const geometry = new THREE.BufferGeometry().setFromPoints(points);

  // Градиент цвета от холодного к тёплому
  const coldColor = new THREE.Color(0x3366ff);
  const warmColor = new THREE.Color(0xffaa00);
  const color = coldColor.clone().lerp(warmColor, weight);

  const lineWidth = 1 + weight * 4; // 1..5

  let material: THREE.LineBasicMaterial | THREE.LineDashedMaterial;
  if (weight < 0.3) {
    material = new THREE.LineDashedMaterial({
      color,
      dashSize: 0.2,
      gapSize: 0.1,
      linewidth: lineWidth,
    });
  } else {
    material = new THREE.LineBasicMaterial({ color, linewidth: lineWidth });
  }

  const line = new THREE.Line(geometry, material);
  if (material instanceof THREE.LineDashedMaterial) {
    line.computeLineDistances();
  }
  return line;
}
