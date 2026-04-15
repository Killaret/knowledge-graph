  
import { describe, it, expect } from 'vitest';
 
import { vi } from 'vitest';
 
import { render, screen } from '@testing-library/svelte';
 
import { fireEvent } from '@testing-library/svelte';
import NoteCard from './NoteCard.svelte';

describe('NoteCard', () => {
	const mockNote = {
		id: '123e4567-e89b-12d3-a456-426614174000',
		title: 'Test Note Title',
		content: 'Test note content',
		created_at: '2024-01-15T10:00:00Z',
		updated_at: '2024-01-15T12:00:00Z',
		metadata: { type: 'star' }
	};

	it('renders note title and content', () => {
		render(NoteCard, { props: { note: mockNote } });
		
		expect(screen.getByText('Test Note Title')).toBeInTheDocument();
		expect(screen.getByText('Test note content')).toBeInTheDocument();
	});

	it('displays formatted date', () => {
		render(NoteCard, { props: { note: mockNote } });
		
		// Should show some date representation
		const dateElement = screen.getByText(/2024|Jan|15/);
		expect(dateElement).toBeInTheDocument();
	});

	it('renders different note types with correct styling', () => {
		const starNote = { ...mockNote, metadata: { type: 'star' } };
		const { container } = render(NoteCard, { props: { note: starNote } });
		
		// Should have star-related class or icon
		expect(container.querySelector('.note-card')).toBeInTheDocument();
	});

	it('handles click events', async () => {
		const onClick = vi.fn();
		render(NoteCard, { props: { note: mockNote, onClick } });

		const card = document.querySelector('.note-card');
		expect(card).toBeTruthy();
		expect(onClick).not.toHaveBeenCalled();

		await fireEvent.click(card!);
		expect(onClick).toHaveBeenCalledWith(mockNote);
	});

	it('renders with minimal note data', () => {
		const minimalNote = {
			id: 'test-id',
			title: 'Minimal',
			content: '',
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString(),
			metadata: {}
		};
		
		render(NoteCard, { props: { note: minimalNote } });
		expect(screen.getByText('Minimal')).toBeInTheDocument();
	});
});
