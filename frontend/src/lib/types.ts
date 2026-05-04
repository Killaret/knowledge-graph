// Global types for the Knowledge Graph application

// Authentication types
export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_at: string;
}

export interface User {
  id: string;
  login: string;
  email?: string;
  role: string;
  created_at: string;
}

export interface LoginRequest {
  login: string;
  password: string;
}

export interface RegisterRequest {
  login: string;
  email?: string;
  password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}

export interface UpdateUserRequest {
  email?: string;
  old_password?: string;
  new_password?: string;
}

// Sharing types
export interface NoteShare {
  id: string;
  note_id: string;
  shared_by_user_id: string;
  shared_with_user_id: string;
  shared_with_login: string;
  permission: 'read' | 'write';
  created_at: string;
  expires_at?: string;
}

export interface ShareLink {
  id: string;
  token: string;
  permission: 'read' | 'write';
  created_at: string;
  expires_at?: string;
  max_uses?: number;
  uses_count: number;
  is_active: boolean;
}

export interface CreateShareRequest {
  user_id: string;
  permission?: 'read' | 'write';
  expires_at?: string;
}

export interface CreateShareLinkRequest {
  permission?: 'read' | 'write';
  expires_at?: string;
  max_uses?: number;
}

// API Key type
export interface APIKey {
  id: string;
  name: string;
  scopes?: string[];
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
}

// Settings types
export interface UserSetting {
  key: string;
  value: unknown;
  updated_at: string;
}

// Achievement types
export interface Achievement {
  id: string;
  code: string;
  title: string;
  description: string;
  icon: string;
  points: number;
  earned: boolean;
  is_hidden: boolean;
}

// Error types
export interface APIError {
  message: string;
  code?: string;
  details?: Record<string, string[]>;
}
