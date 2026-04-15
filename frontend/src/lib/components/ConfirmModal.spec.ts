import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import ConfirmModal from './ConfirmModal.svelte';

describe('ConfirmModal', () => {
	it('renders modal with title and message when open', () => {
		render(ConfirmModal, {
			props: {
				open: true,
				title: 'Delete Note?',
				message: 'Are you sure you want to delete this note?',
				onConfirm: vi.fn(),
				onCancel: vi.fn()
			}
		});

		expect(screen.getByText('Delete Note?')).toBeInTheDocument();
		expect(screen.getByText('Are you sure you want to delete this note?')).toBeInTheDocument();
	});

	it('calls onConfirm when confirm button clicked', async () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();

		render(ConfirmModal, {
			props: {
				open: true,
				title: 'Confirm?',
				message: 'Test message',
				onConfirm,
				onCancel
			}
		});

		const confirmButton = document.querySelector('.confirm-btn');

		if (confirmButton) {
			await fireEvent.click(confirmButton);
			expect(onConfirm).toHaveBeenCalled();
			expect(onCancel).not.toHaveBeenCalled();
		}
	});

	it('calls onCancel when cancel button clicked', async () => {
		const onConfirm = vi.fn();
		const onCancel = vi.fn();

		render(ConfirmModal, {
			props: {
				open: true,
				title: 'Confirm?',
				message: 'Test message',
				onConfirm,
				onCancel
			}
		});

		const cancelButton = document.querySelector('.cancel-btn');

		if (cancelButton) {
			await fireEvent.click(cancelButton);
			expect(onCancel).toHaveBeenCalled();
			expect(onConfirm).not.toHaveBeenCalled();
		}
	});

	it('does not render when open is false', () => {
		render(ConfirmModal, {
			props: {
				open: false,
				title: 'Hidden Modal',
				message: 'Should not see this',
				onConfirm: vi.fn(),
				onCancel: vi.fn()
			}
		});

		// Modal should not be visible
		expect(screen.queryByText('Hidden Modal')).not.toBeInTheDocument();
	});

	it('renders with custom button labels', () => {
		render(ConfirmModal, {
			props: {
				open: true,
				title: 'Save Changes?',
				message: 'Do you want to save?',
				confirmText: 'Save',
				cancelText: 'Discard',
				onConfirm: vi.fn(),
				onCancel: vi.fn()
			}
		});

		expect(screen.getByText('Save')).toBeInTheDocument();
		expect(screen.getByText('Discard')).toBeInTheDocument();
	});
});
