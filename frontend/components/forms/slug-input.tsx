"use client";

import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";

type SlugInputProps = {
  sourceValue: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  hasError?: boolean;
};

export function SlugInput({
  sourceValue,
  value,
  onChange,
  placeholder = "contoh-slug",
  hasError = false
}: SlugInputProps) {
  const [autoGenerate, setAutoGenerate] = useState(value.length === 0);

  useEffect(() => {
    if (autoGenerate) {
      onChange(slugify(sourceValue));
    }
  }, [autoGenerate, onChange, sourceValue]);

  return (
    <Input
      value={value}
      placeholder={placeholder}
      hasError={hasError}
      onChange={(event) => {
        setAutoGenerate(false);
        onChange(event.target.value.toLowerCase());
      }}
    />
  );
}

export function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
