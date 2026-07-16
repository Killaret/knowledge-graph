import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import ModalTestWrapper from './ModalTestWrapper.svelte';

vi.mock('$app/environment', () => ({
  browser: true
}));

describe('Modal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders when open is true', () => {
    render(ModalTestWrapper, {
      props: {
        open: true,
        title: 'Test Modal',
        content: 'Modal content'
      }
    });

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Test Modal')).toBeInTheDocument();
    expect(screen.getByText('Modal content')).toBeInTheDocument();
  });

  it('does not render when open is false', () => {
    const { container } = render(ModalTestWrapper, {
      props: {
        open: false,
        title: 'Hidden Modal',
        content: 'Hidden content'
      }
    });

    expect(container.querySelector('[role="dialog"]')).not.toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn();
    render(ModalTestWrapper, {
      props: {
        open: true,
        title: 'Test Modal',
        onClose,
        content: 'Modal content'
      }
    });

    const closeButton = screen.getByLabelText('Close');
    await fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when overlay is clicked', async () => {
    const onClose = vi.fn();
    render(ModalTestWrapper, {
      props: {
        open: true,
        title: 'Test Modal',
        onClose,
        content: 'Modal content'
      }
    });

    const overlay = screen.getByRole('presentation');
    await fireEvent.click(overlay);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('has correct ARIA attributes', () => {
    render(ModalTestWrapper, {
      props: {
        open: true,
        title: 'Accessible Modal',
        content: 'Content'
      }
    });

    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-labelledby', 'modal-title');
  });

  it('cleans up keydown listener on unmount', () => {
    const addEventListenerSpy = vi.spyOn(document, 'addEventListener');
    const removeEventListenerSpy = vi.spyOn(document, 'removeEventListener');

    render(ModalTestWrapper, {
      props: {
        open: true,
        title: 'Test Modal',
        content: 'Content'
      }
    });

    expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));

    cleanup();

    expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
  });
});
