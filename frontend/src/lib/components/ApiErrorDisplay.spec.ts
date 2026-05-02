import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ApiErrorDisplay from './ApiErrorDisplay.svelte';
import type { ErrorResponse } from '$lib/types/errors';

describe('ApiErrorDisplay', () => {
  const mockError: ErrorResponse = {
    code: 'VALIDATION_ERROR',
    message: 'Некорректные входные данные'
  };

  const mockErrorWithDetails: ErrorResponse = {
    code: 'VALIDATION_ERROR',
    message: 'Некорректные входные данные',
    details: [
      {
        field: 'title',
        reason: 'required',
        message: 'Поле обязательно для заполнения',
        received: '',
        expected: 'non-empty string'
      },
      {
        field: 'content',
        reason: 'too_short',
        message: 'Слишком короткое содержание'
      }
    ]
  };

  it('renders with error data (code and message visible)', () => {
    render(ApiErrorDisplay, { props: { error: mockError } });

    expect(screen.getByText('Ошибка: VALIDATION_ERROR')).toBeInTheDocument();
    expect(screen.getByText('Некорректные входные данные')).toBeInTheDocument();
    expect(screen.getByText('⚠️')).toBeInTheDocument();
  });

  it('does not render when error is null', () => {
    const { container } = render(ApiErrorDisplay, { props: { error: null } });
    
    // Component should render nothing when error is null
    // Svelte leaves a comment node for empty renders, so check content instead
    expect(container.textContent).toBe('');
    expect(container.querySelector('.error-container')).not.toBeInTheDocument();
  });

  it('displays details list when details are provided', () => {
    render(ApiErrorDisplay, { props: { error: mockErrorWithDetails } });

    expect(screen.getByText('Детали:')).toBeInTheDocument();
    expect(screen.getByText('title')).toBeInTheDocument();
    expect(screen.getByText('Поле обязательно для заполнения')).toBeInTheDocument();
    expect(screen.getByText('content')).toBeInTheDocument();
    expect(screen.getByText('Слишком короткое содержание')).toBeInTheDocument();
  });

  it('displays received value in details when available', () => {
    render(ApiErrorDisplay, { props: { error: mockErrorWithDetails } });

    expect(screen.getByText('(получено: "")')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', async () => {
    const onClose = vi.fn();
    render(ApiErrorDisplay, { props: { error: mockError, onClose } });

    const closeButton = screen.getByLabelText('Закрыть ошибку');
    await fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('has correct ARIA attributes', () => {
    render(ApiErrorDisplay, { props: { error: mockError } });

    const alert = screen.getByRole('alert');
    expect(alert).toHaveAttribute('aria-live', 'assertive');
  });

  it('close button has correct aria-label', () => {
    render(ApiErrorDisplay, { props: { error: mockError } });

    const closeButton = screen.getByLabelText('Закрыть ошибку');
    expect(closeButton).toBeInTheDocument();
    expect(closeButton).toHaveAttribute('type', 'button');
  });

  it('does not show details section when details are empty', () => {
    const errorWithoutDetails: ErrorResponse = {
      code: 'NOT_FOUND',
      message: 'Заметка не найдена',
      details: []
    };

    render(ApiErrorDisplay, { props: { error: errorWithoutDetails } });

    expect(screen.queryByText('Детали:')).not.toBeInTheDocument();
  });

  it('does not show details section when details are undefined', () => {
    render(ApiErrorDisplay, { props: { error: mockError } });

    expect(screen.queryByText('Детали:')).not.toBeInTheDocument();
  });

  it('works without onClose callback', async () => {
    render(ApiErrorDisplay, { props: { error: mockError } });

    const closeButton = screen.getByLabelText('Закрыть ошибку');
    // Should not throw when clicking without onClose
    await fireEvent.click(closeButton);
    
    // Component should still be rendered
    expect(screen.getByText('Ошибка: VALIDATION_ERROR')).toBeInTheDocument();
  });
});
