import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import NoteSidePanel from './NoteSidePanel.svelte';

// Mock the API
vi.mock('$lib/api/notes', () => ({
  getNote: vi.fn()
}));

// Mock navigation
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

// Mock date formatter
vi.mock('$lib/utils/date', () => ({
  formatDate: vi.fn((date: string) => new Date(date).toLocaleDateString())
}));

import { getNote } from '$lib/api/notes';
import { goto } from '$app/navigation';

describe('NoteSidePanel', () => {
  const mockNote = {
    id: 'note-1',
    title: 'Test Note Title',
    content: 'This is the note content',
    type: 'star',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-20T14:30:00Z',
    metadata: {
      tags: ['tag1', 'tag2']
    }
  };

  const mockCallbacks = {
    onClose: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn()
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', async () => {
    // Delay the resolution to keep loading state
    vi.mocked(getNote).mockImplementation(() => new Promise(() => {}));

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await tick();

    expect(screen.getByText('Loading note...')).toBeInTheDocument();
    expect(document.querySelector('.spinner')).toBeInTheDocument();
  });

  it('renders note details after loading', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    // Check all note details are rendered
    expect(screen.getByText('This is the note content')).toBeInTheDocument();
    expect(screen.getByText('star')).toBeInTheDocument();
    expect(screen.getByText('#tag1')).toBeInTheDocument();
    expect(screen.getByText('#tag2')).toBeInTheDocument();
    
    // Check type icon
    expect(screen.getByText('⭐')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    const closeBtn = screen.getByLabelText('Close panel');
    await fireEvent.click(closeBtn);

    expect(mockCallbacks.onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onEdit when edit button is clicked', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    const editBtn = screen.getByLabelText('Edit note');
    await fireEvent.click(editBtn);

    expect(mockCallbacks.onEdit).toHaveBeenCalledWith('note-1');
  });

  it('calls onDelete when delete button is clicked', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    const deleteBtn = screen.getByLabelText('Delete note');
    await fireEvent.click(deleteBtn);

    expect(mockCallbacks.onDelete).toHaveBeenCalledWith('note-1');
  });

  it('navigates to full page on "View Full Page" click', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    const viewFullBtn = screen.getByText(/view full page/i);
    await fireEvent.click(viewFullBtn);

    expect(goto).toHaveBeenCalledWith('/notes/note-1');
  });

  it('shows error state when note fails to load', async () => {
    vi.mocked(getNote).mockRejectedValue(new Error('Network error'));

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Failed to load note')).toBeInTheDocument();
    });
  });

  it('shows different type icons based on note type', async () => {
    const types = [
      { type: 'star', icon: '⭐' },
      { type: 'planet', icon: '🪐' },
      { type: 'comet', icon: '☄️' },
      { type: 'galaxy', icon: '🌀' },
      { type: 'asteroid', icon: '☁️' },
      { type: undefined, icon: '⭐' },
      { type: 'unknown', icon: '📄' }
    ];

    for (const { type, icon } of types) {
      vi.clearAllMocks();
      const noteWithType = { ...mockNote, type };
      vi.mocked(getNote).mockResolvedValue(noteWithType);

      const { unmount } = render(NoteSidePanel, {
        props: {
          nodeId: 'note-1',
          ...mockCallbacks
        }
      });

      await waitFor(() => {
        expect(screen.getByText(icon)).toBeInTheDocument();
      });

      unmount();
    }
  });

  it('does not show tags section when note has no tags', async () => {
    const noteWithoutTags = { ...mockNote, metadata: {} };
    vi.mocked(getNote).mockResolvedValue(noteWithoutTags);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    // Tags should not be present
    expect(screen.queryByText('#tag1')).not.toBeInTheDocument();
    expect(screen.queryByText('#tag2')).not.toBeInTheDocument();
  });

  it('loads note for different nodeId', async () => {
    vi.mocked(getNote).mockResolvedValue(mockNote);

    render(NoteSidePanel, {
      props: {
        nodeId: 'note-1',
        ...mockCallbacks
      }
    });

    await waitFor(() => {
      expect(screen.getByText('Test Note Title')).toBeInTheDocument();
    });

    expect(getNote).toHaveBeenCalledWith('note-1');
  });
});
