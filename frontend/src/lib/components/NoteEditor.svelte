<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { createNote, updateNote, getNote } from '$lib/api/notes';
	import { tick } from 'svelte';

	interface Props {
		noteId?: string | null;
		onCancel?: () => void;
	}

	let { noteId = null, onCancel }: Props = $props();

	let title = $state('');
	let content = $state('');
	let noteType = $state('star');
	let isLoading = $state(false);
	let isSaving = $state(false);
	let error = $state<string | null>(null);
	let titleError = $state<string | null>(null);

	// Загрузка данных при редактировании
	$effect(() => {
		if (noteId) {
			loadNote(noteId);
		}
	});

	async function loadNote(id: string) {
		isLoading = true;
		error = null;
		try {
			const note = await getNote(id);
			title = note.title;
			content = note.content || '';
			noteType = note.type || 'star';
		} catch (e) {
			error = 'Failed to load note';
		} finally {
			isLoading = false;
		}
	}

	function validate(): boolean {
		titleError = null;
		if (!title.trim()) {
			titleError = 'Title is required';
			return false;
		}
		return true;
	}

	async function handleSave() {
		if (!validate()) return;

		isSaving = true;
		error = null;

		try {
			const noteData = {
				title: title.trim(),
				content: content.trim(),
				type: noteType
			};

			if (noteId) {
				await updateNote(noteId, noteData);
			} else {
				const newNote = await createNote(noteData);
				await goto(`/notes/${newNote.id}`);
			}
		} catch (e) {
			error = 'Failed to save note. Please try again.';
		} finally {
			isSaving = false;
		}
	}

	function handleCancel() {
		if (onCancel) {
			onCancel();
		} else {
			goto('/');
		}
	}
</script>

<div class="note-editor" data-testid="note-editor">
	{#if isLoading}
		<div class="loading" data-testid="loading">Loading...</div>
	{:else}
		{#if error}
			<div class="error-message" data-testid="error-message" role="alert">
				{error}
			</div>
		{/if}

		<form onsubmit={(e) => { e.preventDefault(); handleSave(); }}>
			<div class="field">
				<label for="title">Title</label>
				<input
					id="title"
					type="text"
					bind:value={title}
					placeholder="Enter note title"
					data-testid="title-input"
					disabled={isSaving}
				/>
				{#if titleError}
					<span class="field-error" data-testid="title-error">{titleError}</span>
				{/if}
			</div>

			<div class="field">
				<label for="type">Type</label>
				<select id="type" bind:value={noteType} data-testid="type-select" disabled={isSaving}>
					<option value="star">Star</option>
					<option value="planet">Planet</option>
					<option value="comet">Comet</option>
				</select>
			</div>

			<div class="field">
				<label for="content">Content</label>
				<textarea
					id="content"
					bind:value={content}
					placeholder="Enter note content"
					rows="10"
					data-testid="content-input"
					disabled={isSaving}
				></textarea>
			</div>

			<div class="actions">
				<button
					type="submit"
					class="btn-primary"
					disabled={isSaving}
					data-testid="save-button"
				>
					{isSaving ? 'Saving...' : (noteId ? 'Update' : 'Create')}
				</button>
				<button
					type="button"
					class="btn-secondary"
					onclick={handleCancel}
					disabled={isSaving}
					data-testid="cancel-button"
				>
					Cancel
				</button>
			</div>
		</form>
	{/if}
</div>

<style>
	.note-editor {
		max-width: 800px;
		margin: 0 auto;
		padding: 2rem;
	}

	.loading {
		text-align: center;
		padding: 2rem;
		color: #666;
	}

	.error-message {
		background: #fee;
		color: #c33;
		padding: 1rem;
		border-radius: 4px;
		margin-bottom: 1rem;
	}

	.field {
		margin-bottom: 1.5rem;
	}

	label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
	}

	input,
	select,
	textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 1rem;
	}

	input:focus,
	select:focus,
	textarea:focus {
		outline: none;
		border-color: #4a90d9;
	}

	.field-error {
		color: #c33;
		font-size: 0.875rem;
		margin-top: 0.25rem;
	}

	.actions {
		display: flex;
		gap: 1rem;
		margin-top: 2rem;
	}

	button {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		transition: opacity 0.2s;
	}

	button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #4a90d9;
		color: white;
	}

	.btn-secondary {
		background: #f0f0f0;
		color: #333;
	}
</style>
