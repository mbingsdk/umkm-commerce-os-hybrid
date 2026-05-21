import { apiFetch } from "@/lib/api/client";
import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";
import type { CourierZone, CourierZoneInput } from "@/features/courier/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type ApiZone = {
  id: string;
  name: string;
  description?: string;
  rate: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
};

type ApiSuccessResponse<T> = {
  success: true;
  message: string;
  data?: T;
};

export async function listCourierZones(): Promise<CourierZone[]> {
  const result = await apiFetch<ApiZone[]>("/api/v1/courier/zones");
  return result.map(normalizeZone);
}

export async function createCourierZone(input: CourierZoneInput): Promise<CourierZone> {
  const result = await apiFetch<ApiZone>("/api/v1/courier/zones", {
    method: "POST",
    body: JSON.stringify(toZonePayload(input))
  });

  return normalizeZone(result);
}

export async function updateCourierZone(zoneId: string, input: Partial<CourierZoneInput>): Promise<CourierZone> {
  const result = await apiFetch<ApiZone>(`/api/v1/courier/zones/${zoneId}`, {
    method: "PATCH",
    body: JSON.stringify(toZonePayload(input))
  });

  return normalizeZone(result);
}

export async function deleteCourierZone(zoneId: string): Promise<void> {
  await apiFetch(`/api/v1/courier/zones/${zoneId}`, {
    method: "DELETE"
  });
}

export async function listPublicCourierZones(storeSlug: string): Promise<CourierZone[]> {
  const result = await publicApiFetch<ApiZone[]>(
    `/api/v1/public/stores/${encodeURIComponent(storeSlug)}/courier/zones`
  );
  return result.map(normalizeZone);
}

async function publicApiFetch<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    cache: "no-store"
  });
  const body = await parseJSON(response);

  if (!response.ok || isApiErrorResponse(body)) {
    if (isApiErrorResponse(body)) {
      throw toApiError(body, response.status);
    }

    throw new ApiError("Permintaan publik gagal diproses.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isSuccessResponse<T>(body)) {
    throw new ApiError("Respons courier tidak valid.", response.status, "INVALID_API_RESPONSE");
  }

  return body.data as T;
}

async function parseJSON(response: Response) {
  const contentType = response.headers.get("content-type");
  if (!contentType?.includes("application/json")) {
    return null;
  }

  return response.json() as Promise<unknown>;
}

function isSuccessResponse<T>(value: unknown): value is ApiSuccessResponse<T> {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<ApiSuccessResponse<T>>;
  return candidate.success === true && typeof candidate.message === "string";
}

function normalizeZone(zone: ApiZone): CourierZone {
  return {
    id: zone.id,
    name: zone.name,
    description: zone.description,
    rate: zone.rate,
    isActive: zone.is_active,
    sortOrder: zone.sort_order,
    createdAt: zone.created_at,
    updatedAt: zone.updated_at
  };
}

function toZonePayload(input: Partial<CourierZoneInput>) {
  return {
    name: input.name,
    description: input.description ?? "",
    rate: input.rate,
    is_active: input.isActive,
    sort_order: input.sortOrder ?? 0
  };
}
