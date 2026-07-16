// Мок для API клиента
import { vi } from 'vitest';

// Создаем мок-функции для HTTP методов
export const mockGet = vi.fn();
export const mockPost = vi.fn();
export const mockPut = vi.fn();
export const mockDelete = vi.fn();
export const mockPatch = vi.fn();

// Фабрика для создания mock response
export const createMockResponse = <T>(data: T, ok = true, status = 200) => ({
  ok,
  status,
  json: vi.fn().mockResolvedValue(data),
  text: vi.fn().mockResolvedValue(JSON.stringify(data)),
  headers: new Headers(),
});

// Базовый мок api клиента
const mockApi = {
  get: mockGet,
  post: mockPost,
  put: mockPut,
  delete: mockDelete,
  patch: mockPatch,
  extend: vi.fn(function(this: any) { return this; }),
};

// Экспортируем api как мок
export const api = mockApi;
export default mockApi;

// Вспомогательные функции для настройки успешных ответов
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

// Сброс всех моков
export const resetApiMocks = () => {
  mockGet.mockReset();
  mockPost.mockReset();
  mockPut.mockReset();
  mockDelete.mockReset();
  mockPatch.mockReset();
};
