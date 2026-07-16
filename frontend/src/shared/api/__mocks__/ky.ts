import { vi } from 'vitest';
import type { MockedFunction } from 'vitest';

// Тип для mock response с методом json()
interface MockResponse<T = unknown> {
  ok: boolean;
  status: number;
  json: () => Promise<T>;
  text: () => Promise<string>;
  headers: Headers;
}

// Фабрика для создания mock response
export function createMockResponse<T>(data: T, ok = true, status = 200): MockResponse<T> {
  return {
    ok,
    status,
    json: vi.fn().mockResolvedValue(data),
    text: vi.fn().mockResolvedValue(JSON.stringify(data)),
    headers: new Headers(),
  };
}

// Создаем моки для методов HTTP с правильными типами
const mockGet = vi.fn() as MockedFunction<() => Promise<MockResponse>>;
const mockPost = vi.fn() as MockedFunction<() => Promise<MockResponse>>;
const mockPut = vi.fn() as MockedFunction<() => Promise<MockResponse>>;
const mockDelete = vi.fn() as MockedFunction<() => Promise<MockResponse>>;
const mockPatch = vi.fn() as MockedFunction<() => Promise<MockResponse>>;

// Мок ky с методом create и extend
const mockKy = {
  create: vi.fn(() => ({
    get: mockGet,
    post: mockPost,
    put: mockPut,
    delete: mockDelete,
    patch: mockPatch,
    extend: vi.fn(function(this: any) { return this; })
  })),
  // Для прямого использования (также поддерживает extend)
  extend: vi.fn(function(this: typeof mockKy) { return this; }),
  get: mockGet,
  post: mockPost,
  put: mockPut,
  delete: mockDelete,
  patch: mockPatch
};

// Экспортируем функции для удобной настройки моков в тестах
export const mockGetSuccess = <T>(data: T) => {
  mockGet.mockResolvedValue(createMockResponse(data));
};

export const mockPostSuccess = <T>(data: T) => {
  mockPost.mockResolvedValue(createMockResponse(data, true, 201));
};

export const mockPutSuccess = <T>(data: T) => {
  mockPut.mockResolvedValue(createMockResponse(data));
};

export const mockDeleteSuccess = <T>(data?: T) => {
  mockDelete.mockResolvedValue(createMockResponse(data ?? {}, true, 204));
};

export const resetApiMocks = () => {
  mockGet.mockReset();
  mockPost.mockReset();
  mockPut.mockReset();
  mockDelete.mockReset();
  mockPatch.mockReset();
};

export default mockKy;
export { mockGet, mockPost, mockPut, mockDelete, mockPatch };
