import { CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';

export function createLabel(text: string, onClick?: () => void): CSS2DObject {
  const div = document.createElement('div');
  div.textContent = text;
  div.style.color = '#ffffff';
  div.style.fontSize = '14px';
  div.style.fontWeight = 'bold';
  div.style.fontFamily = 'Arial, sans-serif';
  div.style.textShadow = '0 0 10px rgba(0,0,0,0.8)';
  div.style.padding = '4px 8px';
  div.style.background = 'rgba(20,30,50,0.7)';
  div.style.borderRadius = '12px';
  div.style.backdropFilter = 'blur(4px)';
  div.style.border = '1px solid rgba(255,255,255,0.2)';
  div.style.pointerEvents = 'auto';
  div.style.cursor = 'pointer';
  div.style.transition = 'transform 0.2s';
  div.addEventListener('mouseenter', () => (div.style.transform = 'scale(1.05)'));
  div.addEventListener('mouseleave', () => (div.style.transform = 'scale(1)'));
  if (onClick) {
    div.addEventListener('click', (e) => {
      e.stopPropagation();
      onClick();
    });
  }
  return new CSS2DObject(div);
}
