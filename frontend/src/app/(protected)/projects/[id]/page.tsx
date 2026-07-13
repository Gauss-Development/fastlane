import { ProjectDetailClient } from "@/components/projects/project-detail-client";

type ProjectDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function ProjectDetailPage({ params }: ProjectDetailPageProps) {
  const { id } = await params;
  return <ProjectDetailClient projectId={decodeURIComponent(id)} />;
}
