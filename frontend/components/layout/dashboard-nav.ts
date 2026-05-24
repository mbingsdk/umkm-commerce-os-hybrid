import { permissions, type Permission } from "@/lib/permissions/permissions";

export type DashboardNavItem = {
  label: string;
  href?: string;
  permission: Permission;
  ready: boolean;
};

export const dashboardNavItems: DashboardNavItem[] = [
  { label: "Dashboard", href: "/dashboard", permission: permissions.dashboardReadSummary, ready: true },
  { label: "Toko", permission: permissions.storeRead, ready: false },
  { label: "Produk", href: "/dashboard/products", permission: permissions.productRead, ready: true },
  { label: "Kategori", href: "/dashboard/categories", permission: permissions.categoryRead, ready: true },
  { label: "Inventori", href: "/dashboard/inventory", permission: permissions.inventoryRead, ready: true },
  { label: "Pesanan", href: "/dashboard/orders", permission: permissions.orderRead, ready: true },
  { label: "POS", href: "/dashboard/pos", permission: permissions.posReadSession, ready: true },
  { label: "Keuangan", href: "/dashboard/finance", permission: permissions.financeReadSummary, ready: true },
  { label: "Kurir", href: "/dashboard/courier/zones", permission: permissions.courierReadZone, ready: true },
  { label: "Pengiriman", href: "/dashboard/shipments", permission: permissions.shipmentRead, ready: true }
];

export function isDashboardNavItemActive(pathname: string, href: string) {
  return pathname === href || (href !== "/dashboard" && pathname.startsWith(`${href}/`));
}
