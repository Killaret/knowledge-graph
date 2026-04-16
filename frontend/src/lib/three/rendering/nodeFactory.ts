import * as THREE from 'three';
import type { GraphNode } from '$lib/api/graph';

function createStarSprite(color: number): THREE.Sprite {
  const canvas = document.createElement('canvas');
  canvas.width = 64;
  canvas.height = 64;
  const ctx = canvas.getContext('2d')!;

  const gradient = ctx.createRadialGradient(32, 32, 0, 32, 32, 32);
  gradient.addColorStop(0, `rgba(255,255,255,1)`);
  gradient.addColorStop(0.4, `#${color.toString(16).padStart(6, '0')}`);
  gradient.addColorStop(0.6, `rgba(${(color >> 16) & 255}, ${(color >> 8) & 255}, ${color & 255}, 0.5)`);
  gradient.addColorStop(1, 'rgba(0,0,0,0)');

  ctx.fillStyle = gradient;
  ctx.fillRect(0, 0, 64, 64);

  const texture = new THREE.CanvasTexture(canvas);
  const material = new THREE.SpriteMaterial({ map: texture, blending: THREE.AdditiveBlending });
  return new THREE.Sprite(material);
}

export function createNodeMesh(node: GraphNode): THREE.Group {
  const group = new THREE.Group();
  const size = getNodeSize(node.type);
  const color = getNodeColor(node.type);

  switch (node.type) {
    case 'star': {
      const starMat = new THREE.MeshStandardMaterial({ color, emissive: color, emissiveIntensity: 0.8 });
      const starMesh = new THREE.Mesh(new THREE.SphereGeometry(size, 32, 16), starMat);
      group.add(starMesh);

      const sprite = createStarSprite(color);
      sprite.scale.set(size * 4, size * 4, 1);
      group.add(sprite);

      for (let i = 0; i < 4; i++) {
        const cone = new THREE.Mesh(
          new THREE.ConeGeometry(size * 0.3, size * 1.5, 8),
          new THREE.MeshStandardMaterial({ color, emissive: color })
        );
        cone.position.y = size * 1.2;
        cone.rotation.z = (i * Math.PI) / 2;
        cone.rotation.x = Math.PI / 2;
        group.add(cone);
      }
      break;
    }
    case 'planet': {
      const planetMat = new THREE.MeshStandardMaterial({ color, roughness: 0.7, metalness: 0.1 });
      const planetMesh = new THREE.Mesh(new THREE.SphereGeometry(size, 32, 16), planetMat);
      group.add(planetMesh);

      const ringGeo = new THREE.TorusGeometry(size * 1.6, size * 0.2, 16, 64);
      const ringMat = new THREE.MeshStandardMaterial({ color: 0xccccaa, side: THREE.DoubleSide });
      const ring = new THREE.Mesh(ringGeo, ringMat);
      ring.rotation.x = Math.PI / 2;
      ring.rotation.z = 0.3;
      group.add(ring);
      break;
    }
    case 'comet': {
      const coreMat = new THREE.MeshStandardMaterial({ color, emissive: color, emissiveIntensity: 0.5 });
      const core = new THREE.Mesh(new THREE.SphereGeometry(size * 0.8, 16, 16), coreMat);
      group.add(core);

      const tailMat = new THREE.MeshStandardMaterial({ color, transparent: true, opacity: 0.6 });
      const tail = new THREE.Mesh(new THREE.ConeGeometry(size * 0.8, size * 3, 8), tailMat);
      tail.position.y = -size * 1.5;
      tail.rotation.x = Math.PI;
      group.add(tail);
      break;
    }
    case 'galaxy': {
      const coreMat = new THREE.MeshStandardMaterial({ color: 0xffdd88, emissive: 0xff8800, emissiveIntensity: 0.8 });
      const core = new THREE.Mesh(new THREE.SphereGeometry(size * 0.6, 16, 16), coreMat);
      group.add(core);

      const particlesGeo = new THREE.BufferGeometry();
      const count = 800;
      const positions = new Float32Array(count * 3);
      const colors = new Float32Array(count * 3);
      for (let i = 0; i < count; i++) {
        const r = (Math.random() * 2 - 1) * size * 3;
        const angle = Math.random() * Math.PI * 2;
        const x = Math.cos(angle) * r;
        const z = Math.sin(angle) * r;
        const y = (Math.random() - 0.5) * size * 0.5;
        positions[i * 3] = x;
        positions[i * 3 + 1] = y;
        positions[i * 3 + 2] = z;

        const col = new THREE.Color().lerpColors(new THREE.Color(0x88aaff), new THREE.Color(0xffdd88), Math.random());
        colors[i * 3] = col.r;
        colors[i * 3 + 1] = col.g;
        colors[i * 3 + 2] = col.b;
      }
      particlesGeo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
      particlesGeo.setAttribute('color', new THREE.BufferAttribute(colors, 3));
      const particlesMat = new THREE.PointsMaterial({ size: 0.1, vertexColors: true, blending: THREE.AdditiveBlending });
      const particles = new THREE.Points(particlesGeo, particlesMat);
      group.add(particles);
      break;
    }
    case 'asteroid': {
      // Irregular rocky shape - dodecahedron with noise
      const asteroidMat = new THREE.MeshStandardMaterial({ color, roughness: 0.9, metalness: 0.2 });
      const geometry = new THREE.DodecahedronGeometry(size, 0);
      const positionAttribute = geometry.attributes.position;
      for (let i = 0; i < positionAttribute.count; i++) {
        const vertex = new THREE.Vector3();
        vertex.fromBufferAttribute(positionAttribute, i);
        // Add noise to create irregular shape
        const noise = 0.7 + Math.random() * 0.6;
        vertex.multiplyScalar(noise);
        positionAttribute.setXYZ(i, vertex.x, vertex.y, vertex.z);
      }
      geometry.computeVertexNormals();
      const asteroidMesh = new THREE.Mesh(geometry, asteroidMat);
      group.add(asteroidMesh);
      break;
    }
    case 'debris': {
      // Scattered small rocks/particles
      const debrisMat = new THREE.MeshStandardMaterial({ color: 0x999999, roughness: 0.8, metalness: 0.1 });
      const particleCount = 8;
      for (let i = 0; i < particleCount; i++) {
        const particleSize = size * (0.3 + Math.random() * 0.4);
        const particle = new THREE.Mesh(new THREE.TetrahedronGeometry(particleSize, 0), debrisMat);
        // Random position around center
        particle.position.x = (Math.random() - 0.5) * size * 2;
        particle.position.y = (Math.random() - 0.5) * size * 2;
        particle.position.z = (Math.random() - 0.5) * size * 2;
        particle.rotation.x = Math.random() * Math.PI;
        particle.rotation.y = Math.random() * Math.PI;
        group.add(particle);
      }
      break;
    }
    default: {
      const mat = new THREE.MeshStandardMaterial({ color });
      const mesh = new THREE.Mesh(new THREE.SphereGeometry(size, 16, 16), mat);
      group.add(mesh);
    }
  }
  return group;
}

function getNodeSize(type?: string): number {
  // Increased sizes for better visibility (were: 1.4, 1.0, 0.9, 1.8, 0.8, 1.2)
  const sizes: Record<string, number> = { star: 2.5, planet: 2.0, comet: 1.8, galaxy: 3.0, asteroid: 1.5, debris: 2.0 };
  return sizes[type || ''] || 2.0;
}

function getNodeColor(type?: string): number {
  const colors: Record<string, number> = { star: 0xffdd44, planet: 0x44aaff, comet: 0xaa88ff, galaxy: 0xff88cc, asteroid: 0x8b7355, debris: 0x999999 };
  return colors[type || ''] || 0x88aaff;
}
