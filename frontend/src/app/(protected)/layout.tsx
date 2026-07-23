import { SessionBootstrap } from "@/components/app/session-bootstrap";
import { AppShell } from "@/components/app/app-shell";

// One shell for every authenticated screen: role-aware collapsible sidebar +
// top bar. Individual pages render content only.
export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SessionBootstrap>
      <AppShell>{children}</AppShell>
    </SessionBootstrap>
  );
}
