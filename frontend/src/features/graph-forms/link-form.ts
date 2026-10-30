import { LinkType } from "$entities";

export interface LinkFormState {
  showLinkForm: boolean;
  linkFormPosition: { x: number; y: number };
  linkSourceNodeId: string | null;
  linkTargetNodeId: string | null;
  newLinkType: string;
  newLinkWeight: number;
}

export interface LinkFormCallbacks {
  onLinkCreate?: (link: {
    source: string;
    target: string;
    link_type: string;
    weight: number;
  }) => void;
  onFormClose?: () => void;
  onDuplicateWarning?: (
    source: string,
    target: string,
    linkType: string,
    x: number,
    y: number
  ) => void;
}

export function createLinkFormState(): LinkFormState {
  return {
    showLinkForm: false,
    linkFormPosition: { x: 0, y: 0 },
    linkSourceNodeId: null,
    linkTargetNodeId: null,
    newLinkType: LinkType.RELATED.type,
    newLinkWeight: LinkType.RELATED.defaultWeight,
  };
}

export function openLinkForm(
  state: LinkFormState,
  sourceId: string,
  targetId: string,
  x: number,
  y: number
): void {
  state.showLinkForm = true;
  state.linkFormPosition = { x, y };
  state.linkSourceNodeId = sourceId;
  state.linkTargetNodeId = targetId;
  state.newLinkType = LinkType.RELATED.type;
  state.newLinkWeight = LinkType.RELATED.defaultWeight;
}

export function closeLinkForm(state: LinkFormState): void {
  state.showLinkForm = false;
  state.linkSourceNodeId = null;
  state.linkTargetNodeId = null;
  state.newLinkType = LinkType.RELATED.type;
  state.newLinkWeight = LinkType.RELATED.defaultWeight;
}

export function createLink(
  state: LinkFormState,
  links: Array<{ source: string; target: string; link_type?: string }>,
  callbacks: LinkFormCallbacks
): void {
  if (state.linkSourceNodeId && state.linkTargetNodeId) {
    const linkType = state.newLinkType;
    if (isDuplicateLink(state.linkSourceNodeId, state.linkTargetNodeId, linkType, links)) {
      callbacks.onDuplicateWarning?.(
        state.linkSourceNodeId,
        state.linkTargetNodeId,
        linkType,
        state.linkFormPosition.x,
        state.linkFormPosition.y
      );
      closeLinkForm(state);
      return;
    }
    if (callbacks.onLinkCreate) {
      callbacks.onLinkCreate({
        source: state.linkSourceNodeId,
        target: state.linkTargetNodeId,
        link_type: linkType,
        weight: state.newLinkWeight,
      });
    }
  }
  closeLinkForm(state);
  callbacks.onFormClose?.();
}

export function isDuplicateLink(
  source: string,
  target: string,
  linkType: string,
  links: Array<{
    source: string | { id: string };
    target: string | { id: string };
    link_type?: string;
  }>
): boolean {
  return links.some((link) => {
    const s = typeof link.source === "string" ? link.source : link.source.id;
    const t = typeof link.target === "string" ? link.target : link.target.id;
    return (
      (s === source && t === target && link.link_type === linkType) ||
      (s === target && t === source && link.link_type === linkType)
    );
  });
}

export function isLinkFormValid(state: LinkFormState): boolean {
  return state.linkSourceNodeId !== null && state.linkTargetNodeId !== null;
}
