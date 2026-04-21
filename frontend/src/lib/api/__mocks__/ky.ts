import { vi } from 'vitest';

 
type KyInstance = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  get: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  post: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  put: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delete: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  extend: any;
};

// Утилита для создания mock response
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const createMockResponse = <T>(data: T, status = 200): any => {
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
// eslint-disable-next-line @typescript-eslint/no-explicit-any
(mockKy as any).create = vi.fn(() => mockKy);
// eslint-disable-next-line @typescript-eslint/no-explicit-any
(mockKy as any).extend = vi.fn(() => mockKy);

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
): void => {
  const response = createMockResponse(data, status);
   
  mockKy[method].mockResolvedValueOnce(response);
};

export default mockKy;
