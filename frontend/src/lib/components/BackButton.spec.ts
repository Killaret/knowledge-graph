import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import BackButton from './BackButton.svelte';

describe('BackButton', () => {
	it('renders back button', () => {
		render(BackButton);
		
		const button = screen.getByRole('button') || screen.getByText(/back|←/i);
		expect(button).toBeInTheDocument();
	});

	it('calls onClick when clicked', async () => {
		const onClick = vi.fn();
		render(BackButton, { props: { onClick } });
		
		const button = screen.getByRole('button') || screen.getByText(/back|←/i).closest('button');
		
		if (button) {
			await fireEvent.click(button);
			expect(onClick).toHaveBeenCalled();
		}
	});

	it('renders with custom label', () => {
		render(BackButton, { props: { label: 'Go Back' } });
		
		expect(screen.getByText('Go Back')).toBeInTheDocument();
	});

	it('is disabled when disabled prop is true', () => {
		render(BackButton, { props: { disabled: true } });
		
		const button = screen.getByRole('button');
		expect(button).toBeDisabled();
	});
});
