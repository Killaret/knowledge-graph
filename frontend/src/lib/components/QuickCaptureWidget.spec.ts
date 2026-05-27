import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import QuickCaptureWidget from './QuickCaptureWidget.svelte';

// Mock the notes API
vi.mock('$lib/api/notes', () => ({
  createNote: vi.fn()
}));

describe('QuickCaptureWidget', () => {
  it('renders floating button when closed', () => {
    const { container } = render(QuickCaptureWidget);
    
    const button = container.querySelector('.quick-capture-btn');
    expect(button).toBeInTheDocument();
    expect(button?.textContent).toContain('✨');
  });

  it('opens modal when button is clicked', async () => {
    const { container } = render(QuickCaptureWidget);
    
    const button = container.querySelector('.quick-capture-btn');
    if (button) {
      await fireEvent.click(button);
    }
    
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).toBeInTheDocument();
  });

  it('has correct keyboard shortcut Ctrl+Shift+N', () => {
    const { container } = render(QuickCaptureWidget);
    
    // Simulate Ctrl+Shift+N keypress
    const event = new KeyboardEvent('keydown', {
      key: 'n',
      ctrlKey: true,
      shiftKey: true
    });
    
    window.dispatchEvent(event);
    
    // Modal should open
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).toBeInTheDocument();
  });

  it('closes modal with Escape key', async () => {
    const { container } = render(QuickCaptureWidget);
    
    // Open modal first
    const button = container.querySelector('.quick-capture-btn');
    if (button) {
      await fireEvent.click(button);
    }
    
    // Press Escape
    const event = new KeyboardEvent('keydown', {
      key: 'Escape'
    });
    
    window.dispatchEvent(event);
    
    // Modal should close
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).not.toBeInTheDocument();
  });
});
