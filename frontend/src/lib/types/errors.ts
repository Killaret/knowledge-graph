/**
 * Types for API error handling
 * Matches backend ErrorResponse format from Go/Gin
 */

export interface FieldError {
  field: string;
  reason: string;
  message: string;
  received?: any;
  expected?: string;
}

export interface ErrorResponse {
  code: string;
  message: string;
  details?: FieldError[];
}
