export const permissions = {
  tenantRead: "tenant.read",
  tenantCreate: "tenant.create",
  storeRead: "store.read",
  storeUpdate: "store.update",
  storePublish: "store.publish",
  storeUpdateBusinessHours: "store.update_business_hours",
  categoryRead: "category.read",
  categoryCreate: "category.create",
  categoryUpdate: "category.update",
  categoryDelete: "category.delete",
  productRead: "product.read",
  productCreate: "product.create",
  productUpdate: "product.update",
  productDelete: "product.delete",
  productUploadImage: "product.upload_image",
  inventoryRead: "inventory.read",
  orderRead: "order.read",
  orderReadDetail: "order.read_detail",
  orderUpdateStatus: "order.update_status",
  orderUpdatePaymentStatus: "order.update_payment_status",
  orderCancel: "order.cancel",
  posCreateTransaction: "pos.create_transaction",
  financeReadSummary: "finance.read_summary"
} as const;

export type Permission = (typeof permissions)[keyof typeof permissions];
