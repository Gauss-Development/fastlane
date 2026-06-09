"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

/**
 * CodeId — renders system identifiers like `RFQ-20260429-0142-SZX`.
 *
 * Always uppercase, monospace, tabular. The leading prefix is tinted by domain
 * (RFQ / ORD / SUP / QUOTE); unknown prefixes render in plain ink. When
 * `copyable`, clicking copies the full code to the clipboard.
 */

const prefixTone: Record<string, string> = {
  RFQ: "text-primary",
  ORD: "text-marker-us",
  SUP: "text-success",
  QUOTE: "text-warning",
};

const codeIdVariants = cva(
  "inline-flex items-center gap-1.5 font-mono uppercase tracking-[0.04em] whitespace-nowrap tabular-nums",
  {
    variants: {
      size: {
        sm: "text-xs",
        md: "text-sm",
        lg: "text-base",
      },
    },
    defaultVariants: { size: "md" },
  },
);

interface CodeIdProps
  extends Omit<React.HTMLAttributes<HTMLElement>, "color">,
    VariantProps<typeof codeIdVariants> {
  code: string;
  copyable?: boolean;
}

function CodeId({ code, copyable = false, size, className, ...props }: CodeIdProps) {
  const upper = code.toUpperCase();
  const prefix = upper.split("-")[0];
  const tone = prefixTone[prefix];
  const [copied, setCopied] = React.useState(false);

  const body = tone ? (
    <span>
      <span className={cn("font-medium", tone)}>{prefix}</span>
      <span className="text-foreground">{upper.slice(prefix.length)}</span>
    </span>
  ) : (
    <span className="text-foreground">{upper}</span>
  );

  if (!copyable) {
    return (
      <span
        data-slot="code-id"
        className={cn(codeIdVariants({ size }), className)}
        {...props}
      >
        {body}
      </span>
    );
  }

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(upper);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1200);
    } catch {
      /* clipboard unavailable — no-op */
    }
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label={`Copy ${upper}`}
      data-slot="code-id"
      className={cn(
        codeIdVariants({ size }),
        "group rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
        className,
      )}
      {...props}
    >
      {body}
      <span
        aria-hidden
        className="text-muted-foreground transition-colors group-hover:text-foreground"
      >
        {copied ? (
          <Check className="size-3.5 text-success" />
        ) : (
          <Copy className="size-3.5" />
        )}
      </span>
    </button>
  );
}

export { CodeId };
