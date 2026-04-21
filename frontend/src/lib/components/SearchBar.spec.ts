import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render as svelteRender, fireEvent, screen, waitFor } from '@testing-library/svelte';
import { goto } from '$app/navigation';
import SearchBar from './SearchBar.svelte';

// Мокаем goto
vi.mock('$app/navigation', () => ({
	goto: vi.fn()
}));

describe('SearchBar', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('renders with default placeholder', () => {
		svelteRender(SearchBar);
		const input = screen.getByPlaceholderText('Search notes (Russian & English)...');
		expect(input).toBeInTheDocument();
	});

	it('renders with custom placeholder', () => {
		svelteRender(SearchBar, { placeholder: 'Custom placeholder' });
		const input = screen.getByPlaceholderText('Custom placeholder');
		expect(input).toBeInTheDocument();
	});

	it('has search button', () => {
		svelteRender(SearchBar);
		const button = screen.getByRole('button', { name: /search/i });
		expect(button).toBeInTheDocument();
	});

	it('navigates to search page on button click', async () => {
		svelteRender(SearchBar);
		const input = screen.getByPlaceholderText('Search notes (Russian & English)...') as HTMLInputElement;
		const button = screen.getByRole('button', { name: /search/i });

		// Устанавливаем значение через свойство input
		input.value = 'test query';
		await fireEvent.input(input);
		await fireEvent.click(button);

		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=test%20query');
		});
	});

	it('navigates with empty query when search is empty', async () => {
		svelteRender(SearchBar);
		const button = screen.getByRole('button', { name: /search/i });

		await fireEvent.click(button);

		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search');
		});
	});

	it('debounces input and navigates after timeout', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		svelteRender(SearchBar);
		const input = screen.getByPlaceholderText('Search notes (Russian & English)...') as HTMLInputElement;

		input.value = 'debounced';
		await fireEvent.input(input);

		// Сразу после ввода не должно быть перехода
		expect(goto).not.toHaveBeenCalled();

		// Ждем debounce delay
		await vi.advanceTimersByTimeAsync(500);

		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=debounced');
		});

		vi.useRealTimers();
	}, 10000);

	it('clears debounce timer on Enter key', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		svelteRender(SearchBar);
		const input = screen.getByPlaceholderText('Search notes (Russian & English)...') as HTMLInputElement;

		input.value = 'enter test';
		await fireEvent.input(input);
		await fireEvent.keyDown(input, { key: 'Enter' });

		// Должен сразу перейти без ожидания debounce
		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=enter%20test');
		});

		// Продвигаем таймер - не должно быть второго вызова
		await vi.advanceTimersByTimeAsync(500);
		expect(goto).toHaveBeenCalledTimes(1);

		vi.useRealTimers();
	});

	it('input has correct aria-label', () => {
		svelteRender(SearchBar);
		// Используем getByRole для input[type=search]
		const input = screen.getByRole('searchbox');
		expect(input).toBeInTheDocument();
		expect(input).toHaveAttribute('aria-label', 'Search');
	});
});
