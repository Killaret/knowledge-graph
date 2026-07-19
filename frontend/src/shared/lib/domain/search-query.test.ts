import { describe, it, expect } from "vitest";
import { SearchQuery } from "./search-query";

describe("SearchQuery", () => {
  it("trims whitespace", () => {
    const q = new SearchQuery("  hello world  ");
    expect(q.value).toBe("hello world");
    expect(q.toString()).toBe("hello world");
  });

  it("detects empty query", () => {
    expect(new SearchQuery("").isEmpty()).toBe(true);
    expect(new SearchQuery("   ").isEmpty()).toBe(true);
    expect(new SearchQuery("x").isEmpty()).toBe(false);
  });

  it("validates length", () => {
    expect(new SearchQuery("ab").isValid()).toBe(true);
    expect(new SearchQuery("a").isValid()).toBe(false);
    expect(new SearchQuery("x".repeat(201)).isValid()).toBe(false);
    expect(new SearchQuery("x".repeat(200)).isValid()).toBe(true);
  });

  it("encodes and decodes URL", () => {
    const q = new SearchQuery("hello world");
    expect(q.toURL()).toBe(encodeURIComponent("hello world"));
    expect(SearchQuery.fromURL(q.toURL()).value).toBe("hello world");
  });

  it("gracefully handles invalid URL encoding", () => {
    const q = SearchQuery.fromURL("%ZZ");
    expect(q.value).toBe("%ZZ");
  });

  it("compares equality by trimmed value", () => {
    const a = new SearchQuery(" cosmos ");
    const b = new SearchQuery("cosmos");
    const c = new SearchQuery("galaxy");
    expect(a.equals(b)).toBe(true);
    expect(a.equals(c)).toBe(false);
  });
});
