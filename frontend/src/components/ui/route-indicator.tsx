import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

/**
 * RouteIndicator — the cross-border brand mark.
 *
 *   SHENZHEN ─────► SAN FRANCISCO
 *
 * Appears on most screens. Origin carries the CN marker, destination the US
 * marker. Coordinates show by default at `lg` (landing hero) and can be forced
 * on/off elsewhere. Endpoints are configurable for product-detail pages where
 * the destination is the buyer's delivery city.
 */

export type RoutePlace = {
  city: string;
  /** e.g. "22.54°N 114.06°E" — shown beneath the city when coords are on. */
  coords?: string;
};

const SHENZHEN: RoutePlace = { city: "SHENZHEN", coords: "22.54°N 114.06°E" };
const SAN_FRANCISCO: RoutePlace = {
  city: "SAN FRANCISCO",
  coords: "37.77°N 122.42°W",
};

const routeVariants = cva(
  "inline-flex items-center font-mono uppercase text-foreground",
  {
    variants: {
      size: {
        sm: "gap-2 text-[11px] tracking-[0.1em]",
        md: "gap-2.5 text-xs tracking-[0.12em]",
        lg: "gap-3 text-base tracking-[0.16em]",
      },
    },
    defaultVariants: { size: "md" },
  },
);

const connectorWidth: Record<NonNullable<RouteSize>, string> = {
  sm: "w-6",
  md: "w-10",
  lg: "w-20",
};

type RouteSize = VariantProps<typeof routeVariants>["size"];

interface RouteIndicatorProps
  extends Omit<React.ComponentProps<"div">, "color">,
    VariantProps<typeof routeVariants> {
  from?: RoutePlace;
  to?: RoutePlace;
  /** Force coordinates on/off. Defaults to on for `lg`, off otherwise. */
  showCoords?: boolean;
}

function Endpoint({
  place,
  marker,
  align,
  showCoords,
}: {
  place: RoutePlace;
  marker: "cn" | "us";
  align: "start" | "end";
  showCoords: boolean;
}) {
  return (
    <span
      className={cn(
        "inline-flex flex-col",
        align === "end" ? "items-end" : "items-start",
      )}
    >
      <span className="inline-flex items-center gap-1.5">
        <span
          aria-hidden
          className={cn(
            "size-1.5 shrink-0",
            marker === "cn" ? "bg-marker-cn" : "bg-marker-us",
          )}
        />
        <span className="font-medium leading-none">{place.city}</span>
      </span>
      {showCoords && place.coords ? (
        <span className="mt-1 text-[0.78em] font-normal normal-case tracking-normal text-muted-foreground">
          {place.coords}
        </span>
      ) : null}
    </span>
  );
}

function RouteIndicator({
  className,
  size = "md",
  from = SHENZHEN,
  to = SAN_FRANCISCO,
  showCoords,
  ...props
}: RouteIndicatorProps) {
  const coords = showCoords ?? size === "lg";

  return (
    <div
      data-slot="route-indicator"
      className={cn(routeVariants({ size }), className)}
      {...props}
    >
      <Endpoint place={from} marker="cn" align="start" showCoords={coords} />
      <span
        aria-hidden
        className={cn(
          "relative inline-flex items-center self-center",
          connectorWidth[size ?? "md"],
        )}
      >
        <span className="h-px w-full bg-border-strong" />
        <span className="-ml-1 leading-none text-border-strong">►</span>
      </span>
      <Endpoint place={to} marker="us" align="end" showCoords={coords} />
    </div>
  );
}

export { RouteIndicator, SHENZHEN, SAN_FRANCISCO };
