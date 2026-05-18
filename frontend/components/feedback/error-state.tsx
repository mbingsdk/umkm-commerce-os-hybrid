"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

type ErrorStateProps = {
  title?: string;
  description?: string;
  onRetry?: () => void;
};

export function ErrorState({
  title = "Data gagal dimuat",
  description = "Koneksi atau server sedang bermasalah. Coba muat ulang.",
  onRetry
}: ErrorStateProps) {
  return (
    <Card className="border-red-200 bg-red-50 shadow-none">
      <CardContent>
        <h2 className="text-base font-semibold text-red-900">{title}</h2>
        <p className="mt-1 text-sm leading-6 text-red-700">{description}</p>
        {onRetry ? (
          <Button className="mt-4" variant="outline" onClick={onRetry}>
            Muat ulang
          </Button>
        ) : null}
      </CardContent>
    </Card>
  );
}
