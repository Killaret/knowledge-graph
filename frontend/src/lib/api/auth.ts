// Authentication API client
import api from './client';
import type { AuthTokens, LoginRequest, RegisterRequest, ForgotPasswordRequest, ResetPasswordRequest } from '$lib/types';

/**
 * Login with credentials
 */
export async function login(login: string, password: string): Promise<AuthTokens> {
  const response = await api.post('v1/auth/login', {
    json: { login, password } as LoginRequest
  }).json<AuthTokens>();
  return response;
}

/**
 * Register new user
 */
export async function register(login: string, password: string, email?: string): Promise<AuthTokens> {
  const body: RegisterRequest & { email?: string } = { login, password };
  if (email) {
    body.email = email;
  }
  
  const response = await api.post('v1/auth/register', {
    json: body
  }).json<AuthTokens>();
  return response;
}

/**
 * Refresh access token
 */
export async function refreshTokens(refreshToken: string): Promise<AuthTokens> {
  const response = await api.post('v1/auth/refresh', {
    json: { refresh_token: refreshToken }
  }).json<AuthTokens>();
  return response;
}

/**
 * Logout user
 */
export async function logout(refreshToken: string): Promise<void> {
  await api.post('v1/auth/logout', {
    json: { refresh_token: refreshToken },
    headers: {
      'X-Refresh-Token': refreshToken
    }
  });
}

/**
 * Request password reset
 */
export async function forgotPassword(email: string): Promise<void> {
  await api.post('v1/auth/forgot-password', {
    json: { email } as ForgotPasswordRequest
  });
}

/**
 * Reset password with token
 */
export async function resetPassword(token: string, newPassword: string): Promise<void> {
  await api.post('v1/auth/reset-password', {
    json: { token, new_password: newPassword } as ResetPasswordRequest
  });
}

/**
 * Get Yandex OAuth login URL
 */
export async function getYandexLoginUrl(): Promise<{ url: string }> {
  const response = await api.get('v1/auth/yandex/login').json<{ url: string }>();
  return response;
}

/**
 * Handle Yandex OAuth callback
 */
export async function handleYandexCallback(code: string, state: string): Promise<AuthTokens> {
  const response = await api.get('v1/auth/yandex/callback', {
    searchParams: { code, state }
  }).json<AuthTokens>();
  return response;
}
