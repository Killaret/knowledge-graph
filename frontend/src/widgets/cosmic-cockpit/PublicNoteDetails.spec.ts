import { describe, it, expect, vi, afterEach } from "vitest";
import { render, cleanup, fireEvent } from "@testing-library/svelte";
import PublicNoteDetails from "./PublicNoteDetails.svelte";

describe("PublicNoteDetails", () => {
  afterEach(() => {
    cleanup();
  });

  const notes = [
    { id: "note-1", title: "Seed star 001", type: "star" },
    { id: "note-2", title: "Seed planet 002", type: "planet" },
    { id: "note-3", title: "Seed moon 003", type: "moon" },
  ];

  const links = [
    { id: "link-1", source: "note-1", target: "note-2", link_type: "related", weight: 0.8 },
    { id: "link-2", source: "note-3", target: "note-1", link_type: "child", weight: 0.4 },
  ];

  it("renders note title, type and public visibility", () => {
    const { getByText } = render(PublicNoteDetails, {
      props: { nodeId: "note-1", notes, links },
    });

    expect(getByText("Seed star 001")).toBeInTheDocument();
    expect(getByText("Star")).toBeInTheDocument();
    expect(getByText("Public")).toBeInTheDocument();
  });

  it("renders connected notes with link type and direction", () => {
    const { getByText } = render(PublicNoteDetails, {
      props: { nodeId: "note-1", notes, links },
    });

    expect(getByText("Connected notes (2)")).toBeInTheDocument();
    expect(getByText("Seed planet 002")).toBeInTheDocument();
    expect(getByText("Seed moon 003")).toBeInTheDocument();
    expect(getByText("Link weight 0.8")).toBeInTheDocument();
    expect(getByText("Link weight 0.4")).toBeInTheDocument();
  });

  it("renders sign-in prompt when onSignIn is provided", () => {
    const onSignIn = vi.fn();
    const { getByText } = render(PublicNoteDetails, {
      props: { nodeId: "note-1", notes, links, onSignIn },
    });

    expect(getByText("Sign in to view full content and manage notes.")).toBeInTheDocument();
    const signInBtn = getByText("Sign in");
    expect(signInBtn).toBeInTheDocument();

    fireEvent.click(signInBtn);
    expect(onSignIn).toHaveBeenCalled();
  });

  it("calls onClose when close button is clicked", () => {
    const onClose = vi.fn();
    const { getByTitle } = render(PublicNoteDetails, {
      props: { nodeId: "note-1", notes, links, onClose },
    });

    const closeBtn = getByTitle("Close panel");
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalled();
  });

  it("renders 'no links' message when there are no related notes", () => {
    const { getByText } = render(PublicNoteDetails, {
      props: { nodeId: "note-1", notes, links: [] },
    });

    expect(getByText("Connected notes (0)")).toBeInTheDocument();
    expect(getByText("No connections yet.")).toBeInTheDocument();
  });

  it("renders not found state when nodeId is not in notes", () => {
    const { getByText } = render(PublicNoteDetails, {
      props: { nodeId: "unknown", notes, links },
    });

    expect(getByText("Note not found")).toBeInTheDocument();
  });
});
