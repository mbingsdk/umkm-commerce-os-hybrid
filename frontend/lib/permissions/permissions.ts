export const permissions = {
  tenantRead: "tenant.read",
  tenantCreate: "tenant.create",
  storeRead: "store.read",
  storeUpdate: "store.update",
  storePublish: "store.publish",
  storeUpdateBusinessHours: "store.update_business_hours",
  productRead: "product.read",
  inventoryRead: "inventory.read",
  orderRead: "order.read",
  posCreateTransaction: "pos.create_transaction",
  financeReadSummary: "finance.read_summary"
} as const;

export type Permission = (typeof permissions)[keyof typeof permissions];
