// Users API client
import api from './client';
import type { User, UpdateUserRequest, APIKey } from '$lib/types';

/**
 * Get current user profile
 */
export async function getMe(): Promise<User> {
  const response = await api.get('v1/users/me').json<User>();
  return response;
}

/**
 * Update current user profile
 */
export async function updateMe(data: UpdateUserRequest): Promise<User> {
  const response = await api.put('v1/users/me', {
    json: data
  }).json<User>();
  return response;
}

/**
 * Delete current user account
 */
export async function deleteMe(password: string): Promise<void> {
  await api.delete('v1/users/me', {
    json: { password }
  });
}

/**
 * Get user settings
 */
export async function getSettings(): Promise<{ settings: Array<{ key: string; value: unknown; updated_at: string }> }> {
  const response = await api.get('v1/users/me/settings').json<{ settings: Array<{ key: string; value: unknown; updated_at: string }> }>();
  return response;
}

/**
 * Update a setting
 */
export async function updateSetting(key: string, value: unknown): Promise<void> {
  await api.put('v1/users/me/settings', {
    json: { key, value }
  });
}

/**
 * Delete a setting
 */
export async function deleteSetting(key: string): Promise<void> {
  await api.delete(`v1/users/me/settings/${key}`);
}

/**
 * Get galactic mode setting
 */
export async function getGalacticMode(): Promise<{ galactic_mode: boolean }> {
  const response = await api.get('v1/users/me/settings/galactic_mode').json<{ galactic_mode: boolean }>();
  return response;
}

/**
 * Toggle galactic mode
 */
export async function toggleGalacticMode(): Promise<{ message: string; galactic_mode: boolean }> {
  const response = await api.post('v1/users/me/settings/galactic_mode/toggle').json<{ message: string; galactic_mode: boolean }>();
  return response;
}

// API Key management

/**
 * List API keys
 */
export async function listAPIKeys(): Promise<{ api_keys: APIKey[] }> {
  const response = await api.get('v1/users/me/api-keys').json<{ api_keys: APIKey[] }>();
  return response;
}

/**
 * Create new API key
 */
export async function createAPIKey(name: string, scopes?: string[]): Promise<{ id: string; api_key: string; name: string }> {
  const response = await api.post('v1/users/me/api-keys', {
    json: { name, scopes }
  }).json<{ id: string; api_key: string; name: string }>();
  return response;
}

/**
 * Revoke API key
 */
export async function revokeAPIKey(keyId: string): Promise<void> {
  await api.delete(`v1/users/me/api-keys/${keyId}`);
}

// Achievements

/**
 * Get user's achievements
 */
export async function getMyAchievements(): Promise<{ achievements: Array<{ id: string; code: string; title: string; description: string; icon: string; points: number; obtained_at?: string }>; total_points: number }> {
  const response = await api.get('v1/users/me/achievements').json<{
    achievements: Array<{ id: string; code: string; title: string; description: string; icon: string; points: number; obtained_at?: string }>;
    total_points: number;
  }>();
  return response;
}

/**
 * Get all available achievements
 */
export async function getAllAchievements(): Promise<{ achievements: Array<{ id: string; code: string; title: string; description: string; icon: string; points: number; earned: boolean; is_hidden: boolean }> }> {
  const response = await api.get('v1/achievements').json<{
    achievements: Array<{ id: string; code: string; title: string; description: string; icon: string; points: number; earned: boolean; is_hidden: boolean }>;
  }>();
  return response;
}

/**
 * Get login streak
 */
export async function getStreak(): Promise<{ streak: number; next_reward: boolean }> {
  const response = await api.get('v1/users/me/streak').json<{ streak: number; next_reward: boolean }>();
  return response;
}
