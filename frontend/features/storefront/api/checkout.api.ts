"use client";

import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type CheckoutPayload = {
  items: Array<{
    product_id: string;
    quantity: number;
  }>;
  customer: {
    name: string;
    phone: string;
    email?: string;
  };
  shipping_address: {
    recipient_name?: string;
    recipient_phone?: string;
    address: string;
    city?: string;
    province?: string;
    postal_code?: string;
  };
  payment_method: "manual_transfer";
  customer_note?: string;
};

export type CheckoutResponse = {
  order_id: string;
  order_number: string;
  status: string;
  payment_status: string;
  totals: {
    subtotal: number;
    discount_total: number;
    shipping_cost: number;
    tax_total: number;
    grand_total: number;
  };
  payment_instruction: {
    method: string;
    message: string;
  };
};

type ApiSuccessResponse<T> = {
  success: true;
  message: string;
  data: T;
};

export async function checkoutPublicStore(
  storeSlug: string,
  payload: CheckoutPayload,
  idempotencyKey: string
): Promise<CheckoutResponse> {
  const response = await fetch(
    `${API_BASE_URL}/api/v1/public/stores/${encodeURIComponent(storeSlug)}/checkout`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": idempotencyKey
      },
      body: JSON.stringify(payload)
    }
  );

  const body = await parseJSON(response);

  if (!response.ok || isApiErrorResponse(body)) {
    if (isApiErrorResponse(body)) {
      throw toApiError(body, response.status);
    }

    throw new ApiError("Checkout gagal diproses.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isSuccessResponse<CheckoutResponse>(body)) {
    throw new ApiError("Respons checkout tidak valid.", response.status, "INVALID_API_RESPONSE");
  }

  return body.data;
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
