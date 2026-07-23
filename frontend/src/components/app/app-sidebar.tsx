"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import {
  Building2,
  Factory,
  FileText,
  FolderKanban,
  Inbox,
  LayoutDashboard,
  LogOut,
  Package,
  PanelLeft,
  PanelLeftClose,
  Search,
  Settings,
  type LucideIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/lib/stores/auth-store";
import { logoutSession } from "@/lib/auth/client-api";
import { cn } from "@/lib/utils";

const SIDEBAR_W_OPEN = 236;
const SIDEBAR_W_CLOSED = 60;
const COLLAPSE_KEY = "fiberlane_sidebar_collapsed";

type NavItem = { href: string; label: string; icon: LucideIcon };

function navItemsForRole(role: string | undefined): NavItem[] {
  if (role === "manufacturer") {
    return [
      { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
      { href: "/rfqs", label: "Open RFQs", icon: Inbox },
      { href: "/orders", label: "Orders", icon: Package },
      { href: "/manufacturer-profile", label: "My Profile", icon: Building2 },
      { href: "/app/profile", label: "Settings", icon: Settings },
    ];
  }
  // startup (buyer) + admin fallback
  return [
    { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/search", label: "Search", icon: Search },
    { href: "/rfqs", label: "RFQs", icon: FileText },
    { href: "/orders", label: "Orders", icon: Package },
    { href: "/projects", label: "Projects", icon: FolderKanban },
    { href: "/manufacturers", label: "Manufacturers", icon: Factory },
    { href: "/app/profile", label: "Settings", icon: Settings },
  ];
}

export function AppSidebar() {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const [collapsed, setCollapsed] = useState(false);
  const navItems = navItemsForRole(user?.role);

  // Restore the collapse preference client-side (avoids SSR hydration mismatch).
  useEffect(() => {
    if (localStorage.getItem(COLLAPSE_KEY) === "1") setCollapsed(true);
  }, []);

  function toggleCollapsed() {
    setCollapsed((c) => {
      const next = !c;
      localStorage.setItem(COLLAPSE_KEY, next ? "1" : "0");
      return next;
    });
  }

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
            className="font-mono text-sm font-bold tracking-[0.2em] text-foreground transition-colors hover:text-primary"
          >
            FIBERLANE
          </Link>
        )}
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleCollapsed}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="h-8 w-8"
        >
          {collapsed ? <PanelLeft /> : <PanelLeftClose />}
        </Button>
      </div>

      <nav className="flex flex-1 flex-col gap-0.5 py-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          const active =
            pathname === item.href || pathname.startsWith(`${item.href}/`);
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              title={collapsed ? item.label : undefined}
              className={cn(
                "flex items-center gap-3 px-3 py-2",
                "border-l-2 border-l-transparent",
                "font-mono text-xs uppercase tracking-[0.1em]",
                "text-muted-foreground transition-colors hover:bg-accent hover:text-foreground",
                collapsed && "justify-center",
                active && "border-l-primary bg-accent text-foreground",
              )}
            >
              <Icon className={cn("size-4 shrink-0", active && "text-primary")} />
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
          title={collapsed ? "Sign out" : undefined}
          className={cn(
            "flex w-full items-center gap-3 px-3 py-2",
            "border-l-2 border-l-transparent",
            "font-mono text-xs uppercase tracking-[0.1em]",
            "text-muted-foreground transition-colors hover:bg-accent hover:text-destructive",
            collapsed && "justify-center",
          )}
        >
          <LogOut className="size-4 shrink-0" />
          {!collapsed && <span>Sign out</span>}
        </button>
      </div>
    </aside>
  );
}
