import { vi, type Mock } from 'vitest';

type KyInstance = {
  get: Mock<() => Promise<{ json: () => Promise<any> }>>;
  post: Mock<() => Promise<{ json: () => Promise<any> }>>;
  put: Mock<() => Promise<{ json: () => Promise<any> }>>;
  delete: Mock<() => Promise<{ json: () => Promise<any> }>>;
  extend: Mock<() => KyInstance>;
};

// Утилита для создания mock response
export const createMockResponse = <T>(data: T, status = 200) => {
  return {
    json: async () => data,
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error'
  };
};

// Создаем мок экземпляр ky
const createMockKy = (): KyInstance => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  extend: vi.fn(() => createMockKy())
});

const mockKy = createMockKy();

// Добавляем статический метод create как у настоящего ky
(mockKy as any).create = vi.fn(() => mockKy);

// Утилита для сброса всех моков
export const resetKyMocks = () => {
  mockKy.get.mockClear();
  mockKy.post.mockClear();
  mockKy.put.mockClear();
  mockKy.delete.mockClear();
  mockKy.extend.mockClear();
};

// Утилита для настройки mock ответа
export const mockKyResponse = <T>(
  method: keyof Omit<KyInstance, 'extend'>,
  data: T,
  status = 200
) => {
  const response = createMockResponse(data, status);
  mockKy[method].mockResolvedValueOnce(response as any);
};

export default mockKy;
