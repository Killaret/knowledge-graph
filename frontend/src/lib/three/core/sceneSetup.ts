import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';

export interface SceneSetupResult {
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  renderer: THREE.WebGLRenderer;
  labelRenderer: CSS2DRenderer;
  controls: OrbitControls;
}

export function initScene(container: HTMLElement): SceneSetupResult {
  // Сцена
  const scene = new THREE.Scene();
  scene.background = new THREE.Color(0x050510);
  // Initial dense fog for "fog of war" effect during progressive loading
  scene.fog = new THREE.FogExp2(0x050510, 0.08);

  // Камера
  const aspect = container.clientWidth / container.clientHeight;
  const camera = new THREE.PerspectiveCamera(45, aspect, 0.1, 1000);
  camera.position.set(20, 15, 30);
  camera.lookAt(0, 0, 0);

  // Рендереры
  const renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setSize(container.clientWidth, container.clientHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.shadowMap.enabled = true;
  renderer.shadowMap.type = THREE.PCFSoftShadowMap;
  container.appendChild(renderer.domElement);

  const labelRenderer = new CSS2DRenderer();
  labelRenderer.setSize(container.clientWidth, container.clientHeight);
  labelRenderer.domElement.style.position = 'absolute';
  labelRenderer.domElement.style.top = '0';
  labelRenderer.domElement.style.left = '0';
  labelRenderer.domElement.style.pointerEvents = 'none';
  container.appendChild(labelRenderer.domElement);

  // Управление
  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.05;
  controls.autoRotate = true;
  controls.autoRotateSpeed = 0.8;
  controls.enableZoom = true;
  controls.enablePan = true;
  controls.maxPolarAngle = Math.PI / 1.8;

  // Освещение
  const ambientLight = new THREE.AmbientLight(0x404060, 0.6);
  scene.add(ambientLight);

  const dirLight = new THREE.DirectionalLight(0xfff5e6, 1.2);
  dirLight.position.set(10, 30, 10);
  dirLight.castShadow = true;
  dirLight.shadow.mapSize.width = 1024;
  dirLight.shadow.mapSize.height = 1024;
  scene.add(dirLight);

  const fillLight1 = new THREE.PointLight(0x4466ff, 0.8);
  fillLight1.position.set(15, 5, 20);
  scene.add(fillLight1);

  const fillLight2 = new THREE.PointLight(0xff66aa, 0.5);
  fillLight2.position.set(-15, 10, -20);
  scene.add(fillLight2);

  // Звёздное небо
  addStarfield(scene);

  return { scene, camera, renderer, labelRenderer, controls };
}

/**
 * Set the fog density for the scene
 * @param scene - The THREE.Scene instance
 * @param density - Fog density value (0.0 to disable, 0.08 for dense, 0.005 for clear)
 */
export function setFogDensity(scene: THREE.Scene, density: number): void {
  if (scene.fog && scene.fog instanceof THREE.FogExp2) {
    scene.fog.density = density;
  }
}

function addStarfield(scene: THREE.Scene) {
  const geometry = new THREE.BufferGeometry();
  const count = 4000;
  const positions = new Float32Array(count * 3);
  for (let i = 0; i < count * 3; i += 3) {
    const r = 80 + Math.random() * 70;
    const theta = Math.random() * Math.PI * 2;
    const phi = Math.acos(2 * Math.random() - 1);
    positions[i] = Math.sin(phi) * Math.cos(theta) * r;
    positions[i + 1] = Math.sin(phi) * Math.sin(theta) * r;
    positions[i + 2] = Math.cos(phi) * r;
  }
  geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  const material = new THREE.PointsMaterial({
    color: 0xffffff,
    size: 0.15,
    transparent: true,
    blending: THREE.AdditiveBlending,
  });
  const stars = new THREE.Points(geometry, material);
  scene.add(stars);
}
