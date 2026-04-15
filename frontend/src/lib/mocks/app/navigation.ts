import { vi } from 'vitest';

export const goto = vi.fn();
export const beforeNavigate = vi.fn();
export const afterNavigate = vi.fn();
export const onNavigate = vi.fn();
export const preloadData = vi.fn();
export const preloadCode = vi.fn();
