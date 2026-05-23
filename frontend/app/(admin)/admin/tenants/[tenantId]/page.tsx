"use client";

import { useParams } from "next/navigation";
import { AdminTenantDetailPage } from "@/features/admin/components/admin-tenant-detail-page";

export default function AdminTenantDetailRoute() {
  const params = useParams<{ tenantId: string }>();
  return <AdminTenantDetailPage tenantId={params.tenantId} />;
}
