import Link from "next/link";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils/cn";

export const publicTheme = {
  bg: "bg-[#F8F1E7]",
  surface: "bg-[#FFFDF8]",
  surfaceMuted: "bg-[#F1E7D8]",
  text: "text-[#251F1A]",
  textMuted: "text-[#6F6256]",
  border: "border-[#E3D2BC]",
  accent: "text-[#B96E45]",
  accentBg: "bg-[#B96E45]",
  accentDarkBg: "bg-[#7C3F25]",
  olive: "text-[#6F7D55]",
  teal: "text-[#2F7C78]",
  amber: "text-[#D99A3D]",
  card: "rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_10px_28px_rgba(80,57,34,0.06)]",
  cardCompact: "rounded-[20px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_8px_22px_rgba(80,57,34,0.05)]",
  primaryButton: "bg-[#251F1A] text-[#FFFDF8] transition hover:bg-[#16110E]",
  outlineButton: "border border-[#E3D2BC] bg-[#FFFDF8] text-[#251F1A] transition hover:bg-[#F1E7D8]"
};

export function PublicPageIntro({
  eyebrow,
  title,
  description,
  children,
  compact = false
}: {
  eyebrow?: string;
  title: string;
  description: string;
  children?: ReactNode;
  compact?: boolean;
}) {
  return (
    <div className={cn(publicTheme.card, compact ? "p-4 sm:p-5" : "p-5 sm:p-6")}>
      {eyebrow ? <p className="text-xs font-semibold uppercase tracking-[0.18em] text-[#B96E45]">{eyebrow}</p> : null}
      <h1 className="mt-2 max-w-3xl text-2xl font-bold tracking-tight text-[#251F1A] sm:text-4xl">{title}</h1>
      <p className="mt-2 max-w-3xl text-sm leading-7 text-[#6F6256] sm:text-base">{description}</p>
      {children ? <div className="mt-4">{children}</div> : null}
    </div>
  );
}

export function PublicSectionHeader({
  title,
  description,
  href,
  linkLabel = "Semua"
}: {
  title: string;
  description?: string;
  href?: string;
  linkLabel?: string;
}) {
  return (
    <div className="flex items-end justify-between gap-3">
      <div className="min-w-0">
        <h2 className="text-xl font-bold tracking-tight text-[#251F1A] sm:text-2xl">{title}</h2>
        {description ? <p className="mt-1 max-w-2xl text-sm leading-6 text-[#6F6256]">{description}</p> : null}
      </div>
      {href ? (
        <Link className="shrink-0 text-sm font-semibold text-[#B96E45] hover:text-[#7C3F25]" href={href}>
          Lihat {linkLabel}
        </Link>
      ) : null}
    </div>
  );
}

export function PublicLinkButton({
  href,
  children,
  variant = "primary",
  className
}: {
  href: string;
  children: ReactNode;
  variant?: "primary" | "outline";
  className?: string;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "inline-flex min-h-10 items-center justify-center rounded-xl px-4 text-sm font-semibold",
        variant === "primary" ? publicTheme.primaryButton : publicTheme.outlineButton,
        className
      )}
    >
      {children}
    </Link>
  );
}

export function PaymentNotice({ children, tone = "amber" }: { children: ReactNode; tone?: "amber" | "olive" | "teal" }) {
  const toneClass =
    tone === "olive"
      ? "border-[#D9DEC8] bg-[#F4F7ED] text-[#4F5F3C]"
      : tone === "teal"
        ? "border-[#CFE1DC] bg-[#EFF7F4] text-[#2F635F]"
        : "border-[#E8D2AA] bg-[#FFF5DE] text-[#7A4D1D]";

  return <div className={cn("rounded-2xl border p-4 text-sm leading-6", toneClass)}>{children}</div>;
}

export function PriceText({ children, className }: { children: ReactNode; className?: string }) {
  return <p className={cn("font-bold text-[#B96E45]", className)}>{children}</p>;
}
