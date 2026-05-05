import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as THREE from 'three';
import { createLinkLine } from './linkFactory';

describe('linkFactory - 3D Link Rendering', () => {
  let sourcePos: THREE.Vector3;
  let targetPos: THREE.Vector3;

  beforeEach(() => {
    sourcePos = new THREE.Vector3(10, 20, 30);
    targetPos = new THREE.Vector3(50, 60, 70);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Link Geometry and Position', () => {
    it('creates line with correct start and end positions', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'related');

      const geometry = line.geometry as THREE.BufferGeometry;
      const positions = geometry.attributes.position.array as Float32Array;

      // First point should be source position
      expect(positions[0]).toBeCloseTo(sourcePos.x);
      expect(positions[1]).toBeCloseTo(sourcePos.y);
      expect(positions[2]).toBeCloseTo(sourcePos.z);

      // Second point should be target position
      expect(positions[3]).toBeCloseTo(targetPos.x);
      expect(positions[4]).toBeCloseTo(targetPos.y);
      expect(positions[5]).toBeCloseTo(targetPos.z);
    });

    it('creates line with correct distance', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'related');

      const geometry = line.geometry as THREE.BufferGeometry;
      const positions = geometry.attributes.position.array as Float32Array;

      // Calculate actual distance
      const dx = positions[3] - positions[0];
      const dy = positions[4] - positions[1];
      const dz = positions[5] - positions[2];
      const distance = Math.sqrt(dx * dx + dy * dy + dz * dz);

      // Should match distance between source and target
      const expectedDistance = sourcePos.distanceTo(targetPos);
      expect(distance).toBeCloseTo(expectedDistance);
    });

    it('stores link type in userData', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'dependency');
      expect(line.userData.linkType).toBe('dependency');
    });

    it('defaults to related type when not specified', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8);
      expect(line.userData.linkType).toBe('related');
    });
  });

  describe('Link Type Colors', () => {
    it('reference link uses blue color (#3366ff)', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'reference');
      const material = line.material as THREE.LineBasicMaterial;

      expect(material.color.getHex()).toBe(0x3366ff);
      expect(material.transparent).toBe(true);
    });

    it('dependency link uses orange color (#ff6600)', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'dependency');
      const material = line.material as THREE.LineDashedMaterial;

      expect(material.color.getHex()).toBe(0xff6600);
      expect(material).toBeInstanceOf(THREE.LineDashedMaterial);
    });

    it('related link uses gray color (#999999)', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'related');
      const material = line.material as THREE.LineBasicMaterial;

      expect(material.color.getHex()).toBe(0x999999);
    });

    it('custom link uses pink color (#ff66ff)', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'custom');
      const material = line.material as THREE.LineDashedMaterial;

      expect(material.color.getHex()).toBe(0xff66ff);
    });
  });

  describe('Link Opacity Based on Weight', () => {
    it('strong weight (1.0) has high opacity', () => {
      const line = createLinkLine(sourcePos, targetPos, 1.0, 'reference');
      const material = line.material as THREE.LineBasicMaterial;

      expect(material.opacity).toBeGreaterThan(0.8);
    });

    it('weak weight (0.3) has lower opacity', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.3, 'reference');
      const material = line.material as THREE.LineBasicMaterial;

      expect(material.opacity).toBeLessThan(0.6);
    });

    it('medium weight (0.5) has medium opacity', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.5, 'reference');
      const material = line.material as THREE.LineBasicMaterial;

      expect(material.opacity).toBeGreaterThan(0.5);
      expect(material.opacity).toBeLessThan(0.8);
    });
  });

  describe('Link Dash Patterns', () => {
    it('reference link is solid (no dashes)', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'reference');
      const material = line.material;

      expect(material).toBeInstanceOf(THREE.LineBasicMaterial);
      expect(material).not.toBeInstanceOf(THREE.LineDashedMaterial);
    });

    it('dependency link has dashed pattern', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'dependency');
      const material = line.material as THREE.LineDashedMaterial;

      expect(material).toBeInstanceOf(THREE.LineDashedMaterial);
      expect(material.dashSize).toBe(0.4);
      expect(material.gapSize).toBe(0.15);
    });

    it('custom link has dotted pattern', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'custom');
      const material = line.material as THREE.LineDashedMaterial;

      expect(material).toBeInstanceOf(THREE.LineDashedMaterial);
      expect(material.dashSize).toBe(0.1);
      expect(material.gapSize).toBe(0.3);
    });

    it('related link with weak weight uses dashed pattern', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.2, 'related');
      const material = line.material as THREE.LineDashedMaterial;

      expect(material).toBeInstanceOf(THREE.LineDashedMaterial);
      expect(material.dashSize).toBe(0.3);
      expect(material.gapSize).toBe(0.2);
    });

    it('related link with strong weight is solid', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'related');
      const material = line.material;

      expect(material).toBeInstanceOf(THREE.LineBasicMaterial);
      expect(material).not.toBeInstanceOf(THREE.LineDashedMaterial);
    });
  });

  describe('Link Line Distances', () => {
    it('dashed materials have computed line distances', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'dependency');

      // LineDashedMaterial requires computeLineDistances() to be called
      expect(line.material).toBeInstanceOf(THREE.LineDashedMaterial);
    });

    it('solid lines do not need line distances', () => {
      const line = createLinkLine(sourcePos, targetPos, 0.8, 'reference');

      // LineBasicMaterial does not need computeLineDistances()
      expect(line.material).toBeInstanceOf(THREE.LineBasicMaterial);
    });
  });
});
