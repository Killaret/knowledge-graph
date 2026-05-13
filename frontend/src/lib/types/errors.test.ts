// Error types tests
import { describe, it, expect } from 'vitest';
import type { FieldError, ErrorResponse } from './errors';

describe('Error Types', () => {
  describe('FieldError', () => {
    it('should create field error with required properties', () => {
      const fieldError: FieldError = {
        field: 'email',
        reason: 'invalid',
        message: 'Invalid email format'
      };

      expect(fieldError.field).toBe('email');
      expect(fieldError.reason).toBe('invalid');
      expect(fieldError.message).toBe('Invalid email format');
    });

    it('should create field error with optional received property', () => {
      const fieldError: FieldError = {
        field: 'age',
        reason: 'invalid',
        message: 'Age must be positive',
        received: -5
      };

      expect(fieldError.received).toBe(-5);
    });

    it('should create field error with optional expected property', () => {
      const fieldError: FieldError = {
        field: 'password',
        reason: 'too_short',
        message: 'Password too short',
        expected: 'min 8 characters'
      };

      expect(fieldError.expected).toBe('min 8 characters');
    });

    it('should create field error with both optional properties', () => {
      const fieldError: FieldError = {
        field: 'status',
        reason: 'invalid',
        message: 'Invalid status value',
        received: 'pending',
        expected: 'active or inactive'
      };

      expect(fieldError.received).toBe('pending');
      expect(fieldError.expected).toBe('active or inactive');
    });
  });

  describe('ErrorResponse', () => {
    it('should create error response with required properties', () => {
      const errorResponse: ErrorResponse = {
        code: 'VALIDATION_ERROR',
        message: 'Validation failed'
      };

      expect(errorResponse.code).toBe('VALIDATION_ERROR');
      expect(errorResponse.message).toBe('Validation failed');
    });

    it('should create error response with optional details', () => {
      const fieldErrors: FieldError[] = [
        {
          field: 'email',
          reason: 'required',
          message: 'Email is required'
        },
        {
          field: 'password',
          reason: 'too_short',
          message: 'Password must be at least 8 characters',
          expected: 'min 8 characters'
        }
      ];

      const errorResponse: ErrorResponse = {
        code: 'VALIDATION_ERROR',
        message: 'Validation failed',
        details: fieldErrors
      };

      expect(errorResponse.details).toHaveLength(2);
      expect(errorResponse.details?.[0].field).toBe('email');
      expect(errorResponse.details?.[1].field).toBe('password');
    });

    it('should handle empty details array', () => {
      const errorResponse: ErrorResponse = {
        code: 'VALIDATION_ERROR',
        message: 'Validation failed',
        details: []
      };

      expect(errorResponse.details).toHaveLength(0);
    });

    it('should match backend error response format', () => {
      // This test ensures the frontend types match backend Go/Gin error format
      const backendError = {
        code: 'INTERNAL_ERROR',
        message: 'Internal server error',
        details: [
          {
            field: 'database',
            reason: 'connection_failed',
            message: 'Database connection failed',
            received: 'connection timeout',
            expected: 'successful connection'
          }
        ]
      };

      const errorResponse: ErrorResponse = backendError;
      expect(errorResponse.code).toBe('INTERNAL_ERROR');
      expect(errorResponse.message).toBe('Internal server error');
      expect(errorResponse.details?.[0].field).toBe('database');
      expect(errorResponse.details?.[0].reason).toBe('connection_failed');
    });
  });
});
