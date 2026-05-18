"use client";

import type { ReactNode } from "react";
import type { Permission } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

type PermissionGateProps = {
  permission: Permission;
  children: ReactNode;
  fallback?: ReactNode;
};

export function PermissionGate({ permission, children, fallback = null }: PermissionGateProps) {
  const allowed = useTenantStore((state) => state.permissions.includes(permission));

  return allowed ? children : fallback;
}
