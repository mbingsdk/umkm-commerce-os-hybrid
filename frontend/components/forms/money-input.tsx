"use client";

import { useMemo } from "react";
import { Input } from "@/components/ui/input";

type MoneyInputProps = {
  value: number | null;
  onChange: (value: number | null) => void;
  placeholder?: string;
  disabled?: boolean;
  hasError?: boolean;
};

export function MoneyInput({ value, onChange, placeholder = "0", disabled = false, hasError = false }: MoneyInputProps) {
  const displayValue = useMemo(() => {
    if (value === null) {
      return "";
    }

    return new Intl.NumberFormat("id-ID", {
      maximumFractionDigits: 0
    }).format(value);
  }, [value]);

  return (
    <div className="relative">
      <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm text-neutral-500">Rp</span>
      <Input
        inputMode="numeric"
        className="pl-10"
        value={displayValue}
        placeholder={placeholder}
        disabled={disabled}
        hasError={hasError}
        onChange={(event) => {
          const digits = event.target.value.replace(/\D/g, "");
          onChange(digits === "" ? null : Number(digits));
        }}
      />
    </div>
  );
}
