export type ApiErrorPayload = {
  code: string;
  details?: unknown;
};

export class ApiError extends Error {
  status: number;
  code: string;
  details?: unknown;

  constructor(message: string, status: number, code: string, details?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

type RawErrorResponse = {
  success: false;
  message: string;
  error: ApiErrorPayload;
};

export function toApiError(payload: RawErrorResponse, status: number) {
  return new ApiError(payload.message, status, payload.error.code, payload.error.details);
}

export function isApiErrorResponse(value: unknown): value is RawErrorResponse {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<RawErrorResponse>;

  return (
    candidate.success === false &&
    typeof candidate.message === "string" &&
    !!candidate.error &&
    typeof candidate.error.code === "string"
  );
}
