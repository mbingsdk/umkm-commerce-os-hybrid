"use client";

import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type PublicPaymentConfirmationPayload = {
  customer_phone: string;
  payer_name: string;
  bank_name: string;
  transfer_amount: number;
  transfer_date: string;
  proof_url?: string;
  note?: string;
};

export type PublicPaymentConfirmationResponse = {
  id: string;
  order_id: string;
  order_number: string;
  status: string;
  message: string;
};

type ApiSuccessResponse<T> = {
  success: true;
  message: string;
  data: T;
};

export async function submitPublicPaymentConfirmation(
  storeSlug: string,
  orderNumber: string,
  payload: PublicPaymentConfirmationPayload,
  idempotencyKey: string
): Promise<PublicPaymentConfirmationResponse> {
  const response = await fetch(
    `${API_BASE_URL}/api/v1/public/stores/${encodeURIComponent(storeSlug)}/orders/${encodeURIComponent(
      orderNumber
    )}/payment-confirmation`,
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

    throw new ApiError("Konfirmasi pembayaran gagal dikirim.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isSuccessResponse<PublicPaymentConfirmationResponse>(body)) {
    throw new ApiError("Respons konfirmasi pembayaran tidak valid.", response.status, "INVALID_API_RESPONSE");
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
