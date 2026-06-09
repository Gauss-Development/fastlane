"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { LogOut, PanelLeft, PanelLeftClose } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/lib/stores/auth-store";
import { logoutSession } from "@/lib/auth/client-api";
import { cn } from "@/lib/utils";

const SIDEBAR_W_OPEN = 240;
const SIDEBAR_W_CLOSED = 56;

type NavItem = {
  href: string;
  label: string;
  code: string;
};

const NAV_ITEMS: NavItem[] = [
  { href: "/dashboard", label: "Search", code: "01" },
  { href: "/rfqs", label: "RFQs", code: "02" },
  { href: "/orders", label: "Orders", code: "03" },
  { href: "/suppliers", label: "Suppliers", code: "04" },
  { href: "/app/profile", label: "Settings", code: "05" },
];

export function AppSidebar() {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const [collapsed, setCollapsed] = useState(false);

  async function handleLogout() {
    await logoutSession();
    window.location.replace("/auth/login");
  }

  return (
    <aside
      style={{ width: collapsed ? SIDEBAR_W_CLOSED : SIDEBAR_W_OPEN }}
      className="flex shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-200 ease-out"
    >
      <div
        className={cn(
          "flex h-12 shrink-0 items-center border-b border-sidebar-border px-3",
          collapsed ? "justify-center" : "justify-between",
        )}
      >
        {!collapsed && (
          <Link
            href="/dashboard"
            className="font-mono text-sm font-bold tracking-[0.2em] text-foreground hover:text-primary transition-colors"
          >
            FIBERLANE
          </Link>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setCollapsed((c) => !c)}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="h-8 w-8"
        >
          {collapsed ? <PanelLeft /> : <PanelLeftClose />}
        </Button>
      </div>

      <nav className="flex flex-1 flex-col py-2">
        {NAV_ITEMS.map((item) => {
          const active =
            pathname === item.href ||
            pathname === item.href + "/" ||
            (item.href === "/dashboard" && pathname.startsWith("/search")) ||
            (item.href !== "/dashboard" && pathname.startsWith(`${item.href}/`));
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "flex items-center gap-3 px-3 py-2",
                "border-l-2 border-l-transparent",
                "font-mono text-xs uppercase tracking-[0.1em]",
                "text-muted-foreground hover:text-foreground hover:bg-accent",
                "transition-colors",
                active && "border-l-primary text-foreground bg-accent",
              )}
            >
              <span className="w-5 shrink-0 tabular-nums text-primary">{item.code}</span>
              {!collapsed && <span className="truncate">{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-sidebar-border">
        {!collapsed && user && (
          <div className="truncate px-3 py-2 font-mono text-xs text-muted-foreground">
            {user.email}
          </div>
        )}
        <button
          type="button"
          onClick={handleLogout}
          aria-label="Sign out"
          className={cn(
            "flex w-full items-center gap-3 px-3 py-2",
            "border-l-2 border-l-transparent",
            "font-mono text-xs uppercase tracking-[0.1em]",
            "text-muted-foreground hover:text-destructive hover:bg-accent",
            "transition-colors",
          )}
        >
          <LogOut className="size-4 shrink-0" />
          {!collapsed && <span>Sign out</span>}
        </button>
      </div>
    </aside>
  );
}
