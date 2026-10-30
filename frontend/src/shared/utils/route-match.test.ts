import { describe, it, expect } from "vitest";
import { isPublicRoute } from "./route-match";

const publicRoutes = ["/", "/graph", "/auth/login", "/auth/yandex/callback"];

describe("isPublicRoute", () => {
  it("matches exact public routes", () => {
    expect(isPublicRoute("/", publicRoutes)).toBe(true);
    expect(isPublicRoute("/graph", publicRoutes)).toBe(true);
    expect(isPublicRoute("/auth/login", publicRoutes)).toBe(true);
    expect(isPublicRoute("/auth/yandex/callback", publicRoutes)).toBe(true);
  });

  it("matches public sub-routes", () => {
    expect(isPublicRoute("/graph/3d", publicRoutes)).toBe(true);
    expect(isPublicRoute("/graph/123", publicRoutes)).toBe(true);
  });

  it("does NOT treat every path as public because of '/'", () => {
    expect(isPublicRoute("/notes", publicRoutes)).toBe(false);
    expect(isPublicRoute("/notes/new", publicRoutes)).toBe(false);
    expect(isPublicRoute("/search", publicRoutes)).toBe(false);
    expect(isPublicRoute("/profile", publicRoutes)).toBe(false);
    expect(isPublicRoute("/import", publicRoutes)).toBe(false);
  });

  it("does NOT match partial path segments", () => {
    expect(isPublicRoute("/auth", publicRoutes)).toBe(false);
    expect(isPublicRoute("/auth/register-now", publicRoutes)).toBe(false);
    expect(isPublicRoute("/graphite", publicRoutes)).toBe(false);
  });
});
