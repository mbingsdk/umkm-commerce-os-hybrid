"use client";

import { isApiErrorResponse, toApiError } from "@/lib/api/errors";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type ApiFetchOptions = RequestInit & {
  tenantScoped?: boolean;
};

export type ApiSuccessResponse<T, M = unknown> = {
  success: true;
  message: string;
  data?: T;
  meta?: M;
};

export type ApiFetchResult<T, M = unknown> = {
  data: T;
  meta?: M;
};

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const result = await apiFetchWithMeta<T>(path, options);
  return result.data;
}

export async function apiFetchWithMeta<T, M = unknown>(
  path: string,
  options: ApiFetchOptions = {}
): Promise<ApiFetchResult<T, M>> {
  const token = useAuthStore.getState().accessToken;
  const tenantId = useTenantStore.getState().selectedTenantId;
  const headers = new Headers(options.headers);
  const isFormDataBody = typeof FormData !== "undefined" && options.body instanceof FormData;

  if (options.body && !headers.has("Content-Type") && !isFormDataBody) {
    headers.set("Content-Type", "application/json");
  }

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  if (options.tenantScoped !== false && tenantId) {
    headers.set("X-Tenant-ID", tenantId);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers
  });

  const body = await parseJSON(response);

  if (!response.ok || isApiErrorResponse(body)) {
    if (isApiErrorResponse(body)) {
      throw toApiError(body, response.status);
    }

    throw toApiError(
      {
        success: false,
        message: "Permintaan gagal diproses.",
        error: {
          code: "UNKNOWN_API_ERROR"
        }
      },
      response.status
    );
  }

  if (!isApiSuccessResponse<T, M>(body)) {
    throw toApiError(
      {
        success: false,
        message: "Respons API tidak valid.",
        error: {
          code: "INVALID_API_RESPONSE"
        }
      },
      response.status
    );
  }

  return {
    data: body.data as T,
    meta: body.meta as M | undefined
  };
}

async function parseJSON(response: Response) {
  const contentType = response.headers.get("content-type");

  if (!contentType?.includes("application/json")) {
    return null;
  }

  return response.json() as Promise<unknown>;
}

function isApiSuccessResponse<T, M = unknown>(value: unknown): value is ApiSuccessResponse<T, M> {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<ApiSuccessResponse<T>>;
  return candidate.success === true && typeof candidate.message === "string";
}
