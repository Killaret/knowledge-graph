import { graphStore } from "$shared/stores/graph.svelte";
import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

// Types used only for public API signatures; no runtime dependency on other features or components.
export interface HotkeysState {
  showSearchBox: boolean;
  searchQuery: string;
  searchMatchIds: string[];
  searchCurrentIndex: number;
  showHelpModal: boolean;
  showHelpTooltip: boolean;
}

const locale = getCurrentLocale();
const t = (key: string, params?: Record<string, string | number>) =>
  formatMessage(key, locale, params);

export interface HoveredLinkInfo {
  source: string;
  target: string;
  link_type: string;
  weight: number;
  source_type: string;
}

export interface DuplicateWarning {
  message: string;
  x: number;
  y: number;
  linkId: string;
}

export function createGraphCanvasState() {
  let hoveredLink = $state<HoveredLinkInfo | null>(null);
  let tooltipPosition = $state({ x: 0, y: 0 });
  let hoveredNodeId = $state<string | null>(null);

  let duplicateWarning = $state<DuplicateWarning | null>(null);
  let duplicateWarningTimeout = $state<ReturnType<typeof setTimeout> | null>(null);
  let highlightedLinkId = $state<string | null>(null);
  let highlightedLinkTimeout = $state<ReturnType<typeof setTimeout> | null>(null);

  let lastDeletedNodeId = $state<string | null>(null);
  let showUndoToast = $state(false);
  let undoToastStage = $state<"done" | "restore">("done");
  let undoToastTimeout = $state<ReturnType<typeof setTimeout> | null>(null);

  let focusMode = $state(false);

  const hotkeyLines = [
    t("hotkey.search"),
    t("hotkey.focus"),
    t("hotkey.help"),
    t("hotkey.ghostNode"),
    t("hotkey.delete"),
    t("hotkey.undo"),
    t("hotkey.quickCapture"),
    t("hotkey.dragLink"),
    t("hotkey.dragDelete"),
    t("hotkey.doubleClick"),
    t("hotkey.zoom"),
  ];

  function showDuplicateWarning(
    source: string,
    target: string,
    linkType: string,
    x: number,
    y: number
  ) {
    const stableLinkId = `${source}-${target}-${linkType}`;
    highlightedLinkId = stableLinkId;
    duplicateWarning = {
      message: t("duplicate.link"),
      x,
      y,
      linkId: stableLinkId,
    };

    if (duplicateWarningTimeout) clearTimeout(duplicateWarningTimeout);
    if (highlightedLinkTimeout) clearTimeout(highlightedLinkTimeout);

    duplicateWarningTimeout = setTimeout(() => {
      duplicateWarning = null;
    }, 2000);

    highlightedLinkTimeout = setTimeout(() => {
      highlightedLinkId = null;
    }, 1000);
  }

  function showUndoToastFor(nodeId: string) {
    lastDeletedNodeId = nodeId;
    showUndoToast = true;
    undoToastStage = "done";
    if (undoToastTimeout) clearTimeout(undoToastTimeout);

    undoToastTimeout = setTimeout(() => {
      undoToastStage = "restore";
      undoToastTimeout = setTimeout(() => {
        showUndoToast = false;
        lastDeletedNodeId = null;
        undoToastStage = "done";
      }, 5000);
    }, 1500);
  }

  function restoreDeletedNode(onNoteRestore?: (nodeId: string) => void) {
    if (lastDeletedNodeId && onNoteRestore) {
      onNoteRestore(lastDeletedNodeId);
    }
    showUndoToast = false;
    lastDeletedNodeId = null;
    undoToastStage = "done";
    if (undoToastTimeout) clearTimeout(undoToastTimeout);
  }

  function cancelUndo() {
    showUndoToast = false;
    lastDeletedNodeId = null;
    undoToastStage = "done";
    if (undoToastTimeout) clearTimeout(undoToastTimeout);
  }

  function handleLinkEdit(
    onLinkEdit?: (link: {
      source: string;
      target: string;
      link_type: string;
      weight: number;
    }) => void
  ) {
    if (hoveredLink && onLinkEdit) {
      onLinkEdit(hoveredLink);
    }
    hoveredLink = null;
  }

  function handleLinkDelete(
    onLinkDelete?: (link: { source: string; target: string; link_type: string }) => void
  ) {
    if (hoveredLink && onLinkDelete) {
      onLinkDelete({
        source: hoveredLink.source,
        target: hoveredLink.target,
        link_type: hoveredLink.link_type,
      });
    }
    hoveredLink = null;
  }

  function openHelpModal(hotkeysState: HotkeysState) {
    hotkeysState.showHelpModal = true;
    hotkeysState.showHelpTooltip = false;
  }

  function closeHelpModal(hotkeysState: HotkeysState) {
    hotkeysState.showHelpModal = false;
  }

  function handleCloseSearch(hotkeysState: HotkeysState, redraw?: () => void) {
    hotkeysState.showSearchBox = false;
    hotkeysState.searchQuery = "";
    hotkeysState.searchMatchIds = [];
    hotkeysState.searchCurrentIndex = 0;
    redraw?.();
  }

  function handleOpenSearch(hotkeysState: HotkeysState) {
    hotkeysState.showSearchBox = true;
    hotkeysState.searchQuery = "";
    hotkeysState.searchMatchIds = [];
    hotkeysState.searchCurrentIndex = 0;
  }

  function handleToggleFocus(redraw?: () => void) {
    focusMode = !focusMode;
    redraw?.();
  }

  return {
    get hoveredLink() {
      return hoveredLink;
    },
    set hoveredLink(value) {
      hoveredLink = value;
    },
    get tooltipPosition() {
      return tooltipPosition;
    },
    set tooltipPosition(value) {
      tooltipPosition = value;
    },
    get hoveredNodeId() {
      return hoveredNodeId;
    },
    set hoveredNodeId(value) {
      hoveredNodeId = value;
    },
    get duplicateWarning() {
      return duplicateWarning;
    },
    get duplicateWarningTimeout() {
      return duplicateWarningTimeout;
    },
    get highlightedLinkId() {
      return highlightedLinkId;
    },
    get highlightedLinkTimeout() {
      return highlightedLinkTimeout;
    },
    get lastDeletedNodeId() {
      return lastDeletedNodeId;
    },
    get showUndoToast() {
      return showUndoToast;
    },
    get undoToastStage() {
      return undoToastStage;
    },
    get undoToastTimeout() {
      return undoToastTimeout;
    },
    get focusMode() {
      return focusMode;
    },
    set focusMode(value) {
      focusMode = value;
    },
    get selectedNodeId() {
      return graphStore.selectedNodeId;
    },
    set selectedNodeId(value) {
      graphStore.selectedNodeId = value;
    },
    hotkeyLines,
    showDuplicateWarning,
    showUndoToastFor,
    restoreDeletedNode,
    cancelUndo,
    handleLinkEdit,
    handleLinkDelete,
    openHelpModal,
    closeHelpModal,
    handleCloseSearch,
    handleOpenSearch,
    handleToggleFocus,
  };
}

export function isTechnicalNode(
  nodes: Array<{ id: string; type?: string }>,
  nodeId: string
): boolean {
  return nodes.some((n) => n.id === nodeId && n.type === "technical");
}

export function pinTechnicalNodes<
  T extends {
    id: string;
    title: string;
    type?: string;
    createdAt?: string;
    created_at?: string;
  },
>(nodeList: T[]): Array<T & { x?: number; y?: number; fx?: number; fy?: number }> {
  return nodeList.map((n) => {
    if (n.type === "technical") {
      const padding = 60;
      return {
        ...n,
        x: padding,
        y: padding,
        fx: padding,
        fy: padding,
      };
    }
    return n;
  });
}
