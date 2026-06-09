import { SearchResultsClient } from "@/components/search/search-results-client";

type SearchPageProps = {
  searchParams: Promise<{
    q?: string | string[];
  }>;
};

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const raw = Array.isArray(params.q) ? params.q[0] : params.q;
  return <SearchResultsClient initialQuery={raw ?? ""} />;
}
