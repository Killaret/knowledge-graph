import { describe, it, expect } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import {
  shareNote,
  createShareLink,
  getNoteShares,
  revokeShare,
  revokeShareLink,
  accessSharedNote,
  searchUsers,
} from "./sharing";
import type { NoteShare, ShareLink } from "$shared/types";

const baseUrl = "http://localhost:8080/api";

describe("sharing API", () => {
  const mockShare: NoteShare = {
    id: "s1",
    note_id: "n1",
    shared_by_user_id: "u1",
    shared_with_user_id: "u2",
    shared_with_login: "friend",
    permission: "read",
    created_at: "2024-01-01T00:00:00Z",
  };

  const mockLink: ShareLink = {
    id: "l1",
    token: "abc",
    permission: "read",
    created_at: "2024-01-01T00:00:00Z",
    uses_count: 0,
    is_active: true,
  };

  it("shareNote creates a share", async () => {
    server.use(
      http.post(`${baseUrl}/v1/notes/n1/share`, () => HttpResponse.json(mockShare, { status: 201 }))
    );
    expect(await shareNote("n1", "u2", "read")).toEqual(mockShare);
  });

  it("createShareLink creates a share link", async () => {
    server.use(
      http.post(`${baseUrl}/v1/notes/n1/share-link`, () =>
        HttpResponse.json(mockLink, { status: 201 })
      )
    );
    expect(await createShareLink("n1", "write")).toEqual(mockLink);
  });

  it("getNoteShares returns shares and links", async () => {
    const response = { user_shares: [mockShare], share_links: [mockLink] };
    server.use(http.get(`${baseUrl}/v1/notes/n1/shares`, () => HttpResponse.json(response)));
    expect(await getNoteShares("n1")).toEqual(response);
  });

  it("revokeShare removes a share", async () => {
    server.use(
      http.delete(`${baseUrl}/v1/notes/n1/shares/s1`, () => HttpResponse.json({}, { status: 204 }))
    );
    await expect(revokeShare("n1", "s1")).resolves.toBeUndefined();
  });

  it("revokeShareLink removes a share link", async () => {
    server.use(
      http.delete(`${baseUrl}/v1/share-links/l1`, () => HttpResponse.json({}, { status: 204 }))
    );
    await expect(revokeShareLink("l1")).resolves.toBeUndefined();
  });

  it("accessSharedNote returns shared note", async () => {
    const response = {
      note: {
        id: "n1",
        title: "Shared",
        content: "",
        type: "star",
        metadata: {},
        created_at: "",
        updated_at: "",
      },
      permission: "read",
    };
    server.use(http.get(`${baseUrl}/v1/share/abc`, () => HttpResponse.json(response)));
    expect(await accessSharedNote("abc")).toEqual(response);
  });

  it("searchUsers returns matching users", async () => {
    const response = { users: [{ id: "u2", login: "friend" }] };
    server.use(
      http.get(`${baseUrl}/v1/users`, ({ request }) => {
        const url = new URL(request.url);
        expect(url.searchParams.get("search")).toBe("friend");
        return HttpResponse.json(response);
      })
    );
    expect(await searchUsers("friend")).toEqual(response);
  });
});
