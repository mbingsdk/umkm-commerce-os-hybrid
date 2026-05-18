import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col justify-center gap-4 px-4 sm:px-6 lg:px-8">
      <Skeleton className="h-6 w-32" />
      <Skeleton className="h-12 w-2/3" />
      <Skeleton className="h-32 w-full" />
    </main>
  );
}
