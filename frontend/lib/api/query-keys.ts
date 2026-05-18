export const queryKeys = {
  me: ["me"] as const,
  tenants: ["tenants"] as const,
  currentStore: (tenantId: string | null) => ["current-store", tenantId] as const
};
