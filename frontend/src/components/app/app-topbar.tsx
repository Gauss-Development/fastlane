"use client";

import { RouteIndicator } from "@/components/ui/route-indicator";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import { useAuthStore } from "@/lib/stores/auth-store";

function roleLabel(role: string | undefined) {
  if (role === "manufacturer") return "SUPPLIER";
  if (role === "admin") return "ADMIN";
  return "BUYER";
}

// One top bar across every authenticated screen: the cross-border brand mark on
// the left, role/identity + theme toggle on the right.
export function AppTopbar() {
  const user = useAuthStore((s) => s.user);

  return (
    <header className="sticky top-0 z-10 flex h-12 shrink-0 items-center justify-between gap-3 border-b border-border bg-background/95 px-4 backdrop-blur">
      <RouteIndicator size="sm" />
      <div className="flex items-center gap-3">
        <span className="hidden font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground sm:inline">
          {roleLabel(user?.role)}
          {user?.email ? ` · ${user.email}` : ""}
        </span>
        <ThemeToggle />
      </div>
    </header>
  );
}
