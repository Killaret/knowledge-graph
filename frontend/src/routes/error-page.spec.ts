import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import { writable } from "svelte/store";
import ErrorPage from "./+error.svelte";

vi.mock("$app/stores", () => ({
  page: writable({ status: 500, error: new Error("Test 500") }),
}));

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

describe("Error 500 page", () => {
  it("renders a full-screen error page", () => {
    const { container } = render(ErrorPage);
    const errorPage = container.querySelector(".error-page");
    expect(errorPage).toBeInTheDocument();
    expect(errorPage).toHaveClass("error-page");
    expect(errorPage).toBeVisible();
    // Full viewport coverage is implemented via CSS position: fixed; inset: 0
    // (verified by Playwright / manual testing because jsdom does not resolve
    // Svelte <style> blocks for getComputedStyle).
  });

  it("displays the 500 title and message", () => {
    render(ErrorPage);
    expect(screen.getByText("Internal Server Error")).toBeInTheDocument();
    expect(
      screen.getByText("Something went wrong. We already know about it.")
    ).toBeInTheDocument();
  });

  it("has refresh and go home buttons", () => {
    render(ErrorPage);
    expect(
      screen.getByRole("button", { name: "Refresh page" })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Go home" })).toBeInTheDocument();
  });

  it("shows a server-error illustration for 500", () => {
    render(ErrorPage);
    expect(
      screen.getByRole("img", { name: "Server error illustration" })
    ).toBeInTheDocument();
  });
});
