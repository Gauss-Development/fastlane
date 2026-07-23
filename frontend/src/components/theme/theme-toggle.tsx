"use client";

import { Moon, Sun } from "lucide-react";

import { Button } from "@/components/ui/button";
import { getResolvedTheme } from "@/lib/theme/apply-theme";
import { useThemeStore } from "@/lib/theme/theme-store";

export function ThemeToggle() {
  const preference = useThemeStore((s) => s.preference);
  const setPreference = useThemeStore((s) => s.setPreference);
  const resolved = getResolvedTheme(preference);

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-8 w-8"
      aria-label={resolved === "dark" ? "Switch to light theme" : "Switch to dark theme"}
      onClick={() => setPreference(resolved === "dark" ? "light" : "dark")}
    >
      {resolved === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
    </Button>
  );
}
