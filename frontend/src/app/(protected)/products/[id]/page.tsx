import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { RouteIndicator } from "@/components/ui/route-indicator";

type ProductPageProps = {
  params: Promise<{
    id: string;
  }>;
};

export default async function ProductPlaceholderPage({ params }: ProductPageProps) {
  const { id } = await params;

  return (
    <main className="mx-auto w-full max-w-[960px] px-6 py-6">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle>Product detail</CardTitle>
              <CardDescription className="font-mono">{id}</CardDescription>
            </div>
            <RouteIndicator size="sm" />
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            GAU-248 builds the full route indicator, spec table, datasheet preview, and sticky quote panel.
            Search results already link here so the buyer flow has a stable destination.
          </p>
          <Button asChild variant="secondary">
            <Link href="/search">Back to search</Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}
