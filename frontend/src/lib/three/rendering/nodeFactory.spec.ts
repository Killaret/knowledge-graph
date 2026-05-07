import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as THREE from 'three';
import { createNodeMesh } from './nodeFactory';
import type { GraphNode } from '$lib/api/graph';

describe('nodeFactory - 3D Node Rendering', () => {
  beforeEach(() => {
    // Mock canvas for star sprite creation
    const mockCanvas = {
      getContext: vi.fn(() => ({
        createRadialGradient: vi.fn(() => ({
          addColorStop: vi.fn(),
        })),
        fillRect: vi.fn(),
      })),
    };
    vi.stubGlobal('document', {
      createElement: vi.fn((tag: string) => {
        if (tag === 'canvas') return mockCanvas;
        return {};
      }),
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  describe('Node Type Colors and Materials', () => {
    it('star node uses golden color (#ffdd44) with emissive material', () => {
      const node: GraphNode = { id: '1', title: 'Star', type: 'star' };
      const mesh = createNodeMesh(node);

      // Find the main sphere mesh (first child)
      const sphereMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(sphereMesh).toBeDefined();

      const material = sphereMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0xffdd44);
      expect(material.emissive.getHex()).toBe(0xffdd44);
      expect(material.emissiveIntensity).toBe(0.8);
    });

    it('planet node uses blue color (#44aaff) with standard material', () => {
      const node: GraphNode = { id: '2', title: 'Planet', type: 'planet' };
      const mesh = createNodeMesh(node);

      const sphereMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(sphereMesh).toBeDefined();

      const material = sphereMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0x44aaff);
      expect(material.roughness).toBe(0.7);
      expect(material.metalness).toBe(0.1);
    });

    it('comet node uses purple color (#aa88ff) with emissive material', () => {
      const node: GraphNode = { id: '3', title: 'Comet', type: 'comet' };
      const mesh = createNodeMesh(node);

      // Find the core mesh
      const coreMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(coreMesh).toBeDefined();

      const material = coreMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0xaa88ff);
      expect(material.emissive.getHex()).toBe(0xaa88ff);
    });

    it('galaxy node has core with warm colors', () => {
      const node: GraphNode = { id: '4', title: 'Galaxy', type: 'galaxy' };
      const mesh = createNodeMesh(node);

      const coreMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(coreMesh).toBeDefined();

      const material = coreMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0xffdd88);
      expect(material.emissive.getHex()).toBe(0xff8800);
    });

    it('asteroid node uses brown rocky color (#8b7355)', () => {
      const node: GraphNode = { id: '5', title: 'Asteroid', type: 'asteroid' };
      const mesh = createNodeMesh(node);

      const asteroidMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(asteroidMesh).toBeDefined();

      const material = asteroidMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0x8b7355);
      expect(material.roughness).toBe(0.9);
    });

    it('debris node uses gray color (#999999)', () => {
      const node: GraphNode = { id: '6', title: 'Debris', type: 'debris' };
      const mesh = createNodeMesh(node);

      // Debris has multiple particles
      const particles = mesh.children.filter(c => c instanceof THREE.Mesh);
      expect(particles.length).toBeGreaterThan(0);

      const material = (particles[0] as THREE.Mesh).material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0x999999);
    });

    it('satellite node uses gray color (#aaaaaa)', () => {
      const node: GraphNode = { id: '7', title: 'Satellite', type: 'satellite' };
      const mesh = createNodeMesh(node);

      const bodyMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(bodyMesh).toBeDefined();

      const material = bodyMesh.material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0xaaaaaa);
      expect(material.metalness).toBe(0.8);
    });

    it('nebula node uses purple color (#c084fc)', () => {
      const node: GraphNode = { id: '8', title: 'Nebula', type: 'nebula' };
      const mesh = createNodeMesh(node);

      // Nebula has multiple cloud particles
      const clouds = mesh.children.filter(c => c instanceof THREE.Mesh);
      expect(clouds.length).toBeGreaterThan(0);

      const material = (clouds[0] as THREE.Mesh).material as THREE.MeshStandardMaterial;
      expect(material.color.getHex()).toBe(0xc084fc);
      expect(material.transparent).toBe(true);
      expect(material.opacity).toBe(0.4);
    });
  });

  describe('Node Geometry and Structure', () => {
    it('star node has sphere geometry with rays (cones)', () => {
      const node: GraphNode = { id: '1', title: 'Star', type: 'star' };
      const mesh = createNodeMesh(node);

      // Should have: sphere + sprite + 4 cones
      const meshes = mesh.children.filter(c => c instanceof THREE.Mesh);
      const sprites = mesh.children.filter(c => c instanceof THREE.Sprite);

      expect(meshes.length).toBeGreaterThan(1); // sphere + cones
      expect(sprites.length).toBe(1); // glow sprite
    });

    it('planet node has sphere with ring geometry', () => {
      const node: GraphNode = { id: '2', title: 'Planet', type: 'planet' };
      const mesh = createNodeMesh(node);

      // Should have sphere + ring (torus)
      const meshes = mesh.children.filter(c => c instanceof THREE.Mesh);
      const torus = meshes.find(m => (m.geometry as THREE.TorusGeometry).type === 'TorusGeometry');

      expect(meshes.length).toBe(2); // sphere + ring
      expect(torus).toBeDefined();
    });

    it('comet node has sphere with tail (cone)', () => {
      const node: GraphNode = { id: '3', title: 'Comet', type: 'comet' };
      const mesh = createNodeMesh(node);

      const meshes = mesh.children.filter(c => c instanceof THREE.Mesh);
      const cones = meshes.filter(m => (m.geometry as THREE.ConeGeometry).type === 'ConeGeometry');

      expect(meshes.length).toBe(2); // core + tail
      expect(cones.length).toBe(1);
    });

    it('galaxy node has sphere core with particle system', () => {
      const node: GraphNode = { id: '4', title: 'Galaxy', type: 'galaxy' };
      const mesh = createNodeMesh(node);

      const meshes = mesh.children.filter(c => c instanceof THREE.Mesh);
      const points = mesh.children.filter(c => c instanceof THREE.Points);

      expect(meshes.length).toBe(1); // core sphere
      expect(points.length).toBe(1); // particle system
    });

    it('asteroid node has irregular dodecahedron geometry', () => {
      const node: GraphNode = { id: '5', title: 'Asteroid', type: 'asteroid' };
      const mesh = createNodeMesh(node);

      const asteroidMesh = mesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      expect(asteroidMesh).toBeDefined();

      const geometry = asteroidMesh.geometry as THREE.DodecahedronGeometry;
      expect(geometry.type).toBe('DodecahedronGeometry');
    });

    it('satellite node has box body with solar panels and antenna', () => {
      const node: GraphNode = { id: '7', title: 'Satellite', type: 'satellite' };
      const mesh = createNodeMesh(node);

      const meshes = mesh.children.filter(c => c instanceof THREE.Mesh);
      const boxes = meshes.filter(m => (m.geometry as THREE.BoxGeometry).type === 'BoxGeometry');
      const cylinders = meshes.filter(m => (m.geometry as THREE.CylinderGeometry).type === 'CylinderGeometry');

      expect(boxes.length).toBe(2); // body + solar panel
      expect(cylinders.length).toBe(1); // antenna
    });
  });

  describe('Node Sizes', () => {
    it('star has larger size than planet', () => {
      const starNode: GraphNode = { id: '1', title: 'Star', type: 'star' };
      const planetNode: GraphNode = { id: '2', title: 'Planet', type: 'planet' };

      const starMesh = createNodeMesh(starNode);
      const planetMesh = createNodeMesh(planetNode);

      const starSphere = starMesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;
      const planetSphere = planetMesh.children.find(c => c instanceof THREE.Mesh) as THREE.Mesh;

      const starGeometry = starSphere.geometry as THREE.SphereGeometry;
      const planetGeometry = planetSphere.geometry as THREE.SphereGeometry;

      // Star radius (2.5) > Planet radius (2.0)
      expect(starGeometry.parameters.radius).toBeGreaterThan(planetGeometry.parameters.radius);
    });

    it('galaxy has correct size configuration', () => {
      const galaxyNode: GraphNode = { id: '4', title: 'Galaxy', type: 'galaxy' };
      const galaxyMesh = createNodeMesh(galaxyNode);

      // Galaxy has core sphere + particle system
      const meshes = galaxyMesh.children.filter(c => c instanceof THREE.Mesh);
      const points = galaxyMesh.children.filter(c => c instanceof THREE.Points);

      expect(meshes.length).toBe(1); // core sphere
      expect(points.length).toBe(1); // particle system

      // Core should have sphere geometry
      const coreSphere = meshes[0] as THREE.Mesh;
      expect(coreSphere.geometry).toBeInstanceOf(THREE.SphereGeometry);
    });
  });
});
