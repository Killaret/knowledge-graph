// Test utils tests
import { describe, it, expect } from 'vitest';
import { createApiResponse, sleep } from './index';

describe('test-utils', () => {
  describe('createApiResponse', () => {
    it('should create response with default options', () => {
      const data = { message: 'success' };
      const response = createApiResponse(data);
      
      expect(response).toBeInstanceOf(Response);
      expect(response.status).toBe(200);
      expect(response.headers.get('content-type')).toBe('text/plain;charset=UTF-8');
    });

    it('should create response with custom status', () => {
      const data = { error: 'not found' };
      const response = createApiResponse(data, { status: 404 });
      
      expect(response.status).toBe(404);
    });

    it('should create response with custom headers', () => {
      const data = { message: 'success' };
      const response = createApiResponse(data, { 
        headers: { 'x-custom': 'value' } 
      });
      
      expect(response.headers.get('x-custom')).toBe('value');
    });

    it('should create response with both custom status and headers', () => {
      const data = { message: 'created' };
      const response = createApiResponse(data, { 
        status: 201,
        headers: { 'location': '/new-resource' } 
      });
      
      expect(response.status).toBe(201);
      expect(response.headers.get('location')).toBe('/new-resource');
    });
  });

  describe('sleep', () => {
    it('should resolve after specified time', async () => {
      const start = Date.now();
      await sleep(100);
      const end = Date.now();
      
      expect(end - start).toBeGreaterThanOrEqual(90); // Allow some tolerance
    });

    it('should return void', async () => {
      const result = await sleep(50);
      expect(result).toBeUndefined();
    });
  });
});
