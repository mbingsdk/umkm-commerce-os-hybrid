import { Skeleton } from "@/components/ui/skeleton";

export default function StorefrontLoading() {
  return (
    <main className="mx-auto max-w-[1500px] space-y-4 px-4 py-4 sm:px-6 sm:py-6 lg:px-8">
      <Skeleton className="h-48 w-full rounded-[28px]" />
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_300px]">
        <div className="space-y-3">
          <Skeleton className="h-20 w-full rounded-[24px]" />
          <Skeleton className="h-10 w-64" />
          <div className="grid grid-cols-2 gap-2 sm:gap-2.5 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
            {Array.from({ length: 10 }).map((_, index) => (
              <Skeleton key={index} className="h-56 w-full rounded-[20px]" />
            ))}
          </div>
        </div>
        <Skeleton className="h-56 w-full rounded-[24px]" />
      </div>
    </main>
  );
}
