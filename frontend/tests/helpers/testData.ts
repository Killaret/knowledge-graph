 import type { APIRequestContext } from "@playwright/test";

// Retry configuration for rate limit handling
const MAX_RETRIES = 3;
const BASE_DELAY_MS = 1000;

/**
 * Sleep utility for delays
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Check if error is a rate limit error (429)
 */
function isRateLimitError(status: number, errorText: string): boolean {
  return status === 429 || errorText.includes("Rate limit exceeded");
}

/**
 * Extract retry_after value from error response (capped for tests)
 */
function getRetryAfter(errorText: string): number {
  try {
    const parsed = JSON.parse(errorText);
    if (parsed.retry_after) {
      // Cap at 5 seconds for tests to prevent long timeouts
      return Math.min(parsed.retry_after * 1000, 5000);
    }
  } catch {
    // If parsing fails, use default
  }
  return BASE_DELAY_MS;
}

export interface NoteData {
  id?: string;
  title: string;
  content?: string;
  type?: string;
  metadata?: Record<string, unknown>;
}

// API Response wrapper structure
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface LinkData {
  id?: string;
  source_note_id: string;
  target_note_id: string;
  weight?: number;
  link_type?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Get backend base URL from environment or default
 */
export function getBackendUrl(): string {
  return process.env.BACKEND_URL || "http://127.0.0.1:9000";
}

/**
 * Create a note via API with retry logic for rate limiting
 */
export async function createNote(
  request: APIRequestContext,
  data: Partial<NoteData>,
  retryCount = 0
): Promise<ApiResponse<NoteData & { id: string; created_at?: string; updated_at?: string }>> {
  const payload = {
    title: data.title || "Test Note",
    content: data.content || "Test content",
    type: data.type,
    metadata: { ...(data.metadata || {}), __testNote: true },
  };

  const response = await request.post(`${getBackendUrl()}/api/v1/notes`, {
    data: payload,
  });

  if (!response.ok()) {
    const errorText = await response.text();
    const status = response.status();

    // Handle rate limiting with retry
    if (isRateLimitError(status, errorText) && retryCount < MAX_RETRIES) {
      const delay = getRetryAfter(errorText) + retryCount * BASE_DELAY_MS;
      console.log(
        `[createNote] Rate limited (attempt ${retryCount + 1}/${MAX_RETRIES}), waiting ${delay}ms before retry...`
      );
      await sleep(delay);
      return createNote(request, data, retryCount + 1);
    }

    throw new Error(`Failed to create note: ${status} - ${errorText}`);
  }

  const result = await response.json();
  console.log("[createNote] API response:", JSON.stringify(result));
  return result;
}

/**
 * Create multiple notes in batch with delay to prevent rate limiting
 */
export async function createNotes(
  request: APIRequestContext,
  notes: Partial<NoteData>[],
  delayMs = 100
): Promise<
  Array<ApiResponse<NoteData & { id: string; created_at?: string; updated_at?: string }>>
> {
  const created = [];
  for (const noteData of notes) {
    const note = await createNote(request, noteData);
    created.push(note);
    // Small delay to prevent rate limiting
    if (delayMs > 0) {
      await sleep(delayMs);
    }
  }
  return created;
}

/**
 * Create a link between two notes with retry logic for rate limiting
 */
export async function createLink(
  request: APIRequestContext,
  sourceId: string,
  targetId: string,
  weight = 0.5,
  linkType = "related",
  retryCount = 0
): Promise<{ id: string; [key: string]: unknown }> {
  // Debug logging
  console.log("[createLink] Creating link:", {
    sourceId,
    targetId,
    weight,
    linkType,
  });

  // Validate inputs
  if (!sourceId || !targetId) {
    throw new Error(`Invalid parameters: sourceId=${sourceId}, targetId=${targetId}`);
  }

  // Go backend expects snake_case field names in JSON (based on struct tags)
  const payload = {
    source_note_id: sourceId,
    target_note_id: targetId,
    weight: weight,
    link_type: linkType,
    metadata: {},
  };

  console.log("[createLink] Payload:", JSON.stringify(payload));

  // Use Playwright request API for proper cookie/auth handling
  const url = `${getBackendUrl()}/api/v1/links`;
  const response = await request.post(url, {
    data: payload,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  });

  if (!response.ok()) {
    const errorText = await response.text();
    const status = response.status();

    // Handle rate limiting with retry
    if (isRateLimitError(status, errorText) && retryCount < MAX_RETRIES) {
      const delay = getRetryAfter(errorText) + retryCount * BASE_DELAY_MS;
      console.log(
        `[createLink] Rate limited (attempt ${retryCount + 1}/${MAX_RETRIES}), waiting ${delay}ms before retry...`
      );
      await sleep(delay);
      return createLink(request, sourceId, targetId, weight, linkType, retryCount + 1);
    }

    throw new Error(`Failed to create link: ${status} - ${errorText}`);
  }

  return await response.json();
}

/**
 * Create a star topology - center note with surrounding notes
 */
export async function createStarTopology(
  request: APIRequestContext,
  centerNote: Partial<NoteData>,
  surroundingNotes: Partial<NoteData>[],
  linkWeight = 0.7
): Promise<{
  center: ApiResponse<NoteData & { id: string; created_at?: string; updated_at?: string }>;
  surrounding: Array<
    ApiResponse<NoteData & { id: string; created_at?: string; updated_at?: string }>
  >;
}> {
  const center = await createNote(request, centerNote);
  const surrounding = [];

  for (const noteData of surroundingNotes) {
    const note = await createNote(request, noteData);
    surrounding.push(note);
    await createLink(request, center.data.id, note.data.id, linkWeight);
  }

  return { center, surrounding };
}

/**
 * Create a chain topology - notes linked in sequence
 */
export async function createChainTopology(
  request: APIRequestContext,
  notes: Partial<NoteData>[],
  linkWeight = 0.8
): Promise<
  Array<ApiResponse<NoteData & { id: string; created_at?: string; updated_at?: string }>>
> {
  const created = [];

  for (let i = 0; i < notes.length; i++) {
    const note = await createNote(request, notes[i]);
    created.push(note);

    // Link to previous note
    if (i > 0) {
      await createLink(request, created[i - 1].data.id, note.data.id, linkWeight);
    }
  }

  return created;
}

/**
 * Delete a note via API
 */
export async function deleteNote(request: APIRequestContext, noteId: string): Promise<void> {
  const response = await request.delete(`${getBackendUrl()}/api/v1/notes/${noteId}`);

  if (!response.ok() && response.status() !== 404) {
    const errorText = await response.text();
    throw new Error(`Failed to delete note: ${response.status()} - ${errorText}`);
  }
}

/**
 * Clean up all test data
 */
export async function cleanupTestData(
  request: APIRequestContext,
  noteIds: string[]
): Promise<void> {
  for (const id of noteIds) {
    try {
      await deleteNote(request, id);
    } catch (e) {
      // Ignore errors during cleanup
      console.warn(`Failed to delete note ${id}:`, e);
    }
  }
}

/**
 * Check if backend is available
 */
export async function isBackendAvailable(request: APIRequestContext): Promise<boolean> {
  try {
    const response = await request.get(`${getBackendUrl()}/api/v1/notes`, {
      timeout: 5000,
    });
    return response.status() < 500;
  } catch {
    return false;
  }
}

/**
 * Test user credentials for E2E tests
 */
export const TEST_USER = {
  login: "e2e_test_user",
  email: "e2e@test.example.com",
  password: "TestPassword123!",
};

/**
 * Register or get existing test user
 */
export async function getOrCreateTestUser(request: APIRequestContext) {
  try {
    // Try to register
    const response = await request.post(`${getBackendUrl()}/api/v1/auth/register`, {
      data: {
        login: TEST_USER.login,
        email: TEST_USER.email,
        password: TEST_USER.password,
      },
    });

    if (response.ok()) {
      const result = await response.json();
      console.log("[TestUser] Registered new user:", TEST_USER.login);
      return { user: result.data, isNew: true };
    }

    // User might already exist
    console.log("[TestUser] Registration returned:", response.status());
    return { user: null, isNew: false };
  } catch (e) {
    console.log("[TestUser] Registration error:", e);
    return { user: null, isNew: false };
  }
}

/**
 * Login with test user and return auth headers
 */
export async function loginAsTestUser(request: APIRequestContext) {
  const response = await request.post(`${getBackendUrl()}/api/v1/auth/login`, {
    data: {
      login: TEST_USER.login,
      password: TEST_USER.password,
    },
  });

  if (!response.ok()) {
    throw new Error(`Login failed: ${response.status()}`);
  }

  const result = await response.json();
  console.log("[TestUser] Login successful, token received");

  return {
    token: result.data.access_token,
    headers: {
      Authorization: `Bearer ${result.data.access_token}`,
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  };
}
