import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import { goto } from '$app/navigation';
import SearchBar from './SearchBar.svelte';

// Хелпер для установки значения input с fireEvent и tick
async function setInputValue(input: HTMLInputElement, value: string) {
	// Properly dispatch input event for Svelte 5 bind:value
	input.value = value;
	input.dispatchEvent(new InputEvent('input', { bubbles: true }));
	await tick(); // Даём Svelte время обновить состояние
	await new Promise(resolve => setTimeout(resolve, 0));
}

describe('SearchBar', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.clearAllTimers?.();
	});

	it('renders with placeholder', () => {
		render(SearchBar);
		const input = screen.getByPlaceholderText(/search notes/i);
		expect(input).toBeInTheDocument();
	});

	it('navigates to search page on button click', async () => {
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		const button = screen.getByRole('button', { name: /search/i });

		await setInputValue(input, 'test query');
		await fireEvent.click(button);

		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=test%20query');
		});
	});

	it('navigates on Enter key press', async () => {
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		await setInputValue(input, 'enter query');
		await fireEvent.keyDown(input, { key: 'Enter' });

		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=enter%20query');
		});
	});

	it('debounces input and navigates after timeout', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		
		// Вводим слово через хелпер
		await setInputValue(input, 'debounce');

		// Сразу после ввода goto не должен вызываться
		expect(goto).not.toHaveBeenCalled();

		// Продвигаем таймер на время дебаунса (500ms)
		vi.advanceTimersByTime(500);

		// Теперь должен быть вызван
		await waitFor(() => {
			expect(goto).toHaveBeenCalledTimes(1);
			expect(goto).toHaveBeenCalledWith('/search?q=debounce');
		});

		vi.useRealTimers();
	}, 10000);

	it('clears debounce timer on Enter key', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		
		// Вводим что-то
		await setInputValue(input, 'partial');
		
		// Не дожидаясь дебаунса, нажимаем Enter
		await fireEvent.keyDown(input, { key: 'Enter' });
		
		// Должен немедленно перейти
		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=partial');
		});
		
		// Продвигаем таймер – повторного вызова быть не должно
		vi.advanceTimersByTime(500);
		expect(goto).toHaveBeenCalledTimes(1);

		vi.useRealTimers();
	});

	it('does not navigate when query is empty', async () => {
		render(SearchBar);
		
		const button = screen.getByRole('button', { name: /search/i });
		await fireEvent.click(button);
		
		expect(goto).not.toHaveBeenCalled();
		
		// Enter с пустым полем тоже не должен переходить
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		await setInputValue(input, '   '); // только пробелы
		await fireEvent.keyDown(input, { key: 'Enter' });
		
		expect(goto).not.toHaveBeenCalled();
	});

	it('trims whitespace from query', async () => {
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i) as HTMLInputElement;
		await setInputValue(input, '  trimmed  query  ');
		await fireEvent.keyDown(input, { key: 'Enter' });
		
		await waitFor(() => {
			expect(goto).toHaveBeenCalledWith('/search?q=trimmed%20%20query');
		});
	});

	it('input has correct aria-label', () => {
		render(SearchBar);
		const input = screen.getByRole('searchbox');
		expect(input).toBeInTheDocument();
		expect(input).toHaveAttribute('aria-label', 'Search');
	});
});
