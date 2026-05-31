import type { ReactNode } from "react";
import { PublicMarketingShell } from "@/components/public/public-shell";

export default function MarketingLayout({ children }: { children: ReactNode }) {
  return <PublicMarketingShell>{children}</PublicMarketingShell>;
}
