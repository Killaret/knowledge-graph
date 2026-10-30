import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import { authState } from "./auth-session.svelte";
import { graphView } from "./graph-view.svelte";
import type { User } from "$shared/types";

vi.mock("$app/environment", () => ({ browser: true }));

function createStorage() {
  const data = new Map<string, string>();
  return {
    getItem: (key: string) => data.get(key) ?? null,
    setItem: (key: string, value: string) => {
      data.set(key, value);
    },
    removeItem: (key: string) => {
      data.delete(key);
    },
    clear: () => {
      data.clear();
    },
  };
}

const storage = createStorage();

beforeEach(() => {
  vi.stubGlobal("localStorage", storage);
  storage.clear();
  authState.currentUser = null;
  authState.accessToken = null;
  authState.apiKey = null;
  graphView.clear();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

function setAnonymous() {
  authState.currentUser = null;
  authState.accessToken = null;
  authState.apiKey = null;
}

function setAuthenticated(id: string = "u1") {
  authState.currentUser = { id } as User;
  authState.accessToken = "tok1";
  authState.apiKey = null;
}

describe("graph-view store", () => {
  it("default is community when anonymous", () => {
    setAnonymous();
    expect(graphView.mode).toBe("community");
  });

  it("setter has no effect while anonymous", () => {
    setAnonymous();
    graphView.mode = "personal";
    expect(graphView.mode).toBe("community");
    expect(storage.getItem("graph-view-mode")).toBeNull();
  });

  it("default is personal when authenticated", () => {
    setAuthenticated();
    expect(graphView.mode).toBe("personal");
  });

  it("authenticated switch persists only selected mode", () => {
    setAuthenticated();
    graphView.mode = "community";
    expect(graphView.mode).toBe("community");
    expect(storage.getItem("graph-view-mode")).toBe("community");
    graphView.mode = "personal";
    expect(graphView.mode).toBe("personal");
    expect(storage.getItem("graph-view-mode")).toBe("personal");
  });

  it("restore resets to localStorage value", () => {
    storage.setItem("graph-view-mode", "community");
    setAuthenticated();
    graphView.mode = "personal";
    storage.removeItem("graph-view-mode");
    graphView.restore();
    expect(graphView.mode).toBe("personal");
  });

  it("stored community is restored after in-memory clear", () => {
    setAuthenticated();
    graphView.mode = "community";
    expect(graphView.mode).toBe("community");
    expect(storage.getItem("graph-view-mode")).toBe("community");
    graphView.clear();
    expect(graphView.mode).toBe("personal");
    graphView.restore();
    expect(graphView.mode).toBe("community");
  });

  it("stored personal is ignored while anonymous", () => {
    storage.setItem("graph-view-mode", "personal");
    setAnonymous();
    expect(graphView.mode).toBe("community");
  });

  it("invalid storage value falls back to default", () => {
    storage.setItem("graph-view-mode", "unknown");
    setAuthenticated();
    expect(graphView.mode).toBe("personal");
  });

  it("localStorage errors do not throw", () => {
    const bad = {
      getItem: () => {
        throw new Error("getItem failed");
      },
      setItem: () => {
        throw new Error("setItem failed");
      },
      removeItem: () => {},
    };
    vi.stubGlobal("__SKIP_AUTH__", true);
    vi.stubGlobal("localStorage", bad);
    setAuthenticated();
    expect(() => {
      graphView.mode = "community";
    }).not.toThrow();
    expect(graphView.mode).toBe("community");
  });

  it("scopeKey differs by mode, identity and revision", () => {
    setAnonymous();
    const anonKey = graphView.scopeKey;
    setAuthenticated("u1");
    const authKey = graphView.scopeKey;
    expect(anonKey).toContain("anonymous");
    expect(authKey).toContain("u1");
    expect(anonKey).not.toBe(authKey);
    graphView.mode = "community";
    const commKey = graphView.scopeKey;
    graphView.mode = "personal";
    const personalKey = graphView.scopeKey;
    expect(commKey).not.toBe(personalKey);
  });

  it("another user and logout change scopeKey", () => {
    setAuthenticated("u1");
    const first = graphView.scopeKey;
    authState.currentUser = { id: "u2" } as User;
    const second = graphView.scopeKey;
    setAnonymous();
    const third = graphView.scopeKey;
    expect(second).not.toBe(first);
    expect(third).not.toBe(second);
  });

  it("A->B->A round-trip changes the final scopeKey", () => {
    setAuthenticated();
    const initial = graphView.scopeKey;
    graphView.mode = "community";
    const middle = graphView.scopeKey;
    graphView.mode = "personal";
    const final = graphView.scopeKey;
    expect(middle).not.toBe(initial);
    expect(final).not.toBe(initial);
  });
});
