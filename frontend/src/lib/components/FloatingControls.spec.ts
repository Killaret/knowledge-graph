import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import FloatingControls from './FloatingControls.svelte';

// Mock navigation
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

import { goto } from '$app/navigation';

describe('FloatingControls', () => {
  const mockCallbacks = {
    onCreate: vi.fn(),
    onSearch: vi.fn(),
    onToggleView: vi.fn(),
    onFilter: vi.fn(),
    onImport: vi.fn(),
    onExport: vi.fn()
  };

  const mockTypeFilters = [
    { id: 'star', label: 'Star', emoji: '⭐' },
    { id: 'planet', label: 'Planet', emoji: '🪐' },
    { id: 'comet', label: 'Comet', emoji: '☄️' }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders all control buttons', () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    // View toggle buttons
    expect(screen.getByTestId('view-toggle-graph')).toBeInTheDocument();
    expect(screen.getByTestId('view-toggle-3d')).toBeInTheDocument();
    expect(screen.getByTestId('view-toggle-list')).toBeInTheDocument();

    // Search input
    expect(screen.getByTestId('search-input')).toBeInTheDocument();

    // Create button
    expect(screen.getByTestId('create-note-button')).toBeInTheDocument();
  });

  it('calls onCreate when create button is clicked', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    const createBtn = screen.getByTestId('create-note-button');
    await fireEvent.click(createBtn);

    expect(mockCallbacks.onCreate).toHaveBeenCalledTimes(1);
  });

  it('calls onToggleView when 2D view button is clicked', async () => {
    render(FloatingControls, {
      props: { ...mockCallbacks, currentView: 'list' }
    });

    const graphBtn = screen.getByTestId('view-toggle-graph');
    await fireEvent.click(graphBtn);

    expect(mockCallbacks.onToggleView).toHaveBeenCalledTimes(1);
  });

  it('calls onToggleView when list view button is clicked', async () => {
    render(FloatingControls, {
      props: { ...mockCallbacks, currentView: 'graph' }
    });

    const listBtn = screen.getByTestId('view-toggle-list');
    await fireEvent.click(listBtn);

    expect(mockCallbacks.onToggleView).toHaveBeenCalledTimes(1);
  });

  it('navigates to 3D view when 3D button is clicked', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    const btn3d = screen.getByTestId('view-toggle-3d');
    await fireEvent.click(btn3d);

    expect(goto).toHaveBeenCalledWith('/graph/3d');
  });

  it('navigates to note-specific 3D view when noteId is provided', async () => {
    render(FloatingControls, {
      props: { ...mockCallbacks, noteId: 'note-123' }
    });

    const btn3d = screen.getByTestId('view-toggle-3d');
    await fireEvent.click(btn3d);

    expect(goto).toHaveBeenCalledWith('/graph/3d/note-123');
  });

  it('calls onSearch when search is submitted', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    const searchInput = screen.getByTestId('search-input');
    await fireEvent.input(searchInput, { target: { value: 'test query' } });
    await fireEvent.keyUp(searchInput, { key: 'Enter' });

    expect(mockCallbacks.onSearch).toHaveBeenCalledWith('test query');
  });

  it('renders type filters when provided', () => {
    render(FloatingControls, {
      props: {
        ...mockCallbacks,
        typeFilters: mockTypeFilters,
        selectedType: 'star'
      }
    });

    mockTypeFilters.forEach(filter => {
      expect(screen.getByTestId(`filter-chip-${filter.id}`)).toBeInTheDocument();
    });
  });

  it('calls onFilter when filter chip is clicked', async () => {
    render(FloatingControls, {
      props: {
        ...mockCallbacks,
        typeFilters: mockTypeFilters,
        selectedType: 'all'
      }
    });

    const starFilter = screen.getByTestId('filter-chip-star');
    await fireEvent.click(starFilter);

    expect(mockCallbacks.onFilter).toHaveBeenCalledWith('star');
  });

  it('displays type counts when provided', () => {
    render(FloatingControls, {
      props: {
        ...mockCallbacks,
        typeFilters: mockTypeFilters,
        typeCounts: { star: 5, planet: 3, comet: 1 }
      }
    });

    expect(screen.getByTestId('filter-count-star')).toHaveTextContent('5');
    expect(screen.getByTestId('filter-count-planet')).toHaveTextContent('3');
    expect(screen.getByTestId('filter-count-comet')).toHaveTextContent('1');
  });

  it('marks active filter chip', () => {
    render(FloatingControls, {
      props: {
        ...mockCallbacks,
        typeFilters: mockTypeFilters,
        selectedType: 'planet'
      }
    });

    const planetChip = screen.getByTestId('filter-chip-planet');
    expect(planetChip).toHaveClass('active');
  });

  it('toggles menu dropdown when menu button is clicked', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    // Menu should be closed initially
    expect(screen.queryByText('Import')).not.toBeInTheDocument();

    // Open menu
    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    // Menu items should be visible
    expect(screen.getByText('Import')).toBeInTheDocument();
    expect(screen.getByText('Export')).toBeInTheDocument();
    expect(screen.getByText('Full 3D View')).toBeInTheDocument();
  });

  it('calls onImport from menu', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const importBtn = screen.getByText('Import');
    await fireEvent.click(importBtn);

    expect(mockCallbacks.onImport).toHaveBeenCalledTimes(1);
  });

  it('calls onExport from menu', async () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const exportBtn = screen.getByText('Export');
    await fireEvent.click(exportBtn);

    expect(mockCallbacks.onExport).toHaveBeenCalledTimes(1);
  });

  it('works without optional callbacks', async () => {
    // Should not throw when callbacks are not provided
    render(FloatingControls, {
      props: {}
    });

    const createBtn = screen.getByTestId('create-note-button');
    await fireEvent.click(createBtn); // Should not throw

    const btn3d = screen.getByTestId('view-toggle-3d');
    await fireEvent.click(btn3d); // Should not throw
  });

  it('shows note-specific 3D view text when noteId is in menu', async () => {
    render(FloatingControls, {
      props: { ...mockCallbacks, noteId: 'note-123' }
    });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    expect(screen.getByText('3D View for Note')).toBeInTheDocument();
  });
});
