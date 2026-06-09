import * as React from "react";
import { ChevronDown, ChevronsUpDown, ChevronUp } from "lucide-react";

import { cn } from "@/lib/utils";

/**
 * Table primitives — dense, Bloomberg-terminal feel.
 *
 * 40px rows, sticky header, zebra OFF. `<Td numeric>` right-aligns and switches
 * to tabular monospace. `<Th sortable>` shows a chevron (hover for unsorted,
 * directional when active); sort state/handlers belong to the consuming screen.
 */

function Table({ className, ...props }: React.ComponentProps<"table">) {
  return (
    <div className="w-full overflow-x-auto rounded-md border border-border">
      <table
        data-slot="table"
        className={cn("w-full border-collapse text-sm", className)}
        {...props}
      />
    </div>
  );
}

function TableHead({ className, ...props }: React.ComponentProps<"thead">) {
  return (
    <thead
      data-slot="table-head"
      className={cn("[&_tr]:border-b [&_tr]:border-border", className)}
      {...props}
    />
  );
}

function TableBody({ className, ...props }: React.ComponentProps<"tbody">) {
  return (
    <tbody
      data-slot="table-body"
      className={cn("[&_tr:last-child]:border-0", className)}
      {...props}
    />
  );
}

function Tr({ className, ...props }: React.ComponentProps<"tr">) {
  return (
    <tr
      data-slot="table-row"
      className={cn(
        "border-b border-border transition-colors hover:bg-muted/60",
        className,
      )}
      {...props}
    />
  );
}

type SortDirection = "asc" | "desc" | null;

interface ThProps extends React.ComponentProps<"th"> {
  numeric?: boolean;
  sortable?: boolean;
  sortDirection?: SortDirection;
}

function Th({
  className,
  numeric,
  sortable,
  sortDirection = null,
  children,
  ...props
}: ThProps) {
  return (
    <th
      data-slot="table-th"
      aria-sort={
        sortDirection === "asc"
          ? "ascending"
          : sortDirection === "desc"
            ? "descending"
            : undefined
      }
      className={cn(
        "sticky top-0 z-10 h-10 bg-card px-3 align-middle",
        "text-xs font-medium uppercase tracking-[0.08em] text-muted-foreground",
        numeric ? "text-right" : "text-left",
        sortable && "cursor-pointer select-none hover:text-foreground",
        className,
      )}
      {...props}
    >
      <span
        className={cn(
          "group inline-flex items-center gap-1",
          numeric && "flex-row-reverse",
        )}
      >
        {children}
        {sortable ? (
          <span aria-hidden className="text-border-strong">
            {sortDirection === "asc" ? (
              <ChevronUp className="size-3.5" />
            ) : sortDirection === "desc" ? (
              <ChevronDown className="size-3.5" />
            ) : (
              <ChevronsUpDown className="size-3.5 opacity-0 transition-opacity group-hover:opacity-100" />
            )}
          </span>
        ) : null}
      </span>
    </th>
  );
}

interface TdProps extends React.ComponentProps<"td"> {
  numeric?: boolean;
}

function Td({ className, numeric, ...props }: TdProps) {
  return (
    <td
      data-slot="table-td"
      className={cn(
        "h-10 px-3 align-middle",
        numeric ? "text-right font-mono tabular-nums" : "text-left",
        className,
      )}
      {...props}
    />
  );
}

export { Table, TableHead, TableBody, Tr, Th, Td };
export type { SortDirection };
