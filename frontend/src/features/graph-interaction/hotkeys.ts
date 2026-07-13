import type { TransformState, SimulationNode } from '$lib/components/GraphCanvas/types';
import type { GhostNodeState } from '$lib/components/GraphCanvas';

export interface HotkeysState {
  showSearchBox: boolean;
  searchQuery: string;
  searchMatchIds: string[];
  searchCurrentIndex: number;
  showHelpModal: boolean;
  showHelpTooltip: boolean;
  helpTooltipPosition: { x: number; y: number };
  helpTooltipMessage: string;
  inactivityTimeout: ReturnType<typeof setTimeout> | null;
  lastActivityTime: number;
}

export interface HotkeysCallbacks {
  onFocusModeToggle?: () => void;
  onSearchOpen?: () => void;
  onSearchClose?: () => void;
  onHelpToggle?: () => void;
  onGhostNodeCreate?: () => void;
  onNodeDelete?: (nodeId: string) => void;
  onUndo?: () => void;
}

export function createHotkeysState(): HotkeysState {
  return {
    showSearchBox: false,
    searchQuery: '',
    searchMatchIds: [],
    searchCurrentIndex: 0,
    showHelpModal: false,
    showHelpTooltip: false,
    helpTooltipPosition: { x: 0, y: 0 },
    helpTooltipMessage: '',
    inactivityTimeout: null,
    lastActivityTime: 0
  };
}

export function handleKeyDownEvent(
  e: KeyboardEvent,
  state: HotkeysState,
  canvas: HTMLCanvasElement | null,
  transform: TransformState,
  simNodes: SimulationNode[],
  ghostNode: GhostNodeState,
  selectedNodeId: string | null,
  showNoteForm: boolean,
  showLinkForm: boolean,
  searchInput: HTMLInputElement | null,
  callbacks: HotkeysCallbacks
): void {
  // Ignore hotkeys when typing in a form or search input
  const active = document.activeElement;
  const isTyping = active instanceof HTMLInputElement || active instanceof HTMLTextAreaElement || active instanceof HTMLSelectElement;

  if (e.key === 'Escape') {
    if (state.showHelpModal) {
      state.showHelpModal = false;
      callbacks.onHelpToggle?.();
    } else if (state.showSearchBox) {
      callbacks.onSearchClose?.();
    } else if (showNoteForm || showLinkForm) {
      // Let form close first; focus mode toggles only when no forms are open
      return;
    } else {
      callbacks.onFocusModeToggle?.();
    }
    e.preventDefault();
    return;
  }

  if (isTyping && !state.showSearchBox) {
    return;
  }

  if (e.key === 'f' && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
    e.preventDefault();
    callbacks.onSearchOpen?.();
    requestAnimationFrame(() => searchInput?.focus());
    return;
  }

  if (e.key === '?' && !e.ctrlKey && !e.metaKey && !e.altKey) {
    e.preventDefault();
    callbacks.onHelpToggle?.();
    return;
  }

  if (e.key === 'Enter' && state.showSearchBox) {
    e.preventDefault();
    focusNextSearchMatch(state, transform, simNodes, canvas);
    return;
  }

  // N - Create ghost node
  if (e.key === 'n' && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
    e.preventDefault();
    callbacks.onGhostNodeCreate?.();
    return;
  }

  // Delete/Backspace - Delete selected node (if not typing)
  if ((e.key === 'Delete' || e.key === 'Backspace') && !isTyping) {
    e.preventDefault();
    if (selectedNodeId) {
      callbacks.onNodeDelete?.(selectedNodeId);
    }
    return;
  }

  // Ctrl+Z - Undo (placeholder for now)
  if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
    e.preventDefault();
    callbacks.onUndo?.();
    return;
  }
}

export function updateSearch(
  state: HotkeysState,
  simNodes: SimulationNode[]
): void {
  const query = state.searchQuery.trim().toLowerCase();
  if (!query) {
    state.searchMatchIds = [];
    state.searchCurrentIndex = 0;
    return;
  }

  state.searchMatchIds = simNodes
    .filter((node) => node.title.toLowerCase().includes(query))
    .map((node) => node.id);
  state.searchCurrentIndex = 0;
}

export function focusNextSearchMatch(
  state: HotkeysState,
  transform: TransformState,
  simNodes: SimulationNode[],
  canvas: HTMLCanvasElement | null
): void {
  if (state.searchMatchIds.length === 0) return;
  state.searchCurrentIndex = (state.searchCurrentIndex + 1) % state.searchMatchIds.length;
  const nodeId = state.searchMatchIds[state.searchCurrentIndex];
  const node = simNodes.find((n) => n.id === nodeId);
  if (node && node.x != null && node.y != null && canvas) {
    const rect = canvas.getBoundingClientRect();
    transform.k = Math.max(transform.k, 1.2);
    transform.x = rect.width / 2 - node.x * transform.k;
    transform.y = rect.height / 2 - node.y * transform.k;
  }
}

export function resetInactivityTimer(
  state: HotkeysState,
  onTipShow: () => void
): void {
  if (state.inactivityTimeout) clearTimeout(state.inactivityTimeout);
  state.inactivityTimeout = setTimeout(() => {
    onTipShow();
  }, 10000);
}

export function updateActivity(state: HotkeysState, onInactivityTip: () => void): void {
  state.lastActivityTime = Date.now();
  // Hide tip on activity (not show it)
  if (state.showHelpTooltip && state.helpTooltipPosition.x === -1) {
    state.showHelpTooltip = false;
  }
  resetInactivityTimer(state, onInactivityTip);
}

export function showRandomTip(state: HotkeysState, tips: string[]): void {
  const tip = tips[Math.floor(Math.random() * tips.length)];
  state.helpTooltipMessage = tip;
  // Position at bottom-center; actual centering is handled via CSS transform
  state.helpTooltipPosition = { x: -1, y: -1 };
  state.showHelpTooltip = true;
  setTimeout(() => {
    state.showHelpTooltip = false;
  }, 4000);
}
