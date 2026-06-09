import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CodeId } from "@/components/ui/code-id";
import { StatusPill } from "@/components/ui/pill";
import { Table, TableBody, Td, Th, TableHead, Tr } from "@/components/ui/table";
import { recentRFQs } from "@/components/search/demo-data";

export default function RFQsPage() {
  return (
    <main className="mx-auto w-full max-w-[1200px] px-6 py-6">
      <Card>
        <CardHeader>
          <CardTitle>RFQs</CardTitle>
          <CardDescription>
            Demo RFQ workbench. Persistent RFQ creation and supplier email dispatch land in GAU-249.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHead>
              <Tr>
                <Th>RFQ</Th>
                <Th>Request</Th>
                <Th numeric>Qty</Th>
                <Th>Status</Th>
                <Th numeric>Age</Th>
              </Tr>
            </TableHead>
            <TableBody>
              {recentRFQs.map((rfq) => (
                <Tr key={rfq.id}>
                  <Td><CodeId code={rfq.id} size="sm" /></Td>
                  <Td className="font-mono text-xs">{rfq.query}</Td>
                  <Td numeric>{rfq.qty}</Td>
                  <Td><StatusPill tone={rfq.tone}>{rfq.status}</StatusPill></Td>
                  <Td numeric>{rfq.age}</Td>
                </Tr>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </main>
  );
}
