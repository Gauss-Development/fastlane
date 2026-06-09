import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill } from "@/components/ui/pill";
import { featuredSuppliers } from "@/components/search/demo-data";

export default function SuppliersPage() {
  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1>Suppliers</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Verified Chinese photonics suppliers seeded for the MVP catalog.
          </p>
        </div>
        <RouteIndicator size="md" />
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {featuredSuppliers.map((supplier) => (
          <Card key={supplier.code}>
            <CardHeader>
              <div className="flex items-start justify-between gap-3">
                <CodeId code={supplier.code} size="sm" />
                <StatusPill tone="success">Verified</StatusPill>
              </div>
              <div>
                <CardTitle>{supplier.name}</CardTitle>
                <CardDescription className="font-mono">
                  {supplier.nameZh} • {supplier.city}
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="font-mono text-xs uppercase tracking-[0.08em] text-muted-foreground">
                {supplier.capability}
              </p>
              <dl className="grid grid-cols-2 gap-3 font-mono text-sm">
                <div>
                  <dt className="text-xs uppercase text-muted-foreground">On-time</dt>
                  <dd className="text-success">{supplier.onTimeRate}</dd>
                </div>
                <div>
                  <dt className="text-xs uppercase text-muted-foreground">Orders</dt>
                  <dd>{supplier.orders}</dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        ))}
      </div>
    </main>
  );
}
