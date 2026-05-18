import { Skeleton } from "@/components/ui/skeleton";

type LoadingStateProps = {
  lines?: number;
};

export function LoadingState({ lines = 3 }: LoadingStateProps) {
  return (
    <div className="space-y-4">
      {Array.from({ length: lines }).map((_, index) => (
        <Skeleton key={index} className={index === 0 ? "h-24 w-full" : "h-16 w-full"} />
      ))}
    </div>
  );
}
