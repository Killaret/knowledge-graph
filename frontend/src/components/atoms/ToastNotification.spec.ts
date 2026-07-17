import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import ToastNotification from './ToastNotification.svelte';

describe('ToastNotification', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
    cleanup();
  });

  it('renders with message and default type', () => {
    render(ToastNotification, {
      props: {
        message: 'Test notification'
      }
    });

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('Test notification')).toBeInTheDocument();
  });

  it('renders with different types and icons', () => {
    const types = [
      { type: 'success', icon: '✅' },
      { type: 'error', icon: '❌' },
      { type: 'info', icon: 'ℹ️' },
      { type: 'warning', icon: '⚠️' }
    ] as const;

    types.forEach(({ type, icon }) => {
      cleanup();
      render(ToastNotification, {
        props: {
          message: `${type} message`,
          type
        }
      });

      const toast = screen.getByRole('alert');
      expect(toast).toHaveClass('toast-notification', `toast-${type}`);
      expect(screen.getByText(icon)).toBeInTheDocument();
    });
  });

  it('uses galactic icons when useGalacticMode is true', () => {
    cleanup();
    render(ToastNotification, {
      props: {
        message: 'Galactic success',
        type: 'success',
        useGalacticMode: true
      }
    });

    expect(screen.getByText('⭐')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn();
    render(ToastNotification, {
      props: {
        message: 'Closeable toast',
        onClose
      }
    });

    const closeButton = screen.getByLabelText('Close notification');
    await fireEvent.click(closeButton);

    vi.advanceTimersByTime(400);
    expect(onClose).toHaveBeenCalled();
  });

  it('auto-closes after duration', () => {
    const onClose = vi.fn();
    render(ToastNotification, {
      props: {
        message: 'Auto close toast',
        duration: 1000,
        onClose
      }
    });

    expect(screen.getByRole('alert')).toBeInTheDocument();

    vi.advanceTimersByTime(1300);
    expect(onClose).toHaveBeenCalled();
  });
});
