import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ButtonTestWrapper from './ButtonTestWrapper.svelte';

describe('Button', () => {
  it('renders with default primary variant', () => {
    render(ButtonTestWrapper, {
      props: {
        label: 'Click me'
      }
    });

    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
    expect(button).toHaveClass('button', 'primary');
    expect(button).toHaveAttribute('type', 'button');
  });

  it('renders with different variants', () => {
    const variants = ['primary', 'secondary', 'danger', 'ghost'] as const;
    
    variants.forEach((variant) => {
      const { container } = render(ButtonTestWrapper, {
        props: {
          variant,
          label: `${variant} button`
        }
      });

      const button = container.querySelector('button');
      expect(button).toHaveClass('button', variant);
    });
  });

  it('renders with different button types', () => {
    const { container } = render(ButtonTestWrapper, {
      props: {
        type: 'submit',
        label: 'Submit'
      }
    });

    const button = container.querySelector('button');
    expect(button).toHaveAttribute('type', 'submit');
  });

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(ButtonTestWrapper, {
      props: {
        onClick,
        label: 'Click me'
      }
    });

    const button = screen.getByRole('button', { name: /click me/i });
    await fireEvent.click(button);

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('does not call onClick when disabled', async () => {
    const onClick = vi.fn();
    render(ButtonTestWrapper, {
      props: {
        onClick,
        disabled: true,
        label: 'Click me'
      }
    });

    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeDisabled();
    await fireEvent.click(button);

    expect(onClick).not.toHaveBeenCalled();
  });

  it('spreads additional props to the button element', () => {
    const { container } = render(ButtonTestWrapper, {
      props: {
        dataTestid: 'custom-button',
        ariaLabel: 'Custom button',
        label: 'Custom'
      }
    });

    const button = container.querySelector('button');
    expect(button).toHaveAttribute('data-testid', 'custom-button');
    expect(button).toHaveAttribute('aria-label', 'Custom button');
  });
});
