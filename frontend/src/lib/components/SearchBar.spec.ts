import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { tick } from 'svelte';
import { goto } from '$app/navigation';
import SearchBar from './SearchBar.svelte';

// Мокаем goto
vi.mock('$app/navigation', () => ({
	goto: vi.fn(),
}));

describe('SearchBar', () => {
	const user = userEvent.setup();

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
		
		const input = screen.getByPlaceholderText(/search notes/i);
		const button = screen.getByRole('button', { name: /search/i });

		// Используем userEvent для ввода + tick для Svelte 5
		await user.type(input, 'test query');
		await tick();
		await user.click(button);

		expect(goto).toHaveBeenCalledWith('/search?q=test%20query');
	});

	it('navigates on Enter key press', async () => {
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i);
		await user.type(input, 'enter query');
		await tick();
		await user.keyboard('{Enter}');

		expect(goto).toHaveBeenCalledWith('/search?q=enter%20query');
	});

	it('debounces input and navigates after timeout', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i);
		
		// Вводим символы по одному, чтобы симулировать реальный ввод
		await user.type(input, 'd');
		await user.type(input, 'e');
		await user.type(input, 'b');
		await user.type(input, 'o');
		await user.type(input, 'u');
		await user.type(input, 'n');
		await user.type(input, 'c');
		await user.type(input, 'e');

		// Сразу после ввода goto не должен вызываться
		expect(goto).not.toHaveBeenCalled();

		// Продвигаем таймер на время дебаунса (500ms)
		vi.advanceTimersByTime(500);

		// Теперь должен быть вызван
		expect(goto).toHaveBeenCalledTimes(1);
		expect(goto).toHaveBeenCalledWith('/search?q=debounce');

		vi.useRealTimers();
	}, 10000);

	it('clears debounce timer on Enter key', async () => {
		vi.useFakeTimers({ shouldAdvanceTime: true });
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i);
		
		// Вводим что-то
		await user.type(input, 'partial');
		await tick();
		
		// Не дожидаясь дебаунса, нажимаем Enter
		await user.keyboard('{Enter}');
		
		// Должен немедленно перейти
		expect(goto).toHaveBeenCalledWith('/search?q=partial');
		
		// Продвигаем таймер – повторного вызова быть не должно
		vi.advanceTimersByTime(500);
		expect(goto).toHaveBeenCalledTimes(1);

		vi.useRealTimers();
	});

	it('does not navigate when query is empty', async () => {
		render(SearchBar);
		
		const button = screen.getByRole('button', { name: /search/i });
		await user.click(button);
		
		expect(goto).not.toHaveBeenCalled();
		
		// Enter с пустым полем тоже не должен переходить
		const input = screen.getByPlaceholderText(/search notes/i);
		await user.type(input, '   '); // только пробелы
		await user.keyboard('{Enter}');
		
		expect(goto).not.toHaveBeenCalled();
	});

	it('trims whitespace from query', async () => {
		render(SearchBar);
		
		const input = screen.getByPlaceholderText(/search notes/i);
		await user.type(input, '  trimmed  query  ');
		await tick();
		await user.keyboard('{Enter}');
		
		expect(goto).toHaveBeenCalledWith('/search?q=trimmed%20%20query');
	});

	it('has proper ARIA attributes for accessibility', () => {
		render(SearchBar);
		const input = screen.getByPlaceholderText(/search notes/i);
		const button = screen.getByRole('button', { name: /search/i });
		
		expect(input).toHaveAttribute('aria-label', 'Search notes (Russian & English)...');
		expect(button).toHaveAttribute('aria-label', 'Search');
	});
});
