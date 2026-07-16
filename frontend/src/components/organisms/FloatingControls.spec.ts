import { describe, it, expect, vi, beforeAll, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import FloatingControls from './FloatingControls.svelte';

// Mock navigation (FloatingControls imports goto; 3D entry points are commented out for v1)
vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

// Mock auth store
vi.mock('$shared/stores/auth.svelte', () => ({
  isAuthenticated: vi.fn()
}));

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

    // View toggles: 2D and list (3D frozen for v1 — see LazyGraph3D / FloatingControls)
    expect(screen.getByTestId('view-toggle-graph')).toBeInTheDocument();
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

  it('does not render 3D toolbar toggle while 3D is frozen', () => {
    render(FloatingControls, {
      props: mockCallbacks
    });

    expect(screen.queryByTestId('view-toggle-3d')).not.toBeInTheDocument();
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

    // Import / Export only (menu 3D entries frozen for v1)
    expect(screen.getByText('Import')).toBeInTheDocument();
    expect(screen.getByText('Export')).toBeInTheDocument();
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

    expect(screen.queryByTestId('view-toggle-3d')).not.toBeInTheDocument();
  });

  it('does not show 3D menu entry while 3D is frozen', async () => {
    render(FloatingControls, {
      props: { ...mockCallbacks, noteId: 'note-123' }
    });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    expect(screen.queryByText('3D View for Note')).not.toBeInTheDocument();
    expect(screen.queryByText('Full 3D View')).not.toBeInTheDocument();
  });
});

describe('Login Button', () => {
  let isAuthenticated: any;
  let goto: any;
  
  const mockCallbacks = {
    onCreate: vi.fn(),
    onSearch: vi.fn(),
    onToggleView: vi.fn(),
    onFilter: vi.fn(),
    onImport: vi.fn(),
    onExport: vi.fn()
  };

  beforeAll(async () => {
    const auth = await import('$shared/stores/auth.svelte');
    const nav = await import('$app/navigation');
    isAuthenticated = auth.isAuthenticated;
    goto = nav.goto;
  });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows Login button when user is not authenticated', async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    render(FloatingControls, { props: mockCallbacks });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const loginBtn = screen.getByTestId('menu-login');
    expect(loginBtn).toBeInTheDocument();
    expect(loginBtn).toHaveTextContent('🔑 Login');
  });

  it('hides Login button when user is authenticated', async () => {
    vi.mocked(isAuthenticated).mockReturnValue(true);
    render(FloatingControls, { props: mockCallbacks });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const loginBtn = screen.queryByTestId('menu-login');
    expect(loginBtn).not.toBeInTheDocument();
  });

  it('navigates to login page when Login is clicked', async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    render(FloatingControls, { props: mockCallbacks });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const loginBtn = screen.getByTestId('menu-login');
    await fireEvent.click(loginBtn);

    expect(goto).toHaveBeenCalledWith('/auth/login');
  });

  it('closes menu after clicking Login', async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    render(FloatingControls, { props: mockCallbacks });

    const menuBtn = screen.getByTitle('Menu');
    await fireEvent.click(menuBtn);

    const loginBtn = screen.getByTestId('menu-login');
    await fireEvent.click(loginBtn);

    // Menu should be closed
    expect(screen.queryByText('Import')).not.toBeInTheDocument();
  });
});
