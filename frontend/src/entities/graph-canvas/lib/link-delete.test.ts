import { describe, expect, it } from "vitest";

import {
  linkDeleteConfirmKey,
  needsLinkDeleteConfirm,
} from "$entities/graph-canvas/lib/link-delete";

describe("link-delete guard", () => {
  it("asks before rejecting a live model suggestion (source_type gamma)", () => {
    expect(needsLinkDeleteConfirm({ source_type: "gamma" })).toBe(true);
  });

  it("asks before deleting a promoted link (gamma_origin survives promotion)", () => {
    expect(needsLinkDeleteConfirm({ source_type: "user", gamma_origin: true })).toBe(true);
  });

  it("does not ask for a plain manual link", () => {
    expect(needsLinkDeleteConfirm({ source_type: "user" })).toBe(false);
    expect(needsLinkDeleteConfirm({})).toBe(false);
  });

  it("picks the suggestion wording for a live suggestion", () => {
    expect(linkDeleteConfirmKey({ source_type: "gamma" })).toBe("link.deleteConfirmSuppress");
  });

  it("picks the promoted wording when the user confirmed the suggestion", () => {
    expect(linkDeleteConfirmKey({ source_type: "user", gamma_origin: true })).toBe(
      "link.deleteConfirmPromoted"
    );
  });
});
