/* eslint-disable @typescript-eslint/no-unused-vars */
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
/* eslint-enable @typescript-eslint/no-unused-vars */
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

		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(screen.getByText('Delete Note?')).toBeInTheDocument();
		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
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
			// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(onConfirm).toHaveBeenCalled();
			// eslint-disable-next-line @typescript-eslint/no-unsafe-call
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
			// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(onCancel).toHaveBeenCalled();
			// eslint-disable-next-line @typescript-eslint/no-unsafe-call
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
		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
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

		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(screen.getByText('Save')).toBeInTheDocument();
		// eslint-disable-next-line @typescript-eslint/no-unsafe-call
		expect(screen.getByText('Discard')).toBeInTheDocument();
	});
});
