import { Skeleton } from "@/components/ui/skeleton";

export default function StorefrontLoading() {
  return (
    <main className="mx-auto max-w-6xl space-y-6 px-4 py-8 sm:px-6 lg:px-8">
      <Skeleton className="h-64 w-full rounded-3xl" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="space-y-4">
          <Skeleton className="h-28 w-full rounded-3xl" />
          <Skeleton className="h-10 w-64" />
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {Array.from({ length: 6 }).map((_, index) => (
              <Skeleton key={index} className="h-80 w-full rounded-3xl" />
            ))}
          </div>
        </div>
        <Skeleton className="h-64 w-full rounded-3xl" />
      </div>
    </main>
  );
}
