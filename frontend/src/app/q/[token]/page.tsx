import { SupplierRFQClient } from "@/components/supplier/supplier-rfq-client";

type SupplierRFQPageProps = {
  params: Promise<{ token: string }>;
};

// Public supplier RFQ response page (Screen 11). No auth, no nav — the
// signed magic-link token in the URL is the only credential.
export default async function SupplierRFQPage({ params }: SupplierRFQPageProps) {
  const { token } = await params;
  return <SupplierRFQClient token={token} />;
}
