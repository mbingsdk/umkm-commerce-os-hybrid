export type ShipmentStatus =
  | "pending"
  | "ready_for_pickup"
  | "picked_up"
  | "on_delivery"
  | "delivered"
  | "failed"
  | "cancelled";

export type CourierType = "internal" | "manual";

export type Shipment = {
  id: string;
  orderId: string;
  orderNumber: string;
  courierType: CourierType | string;
  courierName?: string;
  trackingNumber?: string;
  status: ShipmentStatus | string;
  shippingCost: number;
  assignedToName?: string;
  assignedToPhone?: string;
  note?: string;
  shippedAt?: string | null;
  deliveredAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type ShipmentStatusLog = {
  id: string;
  fromStatus?: ShipmentStatus | string | null;
  toStatus: ShipmentStatus | string;
  note?: string;
  createdBy?: string | null;
  createdAt: string;
};

export type ShipmentDetail = {
  shipment: Shipment;
  timeline: ShipmentStatusLog[];
};

export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type ListShipmentsResult = {
  shipments: Shipment[];
  pagination: Pagination;
};

export type ShipmentFilters = {
  query?: string;
  status?: ShipmentStatus | "";
  dateFrom?: string;
  dateTo?: string;
  cursor?: string | null;
  limit?: number;
};

export type CreateShipmentInput = {
  courierType: CourierType;
  courierName?: string;
  trackingNumber?: string;
  shippingCost: number;
  assignedToName?: string;
  assignedToPhone?: string;
  note?: string;
};

export type CreateShipmentResult = {
  id: string;
  orderId: string;
  trackingNumber?: string;
  status: ShipmentStatus | string;
  shippingCost: number;
};

export type UpdateShipmentStatusInput = {
  status: ShipmentStatus;
  note?: string;
};

export type PublicTrackingItem = {
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
};

export type PublicTrackingShipment = {
  courierType: CourierType | string;
  courierName?: string;
  trackingNumber?: string;
  status: ShipmentStatus | string;
  shippingCost: number;
  shippedAt?: string | null;
  deliveredAt?: string | null;
};

export type PublicTrackingTimelineItem = {
  status: ShipmentStatus | string;
  note?: string;
  createdAt: string;
};

export type PublicTrackingResult = {
  orderNumber: string;
  status: string;
  paymentStatus: string;
  shipmentStatus?: string;
  customerName: string;
  items: PublicTrackingItem[];
  totals: {
    subtotal: number;
    shippingCost: number;
    grandTotal: number;
  };
  shipment?: PublicTrackingShipment | null;
  timeline: PublicTrackingTimelineItem[];
};
