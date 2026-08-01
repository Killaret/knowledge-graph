// Tests for the shared API error message helper
import { describe, it, expect } from "vitest";
import { HTTPError } from "ky";
import { getApiErrorMessage } from "./errorMessage";

function makeHTTPError(status: number, body: unknown): HTTPError {
  const request = new Request("https://example.com/api/v1/users/me", { method: "PUT" });
  const response = new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
  return new HTTPError(response, request, {
    method: "put",
    // ky's Options type has many required-looking fields at runtime, but the
    // helper only ever reads `error.response`, so a minimal cast is safe here.
  } as unknown as ConstructorParameters<typeof HTTPError>[2]);
}

describe("getApiErrorMessage", () => {
  it("maps 'email already in use' to a localized message", async () => {
    const error = makeHTTPError(409, { error: "email already in use" });
    expect(await getApiErrorMessage(error, "en")).toBe(
      "This email is already in use by another account."
    );
    expect(await getApiErrorMessage(error, "ru")).toBe(
      "Этот email уже используется другим аккаунтом."
    );
  });

  it("maps 'invalid old password' to a localized message", async () => {
    const error = makeHTTPError(401, { error: "invalid old password" });
    expect(await getApiErrorMessage(error, "en")).toBe("Current password is incorrect.");
  });

  it("maps 'old password is required' to a localized message", async () => {
    const error = makeHTTPError(400, { error: "old password is required" });
    expect(await getApiErrorMessage(error, "en")).toBe("Please enter your current password.");
  });

  it("maps 'invalid password' (delete account) to a localized message", async () => {
    const error = makeHTTPError(401, { error: "invalid password" });
    expect(await getApiErrorMessage(error, "en")).toBe("Incorrect password.");
  });

  it("maps gin email validator errors to a friendly message", async () => {
    const error = makeHTTPError(400, {
      error: "Key: 'UpdateUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
    });
    expect(await getApiErrorMessage(error, "en")).toBe("Please enter a valid email address.");
  });

  it("falls back to the raw backend message for unknown errors", async () => {
    const error = makeHTTPError(500, { error: "something unexpected happened" });
    expect(await getApiErrorMessage(error, "en")).toBe("something unexpected happened");
  });

  it("falls back to a generic server error when the body isn't JSON", async () => {
    const request = new Request("https://example.com/api/v1/users/me", { method: "PUT" });
    const response = new Response("not json", { status: 500 });
    const error = new HTTPError(
      response,
      request,
      {} as unknown as ConstructorParameters<typeof HTTPError>[2]
    );
    expect(await getApiErrorMessage(error, "en")).toBe("Server error. Please try again later.");
  });

  it("falls back to a generic server error for non-HTTPError values", async () => {
    expect(await getApiErrorMessage(new Error("network down"), "en")).toBe(
      "Server error. Please try again later."
    );
    expect(await getApiErrorMessage("plain string", "ru")).toBe(
      "Ошибка сервера. Попробуйте позже."
    );
  });
});
