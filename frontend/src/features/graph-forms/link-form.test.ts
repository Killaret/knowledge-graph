import { describe, it, expect, vi } from "vitest";
import {
  createLinkFormState,
  openLinkForm,
  closeLinkForm,
  createLink,
  isLinkFormValid,
  isDuplicateLink,
  type LinkFormCallbacks,
} from "./link-form";

describe("link-form", () => {
  it("creates default link form state", () => {
    const state = createLinkFormState();
    expect(state.showLinkForm).toBe(false);
    expect(state.linkSourceNodeId).toBeNull();
    expect(state.newLinkType).toBe("related");
  });

  it("opens the link form with source and target", () => {
    const state = createLinkFormState();
    openLinkForm(state, "n1", "n2", 50, 60);
    expect(state.showLinkForm).toBe(true);
    expect(state.linkSourceNodeId).toBe("n1");
    expect(state.linkTargetNodeId).toBe("n2");
    expect(state.linkFormPosition).toEqual({ x: 50, y: 60 });
  });

  it("closes and resets the form", () => {
    const state = createLinkFormState();
    openLinkForm(state, "n1", "n2", 0, 0);
    closeLinkForm(state);
    expect(state.showLinkForm).toBe(false);
    expect(state.linkSourceNodeId).toBeNull();
    expect(state.newLinkType).toBe("related");
  });

  it("creates a link when valid", () => {
    const state = createLinkFormState();
    openLinkForm(state, "n1", "n2", 10, 20);
    state.newLinkType = "reference";
    state.newLinkWeight = 0.8;

    const callbacks: LinkFormCallbacks = {
      onLinkCreate: vi.fn(),
      onFormClose: vi.fn(),
    };

    createLink(state, [], callbacks);
    expect(callbacks.onLinkCreate).toHaveBeenCalledWith({
      source: "n1",
      target: "n2",
      link_type: "reference",
      weight: 0.8,
    });
    expect(callbacks.onFormClose).toHaveBeenCalled();
    expect(state.showLinkForm).toBe(false);
  });

  it("warns about duplicate links and does not create", () => {
    const state = createLinkFormState();
    openLinkForm(state, "n1", "n2", 10, 20);
    state.newLinkType = "reference";

    const callbacks: LinkFormCallbacks = {
      onLinkCreate: vi.fn(),
      onDuplicateWarning: vi.fn(),
      onFormClose: vi.fn(),
    };

    const links = [{ source: "n1", target: "n2", link_type: "reference" }];
    createLink(state, links, callbacks);

    expect(callbacks.onDuplicateWarning).toHaveBeenCalledWith("n1", "n2", "reference", 10, 20);
    expect(callbacks.onLinkCreate).not.toHaveBeenCalled();
    expect(state.showLinkForm).toBe(false);
  });

  it("detects duplicate links in both directions", () => {
    const links = [{ source: "n2", target: "n1", link_type: "custom" }];
    expect(isDuplicateLink("n1", "n2", "custom", links)).toBe(true);
    expect(isDuplicateLink("n1", "n2", "reference", links)).toBe(false);
  });

  it("detects duplicate links with object node references", () => {
    const links = [{ source: { id: "n1" }, target: { id: "n2" }, link_type: "related" }];
    expect(isDuplicateLink("n1", "n2", "related", links)).toBe(true);
  });

  it("reports link form validity", () => {
    const valid = createLinkFormState();
    openLinkForm(valid, "a", "b", 0, 0);
    expect(isLinkFormValid(valid)).toBe(true);

    const invalid = createLinkFormState();
    expect(isLinkFormValid(invalid)).toBe(false);
  });
});
