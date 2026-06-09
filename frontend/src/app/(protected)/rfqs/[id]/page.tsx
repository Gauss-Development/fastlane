import { RFQDetailClient } from "@/components/rfqs/rfq-detail-client";

type RFQDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function RFQDetailPage({ params }: RFQDetailPageProps) {
  const { id } = await params;
  return <RFQDetailClient rfqId={decodeURIComponent(id)} />;
}
