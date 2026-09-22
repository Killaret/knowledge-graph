import { CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';

export function createLabel(text: string, onClick?: () => void): CSS2DObject {
  const div = document.createElement('div');
  // Сокращаем длинный текст
  div.textContent = text.length > 20 ? text.slice(0, 17) + '...' : text;
  div.style.color = '#ffffff';
  div.style.fontSize = '14px';
  div.style.fontWeight = 'bold';
  div.style.fontFamily = 'Arial, sans-serif';
  div.style.textShadow = '1px 1px 3px rgba(0,0,0,0.8)';
  div.style.padding = '4px 10px';
  div.style.background = 'rgba(20, 30, 50, 0.85)';
  div.style.borderRadius = '20px';
  div.style.backdropFilter = 'blur(4px)';
  div.style.border = '1px solid rgba(255, 255, 255, 0.3)';
  div.style.pointerEvents = 'auto';
  div.style.cursor = 'pointer';
  div.style.transition = 'transform 0.2s, background 0.2s';
  div.style.whiteSpace = 'nowrap';
  
  div.addEventListener('mouseenter', () => {
    div.style.transform = 'scale(1.05)';
    div.style.background = 'rgba(40, 60, 100, 0.9)';
  });
  div.addEventListener('mouseleave', () => {
    div.style.transform = 'scale(1)';
    div.style.background = 'rgba(20, 30, 50, 0.85)';
  });
  
  if (onClick) {
    div.addEventListener('click', (e) => {
      e.stopPropagation();
      onClick();
    });
  }
  
  return new CSS2DObject(div);
}
