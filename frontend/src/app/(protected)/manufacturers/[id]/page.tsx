import { ManufacturerDetailClient } from "@/components/manufacturers/manufacturer-detail-client";

type ManufacturerDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function ManufacturerDetailPage({ params }: ManufacturerDetailPageProps) {
  const { id } = await params;
  return <ManufacturerDetailClient manufacturerId={decodeURIComponent(id)} />;
}
