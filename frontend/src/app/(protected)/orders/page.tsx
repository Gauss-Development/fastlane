import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { RouteIndicator } from "@/components/ui/route-indicator";
import { StatusPill, type StatusTone } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";

const demoOrders: Array<{
  id: string;
  part: string;
  status: string;
  tone: StatusTone;
  total: string;
  location: string;
}> = [
  { id: "ORD-20260429-0089-SFO", part: "100G QSFP28 LR4", status: "shipped", tone: "info", total: "$18,200", location: "Yantian Port" },
  { id: "ORD-20260428-0074-SFO", part: "10G SFP+ LR", status: "qc passed", tone: "success", total: "$19,000", location: "Shenzhen factory" },
  { id: "ORD-20260427-0068-LAX", part: "400G QSFP-DD DR4", status: "in production", tone: "warning", total: "$19,920", location: "Suzhou factory" },
];

export default function OrdersPage() {
  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle>Orders</CardTitle>
              <CardDescription>
                Timeline-backed order tracking arrives with GAU-268 and GAU-253.
              </CardDescription>
            </div>
            <RouteIndicator size="sm" />
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHead>
              <Tr>
                <Th>Order</Th>
                <Th>Part</Th>
                <Th>Status</Th>
                <Th numeric>Total</Th>
                <Th>Location</Th>
              </Tr>
            </TableHead>
            <TableBody>
              {demoOrders.map((order) => (
                <Tr key={order.id}>
                  <Td><CodeId code={order.id} size="sm" /></Td>
                  <Td className="font-mono text-xs">{order.part}</Td>
                  <Td><StatusPill tone={order.tone}>{order.status}</StatusPill></Td>
                  <Td numeric>{order.total}</Td>
                  <Td className="font-mono text-xs text-muted-foreground">{order.location}</Td>
                </Tr>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </main>
  );
}
