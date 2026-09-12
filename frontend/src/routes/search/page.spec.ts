import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import { writable } from "svelte/store";
import Page from "./+page.svelte";

const mockGoto = vi.fn();
const mockSearchNotes = vi.fn();
const mockIsAuthenticated = vi.fn();

vi.mock("$app/navigation", () => ({
  goto: (...args: any[]) => mockGoto(...args),
  beforeNavigate: vi.fn(),
  afterNavigate: vi.fn(),
  onNavigate: vi.fn(),
  preloadData: vi.fn(),
  preloadCode: vi.fn(),
}));

vi.mock("$app/stores", () => {
  const page = writable({
    url: new URL("http://localhost/search"),
    params: {},
    route: { id: "/search" },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  });
  return {
    page,
    updated: { subscribe: vi.fn(() => () => {}), check: vi.fn() },
  };
});

vi.mock("$shared/stores/auth.svelte", () => ({
  isAuthenticated: () => mockIsAuthenticated(),
}));

vi.mock("$shared/api/notes", () => ({
  searchNotes: (...args: any[]) => mockSearchNotes(...args),
}));

vi.mock("$shared/utils/i18n", () => ({
  formatMessage: (key: string) => key,
  getCurrentLocale: () => "en",
}));

vi.mock("tippy.js", () => ({
  default: vi.fn(() => ({ show: vi.fn(), hide: vi.fn(), destroy: vi.fn() })),
}));

function createPageStore(q?: string, pageNum?: string) {
  const search = q ? `?q=${q}` : "";
  if (pageNum) {
    const sep = search ? "&" : "?";
    return new URL(`http://localhost/search${search}${sep}page=${pageNum}`);
  }
  return new URL(`http://localhost/search${search}`);
}

describe("Search page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsAuthenticated.mockReturnValue(true);
    mockSearchNotes.mockResolvedValue({ data: [], total: 0, totalPages: 0 });
  });

  it("renders empty state without query", async () => {
    const { page } = await import("$app/stores");
    (page as any).set({
      url: createPageStore(),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("search.emptyTitle")).toBeInTheDocument());
  });

  it("performs a search and displays results", async () => {
    const { page } = await import("$app/stores");
    mockSearchNotes.mockResolvedValue({
      data: [
        {
          id: "n1",
          title: "Found note",
          content: "content",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: { type: "star" },
          type: "star",
        },
      ],
      total: 1,
      totalPages: 1,
    });

    (page as any).set({
      url: createPageStore("test"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("Found note")).toBeInTheDocument());
    expect(screen.getByText(/search\.found/)).toBeInTheDocument();
    expect(mockSearchNotes).toHaveBeenCalledWith("test", 1, 20);
  });

  it("displays no-results state", async () => {
    const { page } = await import("$app/stores");
    mockSearchNotes.mockResolvedValue({ data: [], total: 0, totalPages: 0 });

    (page as any).set({
      url: createPageStore("nothing"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("search.noResultsTitle")).toBeInTheDocument());
  });

  it("displays error state", async () => {
    const { page } = await import("$app/stores");
    mockSearchNotes.mockRejectedValue(new Error("boom"));

    (page as any).set({
      url: createPageStore("bad"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("search.error")).toBeInTheDocument());
  });

  it("handles pagination", async () => {
    const { page } = await import("$app/stores");
    mockSearchNotes.mockResolvedValue({
      data: [
        {
          id: "n1",
          title: "Note",
          content: "c",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: { type: "star" },
          type: "star",
        },
      ],
      total: 3,
      totalPages: 3,
    });

    (page as any).set({
      url: createPageStore("test", "2"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("Note")).toBeInTheDocument());

    const previous = screen.getByText("search.previous");
    fireEvent.click(previous);
    expect(mockGoto).toHaveBeenLastCalledWith("/search?q=test&page=1");

    (page as any).set({
      url: createPageStore("test", "1"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });
    await waitFor(() => expect(mockSearchNotes).toHaveBeenLastCalledWith("test", 1, 20));

    const next = screen.getByText("search.next");
    fireEvent.click(next);
    expect(mockGoto).toHaveBeenLastCalledWith("/search?q=test&page=2");
  });

  it("respects negative or zero page numbers", async () => {
    const { page } = await import("$app/stores");
    (page as any).set({
      url: createPageStore("test", "-1"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(mockSearchNotes).toHaveBeenCalledWith("test", 1, 20));
  });

  it("renders notes in readonly for anonymous users", async () => {
    const { page } = await import("$app/stores");
    mockIsAuthenticated.mockReturnValue(false);
    mockSearchNotes.mockResolvedValue({
      data: [
        {
          id: "n1",
          title: "Note",
          content: "c",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          metadata: { type: "star" },
          type: "star",
        },
      ],
      total: 1,
      totalPages: 1,
    });

    (page as any).set({
      url: createPageStore("test"),
      params: {},
      route: { id: "/search" },
      status: 200,
      error: null,
      data: {},
      form: undefined,
    });

    render(Page);

    await waitFor(() => expect(screen.getByText("Note")).toBeInTheDocument());
  });
});
