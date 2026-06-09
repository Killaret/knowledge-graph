import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import QuickCaptureWidget from './QuickCaptureWidget.svelte';

// Mock the notes API
vi.mock('$lib/api/notes', () => ({
  createNote: vi.fn()
}));

// Reset mocks before each test
beforeEach(() => {
  vi.clearAllMocks();
});

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
    expect(button).toBeInTheDocument();
    
    await fireEvent.mouseDown(button!);
    
    // Wait for modal to render
    await new Promise(resolve => setTimeout(resolve, 10));
    
    // Check that modal is rendered
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).toBeTruthy();
  });

  it('has correct keyboard shortcut Ctrl+Shift+N', async () => {
    const { container } = render(QuickCaptureWidget);
    
    // Press Ctrl+Shift+N
    const event = new KeyboardEvent('keydown', {
      key: 'n',
      ctrlKey: true,
      shiftKey: true,
      bubbles: true,
      cancelable: true
    });
    
    window.dispatchEvent(event);
    
    // Give it a moment to process
    await new Promise(resolve => setTimeout(resolve, 10));
    
    // Modal should be rendered
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).toBeTruthy();
  });

  it('closes modal with Escape key', async () => {
    const { container } = render(QuickCaptureWidget);
    
    // Open modal first
    const button = container.querySelector('.quick-capture-btn');
    if (button) {
      await fireEvent.click(button);
    }
    
    // Give it a moment to open
    await new Promise(resolve => setTimeout(resolve, 10));
    
    // Press Escape
    const event = new KeyboardEvent('keydown', {
      key: 'Escape',
      bubbles: true,
      cancelable: true
    });
    
    window.dispatchEvent(event);
    
    // Give it a moment to process
    await new Promise(resolve => setTimeout(resolve, 10));
    
    // Modal should be removed from DOM
    const modal = container.querySelector('.quick-capture-modal');
    expect(modal).toBeFalsy();
  });
});
