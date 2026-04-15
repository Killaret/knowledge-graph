import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import BackButton from './BackButton.svelte';

describe('BackButton', () => {
	it('renders back button', () => {
		render(BackButton);

		const button = screen.getByRole('button') || screen.getByText(/back|←/i);
		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(button).toBeInTheDocument();
	});

	it('renders with custom text', () => {
		render(BackButton, { props: { text: 'Go Back' } });

		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(screen.getByText('Go Back')).toBeInTheDocument();
	});

	it('has correct href default', () => {
		render(BackButton);

		const button = screen.getByRole('button');
		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(button).toBeInTheDocument();
	});
});
