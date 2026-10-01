import { describe, it, expect, afterEach, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
import type { User } from "$shared/types";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

const authMock = vi.hoisted(() => ({
  isAuthenticated: vi.fn<() => boolean>(() => true),
  currentUser: vi.fn<() => User | null>(() => null),
  logout: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", () => authMock);

const notesApiMock = vi.hoisted(() => ({
  listComets: vi.fn(),
  downloadCometIcs: vi.fn(),
}));

vi.mock("$shared/api/notes", () => notesApiMock);

const baseComet = {
  content: "",
  metadata: {},
  type: "comet",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

afterEach(() => cleanup());

beforeEach(() => {
  authMock.isAuthenticated.mockReturnValue(true);
  notesApiMock.listComets.mockReset();
});

// COMET-1 этап B: секция «Ближайшие дела» — группы overdue/upcoming/undated.
describe("CockpitLeftPanel — upcoming comets", () => {
  it("renders overdue, upcoming and undated groups from listComets", async () => {
    notesApiMock.listComets.mockResolvedValueOnce({
      overdue: [
        { ...baseComet, id: "c-over", title: "Late visit", due_at: "2020-01-01T10:00:00Z" },
      ],
      upcoming: [{ ...baseComet, id: "c-soon", title: "Dentist", due_at: "2099-01-01T10:00:00Z" }],
      undated: [{ ...baseComet, id: "c-none", title: "Someday", due_at: null }],
    });

    render(CockpitLeftPanel, { props: { onNoteSelect: vi.fn() } });

    await waitFor(() => {
      expect(screen.getByText("Late visit")).toBeTruthy();
      expect(screen.getByText("Dentist")).toBeTruthy();
      expect(screen.getByText("Someday")).toBeTruthy();
    });

    expect(document.querySelector('[data-testid="comet-item-overdue"]')).toBeTruthy();
    expect(document.querySelector('[data-testid="comet-item-upcoming"]')).toBeTruthy();
    expect(document.querySelector('[data-testid="comet-item-undated"]')).toBeTruthy();
  });

  it("calls onNoteSelect when a comet is clicked", async () => {
    const onNoteSelect = vi.fn();
    notesApiMock.listComets.mockResolvedValueOnce({
      overdue: [],
      upcoming: [{ ...baseComet, id: "c-1", title: "Dentist", due_at: "2099-01-01T10:00:00Z" }],
      undated: [],
    });

    render(CockpitLeftPanel, { props: { onNoteSelect } });

    const item = await waitFor(() => screen.getByText("Dentist"));
    await fireEvent.click(item);
    expect(onNoteSelect).toHaveBeenCalledWith("c-1");
  });

  it("hides the section when there are no comets", async () => {
    notesApiMock.listComets.mockResolvedValueOnce({ overdue: [], upcoming: [], undated: [] });

    render(CockpitLeftPanel, { props: {} });

    await waitFor(() => expect(notesApiMock.listComets).toHaveBeenCalled());
    expect(document.querySelector('[data-testid="comet-item-overdue"]')).toBeNull();
    expect(document.querySelector("#comets-content")).toBeNull();
  });

  it("renders a calendar button only on dated comets and downloads .ics on click", async () => {
    notesApiMock.listComets.mockResolvedValueOnce({
      overdue: [],
      upcoming: [{ ...baseComet, id: "c-ics", title: "Dated", due_at: "2099-01-01T10:00:00Z" }],
      undated: [{ ...baseComet, id: "c-nodate", title: "Undated", due_at: null }],
    });
    notesApiMock.downloadCometIcs.mockResolvedValueOnce(
      new Blob(["ics"], { type: "text/calendar" })
    );

    const createObjectURL = vi.fn(() => "blob:ics");
    const revokeObjectURL = vi.fn();
    vi.stubGlobal("URL", Object.assign(URL, { createObjectURL, revokeObjectURL }));

    render(CockpitLeftPanel, { props: {} });

    const btn = await waitFor(() => screen.getByTestId("comet-ics-btn"));
    // only one calendar button — the undated comet has none
    expect(screen.getAllByTestId("comet-ics-btn")).toHaveLength(1);

    await fireEvent.click(btn);
    await waitFor(() => expect(notesApiMock.downloadCometIcs).toHaveBeenCalledWith("c-ics"));
    expect(createObjectURL).toHaveBeenCalled();
  });

  it("does not fetch comets when unauthenticated", async () => {
    authMock.isAuthenticated.mockReturnValue(false);
    notesApiMock.listComets.mockResolvedValue({ overdue: [], upcoming: [], undated: [] });

    render(CockpitLeftPanel, { props: {} });

    await new Promise((r) => setTimeout(r, 10));
    expect(notesApiMock.listComets).not.toHaveBeenCalled();
  });
});
