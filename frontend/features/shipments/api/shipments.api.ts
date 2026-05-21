import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";
import type {
  CreateShipmentInput,
  CreateShipmentResult,
  ListShipmentsResult,
  Pagination,
  PublicTrackingResult,
  Shipment,
  ShipmentDetail,
  ShipmentFilters,
  ShipmentStatusLog,
  UpdateShipmentStatusInput
} from "@/features/shipments/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiShipment = {
  id: string;
  order_id: string;
  order_number: string;
  courier_type: string;
  courier_name?: string;
  tracking_number?: string;
  status: string;
  shipping_cost: number;
  assigned_to_name?: string;
  assigned_to_phone?: string;
  note?: string;
  shipped_at?: string | null;
  delivered_at?: string | null;
  created_at: string;
  updated_at: string;
};

type ApiStatusLog = {
  id: string;
  from_status?: string | null;
  to_status: string;
  note?: string;
  created_by?: string | null;
  created_at: string;
};

type ApiShipmentDetail = {
  shipment: ApiShipment;
  timeline: ApiStatusLog[];
};

type ApiCreateShipment = {
  id: string;
  order_id: string;
  tracking_number?: string;
  status: string;
  shipping_cost: number;
};

type ApiPublicTrackingResult = {
  order_number: string;
  status: string;
  payment_status: string;
  shipment_status?: string;
  customer_name: string;
  items: Array<{
    product_name: string;
    quantity: number;
    unit_price: number;
    subtotal: number;
  }>;
  totals: {
    subtotal: number;
    shipping_cost: number;
    grand_total: number;
  };
  shipment?: {
    courier_type: string;
    courier_name?: string;
    tracking_number?: string;
    status: string;
    shipping_cost: number;
    shipped_at?: string | null;
    delivered_at?: string | null;
  } | null;
  timeline: Array<{
    status: string;
    note?: string;
    created_at: string;
  }>;
};

type ApiSuccessResponse<T> = {
  success: true;
  message: string;
  data?: T;
};

export async function listShipments(filters: ShipmentFilters = {}): Promise<ListShipmentsResult> {
  const params = toSearchParams(filters);
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  const result = await apiFetchWithMeta<ApiShipment[], ApiPaginationMeta>(`/api/v1/shipments${suffix}`);

  return {
    shipments: result.data.map(normalizeShipment),
    pagination: normalizePagination(result.meta)
  };
}

export async function getShipmentDetail(shipmentId: string): Promise<ShipmentDetail> {
  const result = await apiFetch<ApiShipmentDetail>(`/api/v1/shipments/${shipmentId}`);

  return {
    shipment: normalizeShipment(result.shipment),
    timeline: result.timeline.map(normalizeStatusLog)
  };
}

export async function createShipment(orderId: string, input: CreateShipmentInput): Promise<CreateShipmentResult> {
  const result = await apiFetch<ApiCreateShipment>(`/api/v1/orders/${orderId}/shipments`, {
    method: "POST",
    body: JSON.stringify({
      courier_type: input.courierType,
      courier_name: input.courierName ?? "",
      tracking_number: input.trackingNumber ?? "",
      shipping_cost: input.shippingCost,
      assigned_to_name: input.assignedToName ?? "",
      assigned_to_phone: input.assignedToPhone ?? "",
      note: input.note ?? ""
    })
  });

  return {
    id: result.id,
    orderId: result.order_id,
    trackingNumber: result.tracking_number,
    status: result.status,
    shippingCost: result.shipping_cost
  };
}

export async function updateShipmentStatus(
  shipmentId: string,
  input: UpdateShipmentStatusInput
): Promise<{ id: string; status: string }> {
  return apiFetch(`/api/v1/shipments/${shipmentId}/status`, {
    method: "PATCH",
    body: JSON.stringify({
      status: input.status,
      note: input.note ?? ""
    })
  });
}

export async function getPublicOrderTracking(
  storeSlug: string,
  orderNumber: string,
  phone: string
): Promise<PublicTrackingResult> {
  const params = new URLSearchParams({ phone });
  const result = await publicApiFetch<ApiPublicTrackingResult>(
    `/api/v1/public/stores/${encodeURIComponent(storeSlug)}/orders/${encodeURIComponent(orderNumber)}/tracking?${params.toString()}`
  );

  return normalizePublicTracking(result);
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

    throw new ApiError("Tracking pesanan gagal dimuat.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isSuccessResponse<T>(body)) {
    throw new ApiError("Respons tracking tidak valid.", response.status, "INVALID_API_RESPONSE");
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

function toSearchParams(filters: ShipmentFilters) {
  const params = new URLSearchParams();

  if (filters.query) {
    params.set("q", filters.query);
  }
  if (filters.status) {
    params.set("status", filters.status);
  }
  if (filters.dateFrom) {
    params.set("date_from", filters.dateFrom);
  }
  if (filters.dateTo) {
    params.set("date_to", filters.dateTo);
  }
  if (filters.cursor) {
    params.set("cursor", filters.cursor);
  }
  if (filters.limit) {
    params.set("limit", String(filters.limit));
  }

  return params;
}

function normalizeShipment(shipment: ApiShipment): Shipment {
  return {
    id: shipment.id,
    orderId: shipment.order_id,
    orderNumber: shipment.order_number,
    courierType: shipment.courier_type,
    courierName: shipment.courier_name,
    trackingNumber: shipment.tracking_number,
    status: shipment.status,
    shippingCost: shipment.shipping_cost,
    assignedToName: shipment.assigned_to_name,
    assignedToPhone: shipment.assigned_to_phone,
    note: shipment.note,
    shippedAt: shipment.shipped_at,
    deliveredAt: shipment.delivered_at,
    createdAt: shipment.created_at,
    updatedAt: shipment.updated_at
  };
}

function normalizeStatusLog(log: ApiStatusLog): ShipmentStatusLog {
  return {
    id: log.id,
    fromStatus: log.from_status,
    toStatus: log.to_status,
    note: log.note,
    createdBy: log.created_by,
    createdAt: log.created_at
  };
}

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}

function normalizePublicTracking(result: ApiPublicTrackingResult): PublicTrackingResult {
  return {
    orderNumber: result.order_number,
    status: result.status,
    paymentStatus: result.payment_status,
    shipmentStatus: result.shipment_status,
    customerName: result.customer_name,
    items: result.items.map((item) => ({
      productName: item.product_name,
      quantity: item.quantity,
      unitPrice: item.unit_price,
      subtotal: item.subtotal
    })),
    totals: {
      subtotal: result.totals.subtotal,
      shippingCost: result.totals.shipping_cost,
      grandTotal: result.totals.grand_total
    },
    shipment: result.shipment
      ? {
          courierType: result.shipment.courier_type,
          courierName: result.shipment.courier_name,
          trackingNumber: result.shipment.tracking_number,
          status: result.shipment.status,
          shippingCost: result.shipment.shipping_cost,
          shippedAt: result.shipment.shipped_at,
          deliveredAt: result.shipment.delivered_at
        }
      : null,
    timeline: result.timeline.map((item) => ({
      status: item.status,
      note: item.note,
      createdAt: item.created_at
    }))
  };
}
