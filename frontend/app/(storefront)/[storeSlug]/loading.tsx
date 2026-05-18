import { Skeleton } from "@/components/ui/skeleton";

export default function StorefrontLoading() {
  return (
    <main className="mx-auto max-w-5xl space-y-4 px-4 py-12 sm:px-6 lg:px-8">
      <Skeleton className="h-6 w-36" />
      <Skeleton className="h-10 w-56" />
      <Skeleton className="h-28 w-full" />
    </main>
  );
}
