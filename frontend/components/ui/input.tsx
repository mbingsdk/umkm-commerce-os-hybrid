import type { InputHTMLAttributes } from "react";
import { cn } from "@/lib/utils/cn";

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  hasError?: boolean;
};

export function Input({ className, hasError = false, ...props }: InputProps) {
  return (
    <input
      className={cn(
        "h-10 w-full rounded-xl border bg-white px-3 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:ring-2",
        hasError
          ? "border-red-500 focus:border-red-500 focus:ring-red-200"
          : "border-neutral-300 focus:border-primary-500 focus:ring-primary-100",
        className
      )}
      {...props}
    />
  );
}
