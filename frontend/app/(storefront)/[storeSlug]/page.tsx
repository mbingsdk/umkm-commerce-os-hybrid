import { redirect } from "next/navigation";

type StorefrontPageProps = {
  params: Promise<{ storeSlug: string }>;
};

export default async function LegacyStorefrontRedirect({ params }: StorefrontPageProps) {
  const { storeSlug } = await params;
  redirect(`/s/${encodeURIComponent(storeSlug)}`);
}
