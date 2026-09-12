import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import QuickCaptureWidget from "./QuickCaptureWidget.svelte";

// Mock the notes API
vi.mock("$shared/api/notes", () => ({
  createNote: vi.fn(),
}));

import { createNote } from "$shared/api/notes";

// Reset mocks before each test
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(createNote).mockResolvedValue({ id: "new" } as any);
});

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

describe("QuickCaptureWidget", () => {
  it("renders floating button when closed", () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    expect(button).toBeInTheDocument();
    expect(button?.textContent).toContain("✨");
  });

  it("renders in docked mode", () => {
    const { container } = render(QuickCaptureWidget, { props: { docked: true } });

    expect(container.querySelector(".quick-capture-container--docked")).toBeTruthy();
  });

  it("opens modal when button is clicked", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    expect(button).toBeInTheDocument();

    await fireEvent.click(button!);

    const modal = container.querySelector(".quick-capture-modal");
    expect(modal).toBeTruthy();
  });

  it("has correct keyboard shortcut Ctrl+Shift+N", async () => {
    const { container } = render(QuickCaptureWidget);

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "n",
        ctrlKey: true,
        shiftKey: true,
        bubbles: true,
        cancelable: true,
      })
    );

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
    });
  });

  it("does not open on Ctrl+Shift+N when already open", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const createNoteBefore = vi.mocked(createNote).mock.calls.length;

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "n",
        ctrlKey: true,
        shiftKey: true,
        bubbles: true,
        cancelable: true,
      })
    );

    expect(vi.mocked(createNote).mock.calls.length).toBe(createNoteBefore);
    expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
  });

  it("closes modal with Escape key", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    if (button) {
      await fireEvent.click(button);
    }

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
    });

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "Escape",
        bubbles: true,
        cancelable: true,
      })
    );

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeFalsy();
    });
  });

  it("does nothing on Escape when closed", async () => {
    const { container } = render(QuickCaptureWidget);

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "Escape",
        bubbles: true,
        cancelable: true,
      })
    );

    expect(container.querySelector(".quick-capture-modal")).toBeFalsy();
  });

  it("submits a quick capture", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "New note content" } });

    const submit = container.querySelector(".submit-btn");
    await fireEvent.click(submit!);

    await waitFor(() => {
      expect(createNote).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "New note content",
          content: "New note content",
          type: "dust",
        })
      );
    });
  });

  it("does not submit empty content", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const submit = container.querySelector(".submit-btn");
    await fireEvent.click(submit!);

    expect(createNote).not.toHaveBeenCalled();
  });

  it("submits with long content", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const longContent = "a".repeat(300);
    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: longContent } });

    const submit = container.querySelector(".submit-btn");
    await fireEvent.click(submit!);

    await waitFor(() => {
      expect(createNote).toHaveBeenCalledWith(
        expect.objectContaining({
          content: longContent,
          type: "dust",
        })
      );
    });
  });

  it("handles submit error", async () => {
    vi.mocked(createNote).mockRejectedValue(new Error("fail"));

    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "boom" } });

    const submit = container.querySelector(".submit-btn");
    await fireEvent.click(submit!);

    await waitFor(() => {
      expect(createNote).toHaveBeenCalled();
    });
  });

  it("submits with Ctrl+Enter", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "Hotkey submit" } });

    const event = new KeyboardEvent("keydown", {
      key: "Enter",
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    window.dispatchEvent(event);

    await waitFor(() => {
      expect(createNote).toHaveBeenCalled();
    });
  });

  it("submits with Meta+Enter", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "Mac submit" } });

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "Enter",
        metaKey: true,
        bubbles: true,
        cancelable: true,
      })
    );

    await waitFor(() => {
      expect(createNote).toHaveBeenCalled();
    });
  });

  it("ignores Enter without modifier when open", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "no modifier" } });

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "Enter",
        bubbles: true,
        cancelable: true,
      })
    );

    expect(createNote).not.toHaveBeenCalled();
  });

  it("closes on backdrop modal click", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
    });

    const modal = container.querySelector(".quick-capture-modal")!;
    await fireEvent.mouseDown(modal);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeFalsy();
    });
  });

  it("does not close on backdrop click outside modal", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-backdrop")).toBeTruthy();
    });

    const backdrop = container.querySelector(".quick-capture-backdrop")!;
    await fireEvent.mouseDown(backdrop);

    expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
  });

  it("closes with close and cancel buttons", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
    });

    const closeBtn = container.querySelector(".close-btn");
    await fireEvent.click(closeBtn!);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeFalsy();
    });

    await fireEvent.click(button!);
    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeTruthy();
    });

    const cancelBtn = container.querySelector(".cancel-btn");
    await fireEvent.click(cancelBtn!);

    await waitFor(() => {
      expect(container.querySelector(".quick-capture-modal")).toBeFalsy();
    });
  });

  it("shows success message after submit", async () => {
    const { container } = render(QuickCaptureWidget);

    const button = container.querySelector(".quick-capture-btn");
    await fireEvent.click(button!);

    const textarea = container.querySelector(".modal-body textarea") as HTMLTextAreaElement;
    await fireEvent.input(textarea, { target: { value: "Great idea" } });

    const submit = container.querySelector(".submit-btn");
    await fireEvent.click(submit!);

    await waitFor(() => {
      expect(container.querySelector(".success-message")).toBeTruthy();
    });
  });
});
