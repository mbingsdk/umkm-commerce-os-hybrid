import { Badge } from "@/components/ui/badge";
import { CreateStoreWizard } from "@/features/onboarding/components/create-store-wizard";

export default function CreateStoreOnboardingPage() {
  return (
    <section className="space-y-6">
      <div>
        <Badge tone="primary">Onboarding</Badge>
        <h1 className="mt-3 text-2xl font-bold text-neutral-950">Buat toko pertama</h1>
        <p className="mt-3 text-sm leading-6 text-neutral-500">
          Siapkan identitas toko dulu. Setelah toko dibuat, kamu bisa mulai mengelola produk dan operasional.
        </p>
      </div>

      <CreateStoreWizard />
    </section>
  );
}
