import { describe, it, expect } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../../../vitest-setup";
import {
  getMe,
  updateMe,
  deleteMe,
  getSettings,
  updateSetting,
  deleteSetting,
  getGalacticMode,
  toggleGalacticMode,
  listAPIKeys,
  createAPIKey,
  revokeAPIKey,
  getMyAchievements,
  getAllAchievements,
  getStreak,
} from "./users";
import type { User, APIKey, UpdateUserRequest } from "$shared/types";

const baseUrl = "http://localhost:8080/api";

describe("users API", () => {
  const mockUser: User = {
    id: "u1",
    login: "tester",
    email: "tester@example.com",
    role: "user",
    created_at: "2024-01-01T00:00:00Z",
  };

  it("getMe returns current user", async () => {
    server.use(http.get(`${baseUrl}/v1/users/me`, () => HttpResponse.json(mockUser)));
    expect(await getMe()).toEqual(mockUser);
  });

  it("updateMe updates and returns user", async () => {
    const payload: UpdateUserRequest = { email: "new@example.com" };
    server.use(http.put(`${baseUrl}/v1/users/me`, () => HttpResponse.json(mockUser)));
    expect(await updateMe(payload)).toEqual(mockUser);
  });

  it("deleteMe sends delete request", async () => {
    server.use(http.delete(`${baseUrl}/v1/users/me`, () => HttpResponse.json({}, { status: 204 })));
    await expect(deleteMe("secret")).resolves.toBeUndefined();
  });

  it("getSettings returns user settings", async () => {
    const response = {
      settings: [{ key: "theme", value: "dark", updated_at: "" }],
    };
    server.use(http.get(`${baseUrl}/v1/users/me/settings`, () => HttpResponse.json(response)));
    expect(await getSettings()).toEqual(response);
  });

  it("updateSetting sends put request", async () => {
    server.use(
      http.put(`${baseUrl}/v1/users/me/settings`, () => HttpResponse.json({}, { status: 204 }))
    );
    await expect(updateSetting("theme", "light")).resolves.toBeUndefined();
  });

  it("deleteSetting removes a setting", async () => {
    server.use(
      http.delete(`${baseUrl}/v1/users/me/settings/theme`, () =>
        HttpResponse.json({}, { status: 204 })
      )
    );
    await expect(deleteSetting("theme")).resolves.toBeUndefined();
  });

  it("getGalacticMode returns galactic mode setting", async () => {
    const response = { galactic_mode: true };
    server.use(
      http.get(`${baseUrl}/v1/users/me/settings/galactic_mode`, () => HttpResponse.json(response))
    );
    expect(await getGalacticMode()).toEqual(response);
  });

  it("toggleGalacticMode toggles galactic mode", async () => {
    const response = { message: "ok", galactic_mode: false };
    server.use(
      http.post(`${baseUrl}/v1/users/me/settings/galactic_mode/toggle`, () =>
        HttpResponse.json(response)
      )
    );
    expect(await toggleGalacticMode()).toEqual(response);
  });

  it("listAPIKeys returns api keys", async () => {
    const mockKey: APIKey = {
      id: "k1",
      name: "dev",
      scopes: ["read"],
      created_at: "",
    };
    server.use(
      http.get(`${baseUrl}/v1/users/me/api-keys`, () => HttpResponse.json({ api_keys: [mockKey] }))
    );
    const result = await listAPIKeys();
    expect(result.api_keys).toHaveLength(1);
    expect(result.api_keys[0].name).toBe("dev");
  });

  it("createAPIKey creates and returns key", async () => {
    const response = { id: "k1", api_key: "secret", name: "dev" };
    server.use(
      http.post(`${baseUrl}/v1/users/me/api-keys`, () =>
        HttpResponse.json(response, { status: 201 })
      )
    );
    expect(await createAPIKey("dev", ["read"])).toEqual(response);
  });

  it("revokeAPIKey deletes key", async () => {
    server.use(
      http.delete(`${baseUrl}/v1/users/me/api-keys/k1`, () =>
        HttpResponse.json({}, { status: 204 })
      )
    );
    await expect(revokeAPIKey("k1")).resolves.toBeUndefined();
  });

  it("getMyAchievements returns achievements", async () => {
    const response = {
      achievements: [
        {
          id: "a1",
          code: "first",
          title: "First",
          description: "",
          icon: "",
          points: 10,
        },
      ],
      total_points: 10,
    };
    server.use(http.get(`${baseUrl}/v1/users/me/achievements`, () => HttpResponse.json(response)));
    expect(await getMyAchievements()).toEqual(response);
  });

  it("getAllAchievements returns all achievements", async () => {
    const response = {
      achievements: [
        {
          id: "a1",
          code: "first",
          title: "First",
          description: "",
          icon: "",
          points: 10,
          earned: true,
          is_hidden: false,
        },
      ],
    };
    server.use(http.get(`${baseUrl}/v1/achievements`, () => HttpResponse.json(response)));
    expect(await getAllAchievements()).toEqual(response);
  });

  it("getStreak returns login streak", async () => {
    const response = { streak: 5, next_reward: true };
    server.use(http.get(`${baseUrl}/v1/users/me/streak`, () => HttpResponse.json(response)));
    expect(await getStreak()).toEqual(response);
  });
});
