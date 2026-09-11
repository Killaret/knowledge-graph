import { describe, expect, it } from "vitest";
import { extractURLs, extractURLsFromHTML, chunk } from "./extract-urls";

describe("extractURLs", () => {
  it("splits concatenated URLs without delimiters", () => {
    const input =
      "https://animego.me/anime/a-1" +
      "https://animego.me/anime/a-2" +
      "https://senkuro.me/manga/x";

    const items = extractURLs(input);

    expect(items).toHaveLength(3);
    expect(items[0].url).toBe("https://animego.me/anime/a-1");
    expect(items[1].url).toBe("https://animego.me/anime/a-2");
    expect(items[2].url).toBe("https://senkuro.me/manga/x");
  });

  it("keeps title | url format", () => {
    const input = "Example page | https://example.com\nhttps://another.example.com";

    const items = extractURLs(input);

    expect(items).toHaveLength(2);
    expect(items[0]).toEqual({ title: "Example page", url: "https://example.com" });
    expect(items[1]).toEqual({ title: "https://another.example.com", url: "https://another.example.com" });
  });

  it("skips empty lines and comments", () => {
    const input = "# comment\n\nhttps://example.com\n";

    const items = extractURLs(input);

    expect(items).toHaveLength(1);
    expect(items[0].url).toBe("https://example.com");
  });

  it("deduplicates identical URLs", () => {
    const input = "https://example.com\nhttps://example.com";

    const items = extractURLs(input);

    expect(items).toHaveLength(1);
  });

  it("truncates long titles and URLs to safe limits", () => {
    const title = "Title".repeat(60); // 300 chars > 200
    const urlPath = "a".repeat(16500);
    const input = `${title} | https://example.com/${urlPath}`;

    const items = extractURLs(input);

    expect(items).toHaveLength(1);
    expect([...items[0].title].length).toBeLessThanOrEqual(200);
    expect([...items[0].url].length).toBeLessThanOrEqual(16384);
    expect(items[0].url.startsWith("https://example.com/")).toBe(true);
  });

  it("strips trailing punctuation glued to the URL", () => {
    const input = "https://example.com/path, https://other.org/;";

    const items = extractURLs(input);

    expect(items).toHaveLength(2);
    expect(items[0].url).toBe("https://example.com/path");
    expect(items[1].url).toBe("https://other.org/");
  });

  it("extracts URLs from free-form text", () => {
    const input = "Check this out: https://example.com and also https://go.dev/doc ok?";

    const items = extractURLs(input);

    expect(items).toHaveLength(2);
    expect(items[0].url).toBe("https://example.com");
    expect(items[1].url).toBe("https://go.dev/doc");
  });
});

describe("extractURLsFromHTML", () => {
  it("extracts bookmarks from a Netscape HTML file", () => {
    const html = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
  <DT><H3>Job</H3>
  <DL><p>
    <DT><A HREF="https://postman.com">Postman</A>
    <DT><A HREF="https://jwt.io/">JSON Web Tokens</A>
    <DT><A HREF="http://localhost:3000/private">private</A>
    <DT><A HREF="https://jwt.io/">duplicate</A>
  </DL><p>
  <DT><A HREF="https://jsonformatter.org/">JSON Formatter</A>
</DL><p>`;

    const items = extractURLsFromHTML(html);

    // extractURLsFromHTML does not enforce URL policy; preview/import will.
    expect(items).toHaveLength(4);
    expect(items[0]).toEqual({ title: "Postman", url: "https://postman.com" });
    expect(items[1]).toEqual({ title: "JSON Web Tokens", url: "https://jwt.io/" });
    expect(items[2]).toEqual({ title: "private", url: "http://localhost:3000/private" });
    expect(items[3]).toEqual({ title: "JSON Formatter", url: "https://jsonformatter.org/" });
  });

  it("strips HTML tags from bookmark titles", () => {
    const html = `<DT><A HREF="https://example.com"><b>Title</b> with <i>tags</i></A>`;

    const items = extractURLsFromHTML(html);

    expect(items).toHaveLength(1);
    expect(items[0].title).toBe("Title with tags");
  });

  it("decodes HTML entities in titles and URLs", () => {
    const html = `<DT><A HREF="https://example.com/path?foo=1&amp;bar=2">Test &amp; More</A>`;

    const items = extractURLsFromHTML(html);

    expect(items).toHaveLength(1);
    expect(items[0].title).toBe("Test & More");
    expect(items[0].url).toBe("https://example.com/path?foo=1&bar=2");
  });
});

describe("chunk", () => {
  it("splits an array into fixed-size chunks", () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });

  it("returns one chunk when size is larger than array", () => {
    expect(chunk([1, 2], 10)).toEqual([[1, 2]]);
  });

  it("returns empty array for non-positive size", () => {
    expect(chunk([1, 2], 0)).toEqual([]);
  });
});
