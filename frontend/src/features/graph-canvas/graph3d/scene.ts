import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { CSS2DRenderer } from "three/examples/jsm/renderers/CSS2DRenderer.js";
import type { Graph3DConfig } from "./types";

export interface SceneBundle {
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  renderer: THREE.WebGLRenderer;
  labelRenderer: CSS2DRenderer;
  controls: OrbitControls;
  dispose: () => void;
}

export function createScene(container: HTMLElement, config: Graph3DConfig): SceneBundle {
  const scene = new THREE.Scene();
  scene.background = new THREE.Color(0x050510);
  scene.fog = new THREE.FogExp2(0x050510, config.fogDensityInitial);

  const aspect = container.clientWidth / container.clientHeight || 1;
  const camera = new THREE.PerspectiveCamera(45, aspect, 0.1, 2000);
  camera.position.set(0, 0, 80);

  const renderer = new THREE.WebGLRenderer({
    antialias: false,
    powerPreference: "high-performance",
    alpha: false,
  });
  renderer.setSize(container.clientWidth, container.clientHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  container.appendChild(renderer.domElement);

  const labelRenderer = new CSS2DRenderer();
  labelRenderer.setSize(container.clientWidth, container.clientHeight);
  const labelDom = labelRenderer.domElement;
  labelDom.style.position = "absolute";
  labelDom.style.top = "0";
  labelDom.style.left = "0";
  labelDom.style.pointerEvents = "none";
  labelDom.style.width = "100%";
  labelDom.style.height = "100%";
  container.appendChild(labelDom);

  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.05;
  controls.autoRotate = config.autoRotate;
  controls.autoRotateSpeed = config.autoRotate ? 0.8 : 0;
  controls.enableZoom = true;
  controls.enablePan = true;
  controls.maxPolarAngle = Math.PI / 1.8;

  scene.add(new THREE.AmbientLight(0x404060, 0.8));

  const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
  dirLight.position.set(20, 30, 20);
  scene.add(dirLight);

  const fillLight = new THREE.PointLight(0x4466ff, 0.4);
  fillLight.position.set(-20, -10, 20);
  scene.add(fillLight);

  const starfield = createStarfield();
  scene.add(starfield);

  return {
    scene,
    camera,
    renderer,
    labelRenderer,
    controls,
    dispose: () => {
      controls.dispose();
      renderer.dispose();
      scene.remove(starfield);
      starfield.geometry.dispose();
      if (Array.isArray(starfield.material)) {
        starfield.material.forEach((m) => m.dispose());
      } else {
        starfield.material.dispose();
      }
      renderer.domElement.remove();
      labelDom.remove();
    },
  };
}

function createStarfield(count = 1500): THREE.Points {
  const geometry = new THREE.BufferGeometry();
  const positions = new Float32Array(count * 3);
  for (let i = 0; i < count; i++) {
    const r = 80 + Math.random() * 120;
    const theta = Math.random() * Math.PI * 2;
    const phi = Math.acos(2 * Math.random() - 1);
    positions[i * 3] = r * Math.sin(phi) * Math.cos(theta);
    positions[i * 3 + 1] = r * Math.sin(phi) * Math.sin(theta);
    positions[i * 3 + 2] = r * Math.cos(phi);
  }
  geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
  const material = new THREE.PointsMaterial({
    color: 0xffffff,
    size: 0.25,
    transparent: true,
    opacity: 0.7,
    sizeAttenuation: true,
  });
  return new THREE.Points(geometry, material);
}

export function setFogDensity(scene: THREE.Scene, density: number): void {
  if (scene.fog && scene.fog instanceof THREE.FogExp2) {
    scene.fog.density = density;
  }
}

export function resizeScene(
  container: HTMLElement,
  camera: THREE.PerspectiveCamera,
  renderer: THREE.WebGLRenderer,
  labelRenderer: CSS2DRenderer
): void {
  const width = container.clientWidth || 1;
  const height = container.clientHeight || 1;
  camera.aspect = width / height;
  camera.updateProjectionMatrix();
  renderer.setSize(width, height);
  labelRenderer.setSize(width, height);
}
