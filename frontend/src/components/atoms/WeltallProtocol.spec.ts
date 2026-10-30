import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import WeltallProtocol from "./WeltallProtocol.svelte";

describe("WeltallProtocol", () => {
  beforeEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("does not render when show is false", () => {
    render(WeltallProtocol, { props: { show: false } });

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders modal when show is true", async () => {
    const { container } = render(WeltallProtocol, { props: { show: true } });

    const dialog = container.querySelector(".protocol-modal");
    expect(dialog).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /weltall/i })).toBeInTheDocument();
  });

  it("closes when close button is clicked", async () => {
    const { container } = render(WeltallProtocol, { props: { show: true } });

    const closeButton = screen.getByLabelText(/close protocol panel/i);
    await fireEvent.click(closeButton);

    await waitFor(() => {
      expect(container.querySelector(".protocol-modal")).not.toBeInTheDocument();
    });
  });

  it("closes when overlay is clicked", async () => {
    const { container } = render(WeltallProtocol, { props: { show: true } });

    const overlay = container.querySelector(".protocol-overlay");
    expect(overlay).toBeInTheDocument();
    await fireEvent.click(overlay!);

    await waitFor(() => {
      expect(container.querySelector(".protocol-modal")).not.toBeInTheDocument();
    });
  });
});
