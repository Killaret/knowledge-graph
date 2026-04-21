import { MockedFunction } from 'vitest';

type MockResponse = { json: () => Promise<any> };

declare const mockGet: MockedFunction<(...args: any[]) => Promise<MockResponse>>;
declare const mockPost: MockedFunction<(...args: any[]) => Promise<MockResponse>>;
declare const mockPut: MockedFunction<(...args: any[]) => Promise<MockResponse>>;
declare const mockDelete: MockedFunction<(...args: any[]) => Promise<MockResponse>>;

export { mockGet, mockPost, mockPut, mockDelete };

interface KyMock {
  create: () => {
    get: typeof mockGet;
    post: typeof mockPost;
    put: typeof mockPut;
    delete: typeof mockDelete;
    extend: () => KyMock['create'];
  };
  get: typeof mockGet;
  post: typeof mockPost;
  put: typeof mockPut;
  delete: typeof mockDelete;
}

declare const ky: KyMock;
export default ky;
