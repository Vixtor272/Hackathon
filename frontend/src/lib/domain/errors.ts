/** Error codes returned by the backend envelope `{ error: { code, message } }`. */
export type ApiErrorCode =
  | 'VALIDATION'
  | 'NOT_FOUND'
  | 'INVALID_STATE'
  | 'INSUFFICIENT_STOCK'
  | 'RX_INCREASE_NOT_ALLOWED'
  | 'RESERVATION_EXPIRED'
  | 'EMPTY_CART'
  | 'PAYMENT_INVALIDATED'
  | 'OCR_ILLEGIBLE'
  | 'PRESCRIPTION_INVALID'
  | 'INTERNAL'
  | 'NETWORK';

/** Error carrying the backend code so the UI can react to it (e.g. show a retry). */
export class ApiError extends Error {
  readonly code: ApiErrorCode | string;
  readonly status: number | undefined;

  constructor(code: ApiErrorCode | string, message: string, status?: number) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

export function isApiError(value: unknown): value is ApiError {
  return value instanceof ApiError;
}

/** Human-readable message for any thrown value, ready to display. */
export function errorMessage(value: unknown): string {
  if (isApiError(value)) return value.message;
  if (value instanceof Error) return value.message;
  return 'Ocurrió un error inesperado';
}
