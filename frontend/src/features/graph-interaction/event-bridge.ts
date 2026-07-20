import {
  getLinkEndpointId,
  type SimulationNode,
  type SimulationLink,
  type TransformState,
  type DragState,
} from "$components/organisms/GraphCanvas";
import type {
  BlackHoleState,
  GhostNodeState,
  GravitySystem,
} from "$components/organisms/GraphCanvas";
import {
  findLinkAtPosition,
  getSimulationNodes,
  isNodeOverBlackHole,
  isPointOverBlackHole,
} from "$components/organisms/GraphCanvas";
import { createGhostNode } from "$components/organisms/GraphCanvas/ghost-node";
import { graphConfig2D } from "$shared/config";
import type { DragDropState } from "$features/graph-interaction/drag-and-drop";
import {
  getMouseWorldPosition,
  findNodeAtPosition,
  handleMouseDown,
} from "$features/graph-interaction/drag-and-drop";
import type { HotkeysState } from "$features/graph-interaction/hotkeys";
import {
  handleKeyDownEvent,
  updateActivity,
  showRandomTip,
} from "$features/graph-interaction/hotkeys";
import type { ZoomPanState } from "$features/graph-interaction/zoom-pan";
import {
  handleZoom,
  handleTouchStart,
} from "$features/graph-interaction/zoom-pan";
import type { NoteFormState } from "$features/graph-forms/note-form";
import { openNoteForm, closeNoteForm } from "$features/graph-forms/note-form";
import type { LinkFormState } from "$features/graph-forms/link-form";
import { openLinkForm, closeLinkForm } from "$features/graph-forms/link-form";
import type { SimulationState } from "$components/organisms/GraphCanvas/types";

export interface HoveredLinkInfo {
  source: string;
  target: string;
  link_type: string;
  weight: number;
  source_type: string;
}

export interface GraphCanvasEventContext {
  readonly: boolean;
  browser: boolean;
  isTechnicalNode(nodeId: string): boolean;

  getCanvas(): HTMLCanvasElement | null;
  getCtx(): CanvasRenderingContext2D | null;
  getWidth(): number;
  getHeight(): number;

  transform: TransformState;
  dragState: DragState;
  dragDropState: DragDropState;
  simState: SimulationState;
  hotkeysState: HotkeysState;
  noteFormState: NoteFormState;
  linkFormState: LinkFormState;
  zoomPanState: ZoomPanState;

  getGhostNode(): GhostNodeState;
  getBlackHole(): BlackHoleState;
  getGravitySystem(): GravitySystem | null;

  getHoveredNodeId(): string | null;
  setHoveredNodeId(value: string | null): void;
  getHoveredLink(): HoveredLinkInfo | null;
  setHoveredLink(value: HoveredLinkInfo | null): void;
  getTooltipPosition(): { x: number; y: number };
  setTooltipPosition(value: { x: number; y: number }): void;
  getFocusMode(): boolean;
  setFocusMode(value: boolean): void;
  getSelectedNodeId(): string | null;
  setSelectedNodeId(value: string | null): void;

  redraw(): void;
  toggleFocus(): void;
  openSearch(): void;
  closeSearch(): void;
  openHelp(): void;
  toggleHelp(): void;
  setGhostNode(node: GhostNodeState): void;

  onNodeClick?(node: { id: string; title: string; type?: string }): void;
  onNoteDelete?(nodeId: string): void;

  getKeyLines(): string[];
}

export interface GraphEventBridge {
  onMouseDown: (e: MouseEvent) => void;
  onMouseMove: (e: MouseEvent) => void;
  onMouseUp: (e: MouseEvent) => void;
  onClick: (e: MouseEvent) => void;
  onDblClick: (e: MouseEvent) => void;
  onZoom: (e: WheelEvent) => void;
  onTouchStart: (e: TouchEvent) => void;
  onWindowMouseUp: (e: MouseEvent) => void;
  handleKeyDown: (e: KeyboardEvent) => void;
  cleanup: () => void;
}

export function createGraphEventBridge(
  context: GraphCanvasEventContext,
): GraphEventBridge {
  const HOVER_DELAY_MS = graphConfig2D.hover_delay_ms;

  let hoverNodeTimeout: ReturnType<typeof setTimeout> | null = null;
  let hoverCandidateNodeId: string | null = null;
  let hoverLinkTimeout: ReturnType<typeof setTimeout> | null = null;
  let hoverCandidateLink: HoveredLinkInfo | null = null;
  let hoverCandidateLinkKey: string | null = null;

  function getLinkKey(
    link:
      | HoveredLinkInfo
      | SimulationLink
      | { source: string | { id: string }; target: string | { id: string } },
  ): string {
    const s = getLinkEndpointId(link.source);
    const t = getLinkEndpointId(link.target);
    return `${s}|${t}`;
  }

  function clearNodeHover() {
    if (hoverNodeTimeout) {
      clearTimeout(hoverNodeTimeout);
      hoverNodeTimeout = null;
    }
    hoverCandidateNodeId = null;
    context.setHoveredNodeId(null);
  }

  function scheduleNodeHover(nodeId: string) {
    if (hoverCandidateNodeId === nodeId && hoverNodeTimeout) {
      return;
    }
    if (hoverNodeTimeout) {
      clearTimeout(hoverNodeTimeout);
    }
    hoverCandidateNodeId = nodeId;
    hoverNodeTimeout = setTimeout(() => {
      if (hoverCandidateNodeId === nodeId) {
        context.setHoveredNodeId(nodeId);
      }
      hoverNodeTimeout = null;
    }, HOVER_DELAY_MS);
  }

  function clearLinkHover() {
    if (hoverLinkTimeout) {
      clearTimeout(hoverLinkTimeout);
      hoverLinkTimeout = null;
    }
    hoverCandidateLink = null;
    hoverCandidateLinkKey = null;
    context.setHoveredLink(null);
  }

  function scheduleLinkHover(
    link: HoveredLinkInfo,
    mouseX: number,
    mouseY: number,
  ) {
    const linkKey = getLinkKey(link);
    if (hoverCandidateLinkKey === linkKey && hoverLinkTimeout) {
      return;
    }
    if (hoverLinkTimeout) {
      clearTimeout(hoverLinkTimeout);
    }
    hoverCandidateLink = link;
    hoverCandidateLinkKey = linkKey;
    hoverLinkTimeout = setTimeout(() => {
      if (hoverCandidateLinkKey === linkKey) {
        context.setHoveredLink(hoverCandidateLink);
        context.setTooltipPosition({ x: mouseX, y: mouseY });
      }
      hoverLinkTimeout = null;
    }, HOVER_DELAY_MS);
  }

  function onMouseDown(e: MouseEvent) {
    if (context.readonly) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;

    handleMouseDown(
      e,
      canvas,
      context.transform,
      context.dragState,
      context.dragDropState,
      getSimulationNodes(context.simState),
      context.getGhostNode(),
      context.isTechnicalNode,
      {
        onNodeDragStart: (nodeId) => {
          if (nodeId === "ghost") {
            openNoteForm(context.noteFormState, e.clientX, e.clientY);
            context.redraw();
          }
        },
      },
    );

    if (context.dragDropState.draggedNodeId) {
      const node = getSimulationNodes(context.simState).find(
        (n) => n.id === context.dragDropState.draggedNodeId,
      );
      if (node && node.x != null && node.y != null) {
        node.fx = node.x;
        node.fy = node.y;
      }
    }

    clearNodeHover();
    clearLinkHover();
  }

  function onMouseMove(e: MouseEvent) {
    if (context.readonly) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;

    const pos = getMouseWorldPosition(e, canvas, context.transform);
    context.dragDropState.mouseWorldPosition = pos;

    const blackHole = context.getBlackHole();
    const ghostNode = context.getGhostNode();

    blackHole.hovered = isPointOverBlackHole(
      e.clientX,
      e.clientY,
      blackHole,
      context.transform,
    );

    const canvasRect = canvas.getBoundingClientRect();
    const screenX = e.clientX - canvasRect.left;
    const screenY = e.clientY - canvasRect.top;
    const dx = screenX - 60;
    const dy = screenY - 60;
    ghostNode.hovered = Math.sqrt(dx * dx + dy * dy) < ghostNode.radius;

    if (context.dragDropState.draggedNodeId && context.dragState.dragging) {
      const node = getSimulationNodes(context.simState).find(
        (n) => n.id === context.dragDropState.draggedNodeId,
      );
      if (node && node.x != null && node.y != null) {
        blackHole.hovered = isNodeOverBlackHole(node, blackHole);

        const targetNode = findNodeAtPosition(
          pos.x,
          pos.y,
          getSimulationNodes(context.simState),
        );
        if (
          targetNode &&
          targetNode.id !== context.dragDropState.draggedNodeId &&
          !context.isTechnicalNode(targetNode.id)
        ) {
          context.dragDropState.isDraggingForLink = true;
          context.dragDropState.linkTargetNodeId = targetNode.id;
          context.dragDropState.linkPreviewTarget = {
            sourceId: context.dragDropState.draggedNodeId,
            targetId: targetNode.id,
          };
        } else {
          context.dragDropState.isDraggingForLink = false;
          context.dragDropState.linkTargetNodeId = null;
          context.dragDropState.linkPreviewTarget = null;
        }
      }
      context.redraw();
      return;
    }

    if (context.dragState.dragging) {
      context.transform.x = e.clientX - context.dragState.dragStart.x;
      context.transform.y = e.clientY - context.dragState.dragStart.y;
      context.redraw();
      return;
    }

    const hovered = findLinkAtPosition(
      pos.x,
      pos.y,
      context.simState.simLinks,
      getSimulationNodes(context.simState),
      context.transform,
    );

    let foundHoveredNode = false;
    let hoveredTechnicalNode: SimulationNode | null = null;
    const simNodes = getSimulationNodes(context.simState);
    for (const node of simNodes) {
      if (node.x && node.y) {
        const ndx = pos.x - node.x;
        const ndy = pos.y - node.y;
        if (Math.sqrt(ndx * ndx + ndy * ndy) < 30) {
          foundHoveredNode = true;
          if (node.type === "technical") {
            hoveredTechnicalNode = node;
          }
          if (context.getHoveredNodeId() !== node.id) {
            scheduleNodeHover(node.id);
          }
          break;
        }
      }
    }
    if (!foundHoveredNode) {
      clearNodeHover();
    }

    if (hoveredTechnicalNode) {
      context.hotkeysState.helpTooltipMessage =
        "Click to open help, or press ?";
      context.hotkeysState.helpTooltipPosition = {
        x: e.clientX,
        y: e.clientY - 10,
      };
      context.hotkeysState.showHelpTooltip = true;
    } else if (
      context.hotkeysState.showHelpTooltip &&
      context.hotkeysState.helpTooltipMessage ===
        "Click to open help, or press ?"
    ) {
      context.hotkeysState.showHelpTooltip = false;
    }

    if (hovered) {
      const sourceNode =
        typeof hovered.source === "string"
          ? simNodes.find((n) => n.id === hovered.source)
          : (hovered.source as SimulationNode);
      const targetNode =
        typeof hovered.target === "string"
          ? simNodes.find((n) => n.id === hovered.target)
          : (hovered.target as SimulationNode);

      if (sourceNode && targetNode) {
        const linkData: HoveredLinkInfo = {
          source:
            typeof hovered.source === "string"
              ? hovered.source
              : (hovered.source as { id: string }).id,
          target:
            typeof hovered.target === "string"
              ? hovered.target
              : (hovered.target as { id: string }).id,
          link_type: hovered.link_type || "related",
          weight: hovered.weight ?? 0.5,
          source_type:
            (hovered as { source_type?: string }).source_type || "user",
        };

        const centerX =
          ((sourceNode.x! + targetNode.x!) / 2) * context.transform.k +
          context.transform.x +
          10;
        const centerY =
          ((sourceNode.y! + targetNode.y!) / 2) * context.transform.k +
          context.transform.y +
          10;

        if (
          !context.getHoveredLink() ||
          getLinkKey(context.getHoveredLink()!) !== getLinkKey(linkData)
        ) {
          scheduleLinkHover(linkData, centerX, centerY);
        }
      }
    } else {
      clearLinkHover();
    }
  }

  function onMouseUp(e: MouseEvent) {
    if (context.readonly) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;

    const pos = getMouseWorldPosition(e, canvas, context.transform);
    const wasDraggingNode = context.dragDropState.draggedNodeId !== null;

    if (context.dragDropState.draggedNodeId) {
      const node = getSimulationNodes(context.simState).find(
        (n) => n.id === context.dragDropState.draggedNodeId,
      );
      if (node) {
        node.fx = undefined;
        node.fy = undefined;

        const targetNode = findNodeAtPosition(
          pos.x,
          pos.y,
          getSimulationNodes(context.simState),
        );
        if (
          targetNode &&
          targetNode.id !== context.dragDropState.draggedNodeId &&
          !context.isTechnicalNode(targetNode.id)
        ) {
          openLinkForm(
            context.linkFormState,
            context.dragDropState.draggedNodeId,
            targetNode.id,
            e.clientX,
            e.clientY,
          );
        }
      }
      context.dragDropState.draggedNodeId = null;
      context.dragDropState.isDraggingForLink = false;
      context.dragDropState.linkPreviewTarget = null;
    }

    context.dragState.dragging = false;
    canvas.style.cursor = "grab";
    context.getBlackHole().hovered = false;

    if (!wasDraggingNode && !context.getGhostNode().hovered) {
      const clickedNode = findNodeAtPosition(
        pos.x,
        pos.y,
        getSimulationNodes(context.simState),
      );
      if (!clickedNode) {
        context.setHoveredLink(null);
      }
    }

    context.redraw();
  }

  function onClick(e: MouseEvent) {
    if (context.readonly) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;

    const pos = getMouseWorldPosition(e, canvas, context.transform);
    context.dragDropState.mouseWorldPosition = pos;

    const clickedNode = findNodeAtPosition(
      pos.x,
      pos.y,
      getSimulationNodes(context.simState),
    );
    if (clickedNode) {
      if (context.isTechnicalNode(clickedNode.id)) {
        context.openHelp();
        return;
      }
      context.setSelectedNodeId(clickedNode.id);
      context.onNodeClick?.({
        id: clickedNode.id,
        title: clickedNode.title,
        type: clickedNode.type,
      });
    } else {
      context.setSelectedNodeId(null);
    }
  }

  function onDblClick(e: MouseEvent) {
    if (context.readonly) return;
    const canvas = context.getCanvas();
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const zoomFactor = 1.8;
    const newScale = Math.min(context.transform.k * zoomFactor, 5);
    const scaleChange = newScale / context.transform.k;

    context.transform.x = mouseX - (mouseX - context.transform.x) * scaleChange;
    context.transform.y = mouseY - (mouseY - context.transform.y) * scaleChange;
    context.transform.k = newScale;

    context.redraw();
  }

  function onWindowMouseUp(e: MouseEvent) {
    if (context.readonly) return;
    if (!context.dragState.dragging) return;
    onMouseUp(e);
  }

  function onZoom(e: WheelEvent) {
    if (context.readonly) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;
    handleZoom(e, context.transform, canvas, context.redraw);
  }

  function onTouchStart(e: TouchEvent) {
    if (context.readonly || !context.browser) return;
    updateActivity(context.hotkeysState, () =>
      showRandomTip(context.hotkeysState, context.getKeyLines()),
    );
    const canvas = context.getCanvas();
    if (!canvas) return;
    handleTouchStart(
      e,
      context.zoomPanState,
      context.transform,
      canvas,
      getSimulationNodes(context.simState),
      context.getCtx(),
      context.getWidth(),
      context.getHeight(),
    );
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (context.readonly) return;
    handleKeyDownEvent(
      e,
      context.hotkeysState,
      context.getCanvas(),
      context.transform,
      getSimulationNodes(context.simState),
      context.getGhostNode(),
      context.getSelectedNodeId(),
      context.noteFormState.showNoteForm,
      context.linkFormState.showLinkForm,
      null,
      {
        onFocusModeToggle: () => context.toggleFocus(),
        onSearchOpen: () => context.openSearch(),
        onSearchClose: () => context.closeSearch(),
        onHelpToggle: () => context.toggleHelp(),
        onNoteFormClose: () => {
          closeNoteForm(context.noteFormState);
          context.redraw();
        },
        onLinkFormClose: () => {
          closeLinkForm(context.linkFormState);
          context.redraw();
        },
        onGhostNodeCreate: () => {
          const canvas = context.getCanvas();
          if (!canvas) return;
          const rect = canvas.getBoundingClientRect();
          const centerX =
            (rect.width / 2 - context.transform.x) / context.transform.k;
          const centerY =
            (rect.height / 2 - context.transform.y) / context.transform.k;
          context.setGhostNode(
            createGhostNode(
              rect.width,
              rect.height,
              getSimulationNodes(context.simState),
            ),
          );
          openNoteForm(context.noteFormState, centerX, centerY);
          context.redraw();
        },
        onNodeDelete: (nodeId) => {
          context.onNoteDelete?.(nodeId);
          context.setSelectedNodeId(null);
          context.redraw();
        },
        onUndo: () => {},
      },
    );
  }

  function cleanup() {
    if (hoverNodeTimeout) {
      clearTimeout(hoverNodeTimeout);
      hoverNodeTimeout = null;
    }
    if (hoverLinkTimeout) {
      clearTimeout(hoverLinkTimeout);
      hoverLinkTimeout = null;
    }
    hoverCandidateNodeId = null;
    hoverCandidateLink = null;
    hoverCandidateLinkKey = null;
  }

  return {
    onMouseDown,
    onMouseMove,
    onMouseUp,
    onClick,
    onDblClick,
    onWindowMouseUp,
    onZoom,
    onTouchStart,
    handleKeyDown,
    cleanup,
  };
}

export function attachEvents(
  canvas: HTMLCanvasElement,
  context: GraphCanvasEventContext,
  windowImpl: Window = globalThis.window,
): () => void {
  const bridge = createGraphEventBridge(context);

  canvas.addEventListener("mousedown", bridge.onMouseDown);
  canvas.addEventListener("mousemove", bridge.onMouseMove);
  canvas.addEventListener("mouseup", bridge.onMouseUp);
  canvas.addEventListener("click", bridge.onClick);
  canvas.addEventListener("dblclick", bridge.onDblClick);
  canvas.addEventListener("wheel", bridge.onZoom, { passive: false });
  canvas.addEventListener("touchstart", bridge.onTouchStart, {
    passive: false,
  });
  windowImpl.addEventListener("mouseup", bridge.onWindowMouseUp);
  windowImpl.addEventListener("keydown", bridge.handleKeyDown);

  return () => {
    canvas.removeEventListener("mousedown", bridge.onMouseDown);
    canvas.removeEventListener("mousemove", bridge.onMouseMove);
    canvas.removeEventListener("mouseup", bridge.onMouseUp);
    canvas.removeEventListener("click", bridge.onClick);
    canvas.removeEventListener("dblclick", bridge.onDblClick);
    canvas.removeEventListener("wheel", bridge.onZoom);
    canvas.removeEventListener("touchstart", bridge.onTouchStart);
    windowImpl.removeEventListener("mouseup", bridge.onWindowMouseUp);
    windowImpl.removeEventListener("keydown", bridge.handleKeyDown);
    bridge.cleanup();
  };
}
