/**
 * LINKS-2-TAIL: deleting a model-suggested link — or one the model suggested and
 * the user confirmed — is a decision, not cleanup: the pair is recorded as
 * "not related" and will not be suggested again, so the page must ask first.
 */
export interface LinkDeleteCandidate {
  source_type?: string;
  gamma_origin?: boolean;
}

export function needsLinkDeleteConfirm(link: LinkDeleteCandidate): boolean {
  return link.source_type === "gamma" || link.gamma_origin === true;
}

export function linkDeleteConfirmKey(link: LinkDeleteCandidate): string {
  return link.source_type === "gamma" ? "link.deleteConfirmSuppress" : "link.deleteConfirmPromoted";
}
