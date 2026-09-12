import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/svelte";
import { goto } from "$app/navigation";
import BackButton from "./BackButton.svelte";

vi.mock("$app/environment", () => ({
  browser: true,
}));

describe("BackButton", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
  });
  it("renders back button", () => {
    render(BackButton);

    const button = screen.getByRole("button") || screen.getByText(/back|←/i);

    expect(button).toBeInTheDocument();
  });

  it("renders with custom text", () => {
    render(BackButton, { props: { text: "Go Back" } });

    expect(screen.getByText("Go Back")).toBeInTheDocument();
  });

  it("has correct href default", () => {
    render(BackButton);

    const button = screen.getByRole("button");

    expect(button).toBeInTheDocument();
  });

  it("calls history.back when browser history is available", async () => {
    const backSpy = vi.spyOn(window.history, "back").mockImplementation(() => {});
    Object.defineProperty(window.history, "length", { value: 2, configurable: true });

    render(BackButton);
    const button = screen.getByRole("button");
    await fireEvent.click(button);

    expect(backSpy).toHaveBeenCalled();
    expect(goto).not.toHaveBeenCalled();
  });

  it("falls back to goto when no browser history", async () => {
    Object.defineProperty(window.history, "length", { value: 1, configurable: true });

    render(BackButton);
    const button = screen.getByRole("button");
    await fireEvent.click(button);

    expect(goto).toHaveBeenCalledWith("/");
  });
});
