<script setup lang="ts">
import { computed, ref } from "vue";
import { MousePointer2, PencilRuler, Printer } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { MapData, MissionData } from "@/types/api";

const props = withDefaults(
  defineProps<{
    missionData: MissionData;
    mapData: MapData;
    title?: string;
  }>(),
  { title: "Astrohop" },
);

// ---------------------------------------------------------------------------
// Page geometry — A4 landscape, laid out in millimetres. The SVG viewBox uses
// mm as its user unit, so the same markup prints at true A4 size and scales
// cleanly on screen via CSS (width: 100%; aspect-ratio: 297 / 210).
// ---------------------------------------------------------------------------
const PAGE_W = 297;
const PAGE_H = 210;
const MARGIN = 10;
const HEADER_H = 15;
const SIDEBAR_W = 54;
const GAP = 7;

const frame = {
  x: MARGIN,
  y: MARGIN,
  w: PAGE_W - MARGIN * 2,
  h: PAGE_H - MARGIN * 2,
};

const circleArea = {
  x: MARGIN + 4,
  y: MARGIN + HEADER_H,
  w: frame.w - SIDEBAR_W - GAP - 8,
  h: frame.h - HEADER_H - 8,
};

const center = {
  x: circleArea.x + circleArea.w / 2,
  y: circleArea.y + circleArea.h / 2,
};
const radius = Math.min(circleArea.w, circleArea.h) / 2 - 9;

const sidebarX = PAGE_W - MARGIN - SIDEBAR_W - 2;

// ---------------------------------------------------------------------------
// Polar projection: centre = zenith (alt 90°), outer ring = horizon (alt 0°).
// North is up, azimuth increases clockwise (N-E-S-W), matching a chart you'd
// hold facing north rather than a naked-eye "lying on your back" view.
// ---------------------------------------------------------------------------
const ALT_RINGS = [15, 30, 45, 60, 75];
const AZ_TICKS = [0, 30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330];
const CARDINALS: Record<number, string> = {
  0: "N",
  90: "E",
  180: "S",
  270: "W",
};

function ringRadius(alt: number): number {
  return radius * (1 - Math.max(0, Math.min(90, alt)) / 90);
}

function toXY(alt: number, az: number) {
  const r = ringRadius(alt);
  const theta = (az * Math.PI) / 180;
  return {
    x: center.x + r * Math.sin(theta),
    y: center.y - r * Math.cos(theta),
  };
}

function edgeXY(az: number, extra: number) {
  const r = radius + extra;
  const theta = (az * Math.PI) / 180;
  return {
    x: center.x + r * Math.sin(theta),
    y: center.y - r * Math.cos(theta),
  };
}

// ---------------------------------------------------------------------------
// Data joins.
//
// ASSUMPTION: `mapData.positions[i]` lines up by array index with
// `missionData.objectives[i]` — the API types don't carry a name on
// `Horizontal`, so this is the only way to label each point. If the backend
// ever returns positions in a different order than objectives (e.g. with
// extra anchor stars spliced in), this will mislabel points and needs a
// dedicated id on Horizontal to fix properly.
//
// `mapData.tour.order` is read as a visiting sequence of indices into that
// same array, and `mapData.tour.distances[i][j]` as the angular distance
// between objectives i and j (used to label each hop).
// ---------------------------------------------------------------------------
const points = computed(() =>
  props.mapData.positions.map((pos, i) => ({
    index: i,
    alt: pos.alt,
    az: pos.az,
    name: props.missionData.objectives[i]?.name ?? `Object ${i + 1}`,
    ...toXY(pos.alt, pos.az),
  })),
);

const routeOrder = computed(() =>
  props.mapData.tour.order.filter((i) => points.value[i] != null),
);

const segments = computed(() => {
  const order = routeOrder.value;
  const list: {
    from: (typeof points.value)[number];
    to: (typeof points.value)[number];
    distance: number | null;
  }[] = [];
  for (let k = 0; k < order.length - 1; k++) {
    const from = points.value[order[k]];
    const to = points.value[order[k + 1]];
    const distance =
      props.mapData.tour.distances?.[order[k]]?.[order[k + 1]] ?? null;
    list.push({ from, to, distance });
  }
  return list;
});

const ANGULAR_TIPS = [
  "Finger width ~ 1°",
  "Three middle fingers ~ 5°",
  "Fist width ~ 10°",
  "Index to pinky ~ 15°",
  "Open hand width ~ 20°",
];

// ---------------------------------------------------------------------------
// Sidebar — Angular Distance Tips + Legend, single column. Text sized up a
// notch from the original print-style proportions since it reads too small
// on screen; the geometry below (line-height increments) grew to match.
// ---------------------------------------------------------------------------
interface SidebarLine {
  x: number;
  y: number;
  text: string;
  cls: string;
}

const sidebar = computed(() => {
  const items: SidebarLine[] = [];
  let y = circleArea.y + 4;

  const heading = (text: string) => {
    items.push({ x: sidebarX, y, text, cls: "sb-heading" });
    y += 5.2;
  };
  const line = (text: string, cls = "sb-body") => {
    items.push({ x: sidebarX, y, text, cls });
    y += 4.4;
  };

  heading("Angular Distance Tips");
  ANGULAR_TIPS.forEach((t) => line(t));
  line("Extend your arm as far as possible", "sb-note");

  y += 3.2;
  heading("Legend");
  const legendStarY = y + 3.6;
  const legendPathY = legendStarY + 6;

  return { items, legendStarY, legendPathY };
});

// ---------------------------------------------------------------------------
// Interactivity: hovering or focusing a star highlights it, its label, and
// any route segments touching it. The visible markers live in the SVG, but
// hover/focus detection and the info popup are handled by an HTML overlay
// (see template) using shadcn Tooltip — SVG has no native equivalent for
// that component, so plain <title> was swapped out for it.
// ---------------------------------------------------------------------------
const hoveredIndex = ref<number | null>(null);

function isTouching(seg: (typeof segments.value)[number]): boolean {
  return (
    hoveredIndex.value !== null &&
    (seg.from.index === hoveredIndex.value ||
      seg.to.index === hoveredIndex.value)
  );
}

// ---------------------------------------------------------------------------
// Action bar above the map. Select/measure are placeholder tool modes for
// now (just track which one is active).
// ---------------------------------------------------------------------------
const activeTool = ref<"select" | "measure">("select");

// ---------------------------------------------------------------------------
// Printing.
//
// Printing the current page directly was pulling in the rest of the app
// (sidebar, header, toolbar) and, worse, duplicating the map across two
// pages — a `position: fixed` print-isolation trick repeats on every
// generated page, and the hidden-but-still-laid-out surrounding content was
// tall enough to trigger a second page. Opening a small standalone document
// with nothing but the map sidesteps both problems: there's nothing else to
// paginate, so it's always exactly one page.
//
// The trade-off: PRINT_STYLES duplicates the drawing rules from the scoped
// <style> block as plain (non-scoped) CSS text, since Vue's scoped attribute
// hashes mean nothing in a document this component didn't render. Keep the
// two in sync if the map's look changes.
// ---------------------------------------------------------------------------
const svgRef = ref<SVGSVGElement | null>(null);

const PRINT_STYLES = `
  @page { size: A4 landscape; margin: 0; }
  html, body { margin: 0; padding: 0; }
  svg { display: block; width: 297mm; height: 210mm; }
  .page-frame { fill: none; stroke: #000; stroke-width: 0.9; }
  .page-title { font: 700 7px Georgia, 'Times New Roman', serif; letter-spacing: 0.3px; fill: #000; }
  .grid-dashed circle, .grid-dashed line { stroke: #999; stroke-width: 0.15; stroke-dasharray: 0.6, 0.8; }
  .axis-line { stroke: #000; stroke-width: 0.2; }
  .horizon-circle { fill: none; stroke: #000; stroke-width: 0.5; }
  .alt-label, .az-label { font: 400 2.2px Arial, Helvetica, sans-serif; fill: #777; text-anchor: start; }
  .cardinal-label { font: 700 4px Georgia, 'Times New Roman', serif; fill: #000; text-anchor: middle; dominant-baseline: middle; }
  .route-line { stroke: #000; stroke-width: 0.3; stroke-dasharray: 1.3, 1; }
  .distance-label { font: 400 2.6px Arial, Helvetica, sans-serif; fill: #000; text-anchor: middle; paint-order: stroke; stroke: #fff; stroke-width: 1px; stroke-linejoin: round; }
  .star-glyph { fill: #000; }
  .star-label { font: 400 2.8px Georgia, 'Times New Roman', serif; fill: #000; text-anchor: middle; }
  .sb-heading { font: 700 4px Georgia, 'Times New Roman', serif; fill: #000; }
  .sb-body { font: 400 3.4px Arial, Helvetica, sans-serif; fill: #1a1a1a; }
  .sb-note { font: italic 400 3px Arial, Helvetica, sans-serif; fill: #666; }
  .legend-label { font: 400 3.2px Arial, Helvetica, sans-serif; fill: #1a1a1a; }
`;

function triggerPrint(): void {
  const svgEl = svgRef.value;
  if (!svgEl) return;

  const printWindow = window.open("", "_blank", "width=1200,height=850");
  if (!printWindow) {
    // Popup blocked — fall back to printing the current page rather than doing nothing.
    window.print();
    return;
  }

  printWindow.document.open();
  printWindow.document.write(
    `<!doctype html><html><head><meta charset="utf-8" /><title>${props.title} — star map</title>` +
      `<style>${PRINT_STYLES}</style></head><body>${svgEl.outerHTML}</body></html>`,
  );
  printWindow.document.close();

  printWindow.onload = () => {
    printWindow.focus();
    printWindow.print();
  };
}
</script>

<template>
  <div class="mx-auto w-full max-w-[1200px]">
    <TooltipProvider :delay-duration="150">
      <div class="overflow-hidden rounded-xl border border-border shadow-sm">
        <div
          class="flex items-center justify-between gap-1 border-b border-border bg-card p-1"
        >
          <div class="flex items-center gap-1">
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 rounded-md"
                  :class="
                    activeTool === 'select'
                      ? 'bg-accent text-accent-foreground'
                      : 'text-muted-foreground'
                  "
                  aria-label="Select"
                  :aria-pressed="activeTool === 'select'"
                  @click="activeTool = 'select'"
                >
                  <MousePointer2 class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Select</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 rounded-md"
                  :class="
                    activeTool === 'measure'
                      ? 'bg-accent text-accent-foreground'
                      : 'text-muted-foreground'
                  "
                  aria-label="Measure"
                  :aria-pressed="activeTool === 'measure'"
                  @click="activeTool = 'measure'"
                >
                  <PencilRuler class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Measure</TooltipContent>
            </Tooltip>
          </div>

          <div class="flex items-center gap-1">
            <div class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 rounded-md text-muted-foreground"
                  aria-label="Print or save as PDF"
                  @click="triggerPrint"
                >
                  <Printer class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Print / Save as PDF</TooltipContent>
            </Tooltip>
          </div>
        </div>

        <div class="relative">
          <svg
            ref="svgRef"
            class="starmap-page block aspect-[297/210] w-full bg-white"
            viewBox="0 0 297 210"
            xmlns="http://www.w3.org/2000/svg"
            role="img"
            :aria-label="`Star-hopping map for ${title}`"
          >
            <rect x="0" y="0" width="297" height="210" fill="white" />
            <rect
              :x="frame.x"
              :y="frame.y"
              :width="frame.w"
              :height="frame.h"
              class="page-frame"
              rx="2"
            />

            <text :x="frame.x + 4" :y="MARGIN + 10.5" class="page-title">
              {{ title.toUpperCase() }}
            </text>

            <!-- polar grid -->
            <g class="grid-dashed">
              <circle
                v-for="alt in ALT_RINGS"
                :key="`ring-${alt}`"
                :cx="center.x"
                :cy="center.y"
                :r="ringRadius(alt)"
                fill="none"
              />
              <line
                v-for="az in AZ_TICKS.filter((a) => !CARDINALS[a])"
                :key="`az-${az}`"
                :x1="center.x"
                :y1="center.y"
                :x2="edgeXY(az, 0).x"
                :y2="edgeXY(az, 0).y"
              />
            </g>

            <!-- solid N-S / E-W axes -->
            <line
              :x1="edgeXY(0, 0).x"
              :y1="edgeXY(0, 0).y"
              :x2="edgeXY(180, 0).x"
              :y2="edgeXY(180, 0).y"
              class="axis-line"
            />
            <line
              :x1="edgeXY(90, 0).x"
              :y1="edgeXY(90, 0).y"
              :x2="edgeXY(270, 0).x"
              :y2="edgeXY(270, 0).y"
              class="axis-line"
            />

            <!-- horizon boundary + zenith mark -->
            <circle
              :cx="center.x"
              :cy="center.y"
              :r="radius"
              class="horizon-circle"
            />
            <circle :cx="center.x" :cy="center.y" r="0.6" fill="black" />

            <!-- altitude ring labels, along the north line -->
            <text
              v-for="alt in ALT_RINGS"
              :key="`alt-label-${alt}`"
              :x="center.x + 2.2"
              :y="center.y - ringRadius(alt) + 1"
              class="alt-label"
            >
              {{ alt }}°
            </text>

            <!-- azimuth labels around the rim -->
            <text
              v-for="az in AZ_TICKS"
              :key="`az-label-${az}`"
              :x="edgeXY(az, CARDINALS[az] ? 6.5 : 5).x"
              :y="edgeXY(az, CARDINALS[az] ? 6.5 : 5).y"
              :class="CARDINALS[az] ? 'cardinal-label' : 'az-label'"
            >
              {{ CARDINALS[az] ?? `${az}°` }}
            </text>

            <!-- star-hop route -->
            <g v-for="(seg, k) in segments" :key="`seg-${k}`">
              <line
                :x1="seg.from.x"
                :y1="seg.from.y"
                :x2="seg.to.x"
                :y2="seg.to.y"
                class="route-line"
                :class="{ 'is-active': isTouching(seg) }"
              />
              <text
                :x="(seg.from.x + seg.to.x) / 2"
                :y="(seg.from.y + seg.to.y) / 2 - 0.6"
                class="distance-label"
                :class="{ 'is-active': isTouching(seg) }"
              >
                {{ seg.distance != null ? `${seg.distance.toFixed(1)}°` : "—" }}
              </text>
            </g>

            <!-- stars: plain circle markers. Hover/focus state is driven by the
                 HTML tooltip overlay below, not by pointer events on the SVG. -->
            <g v-for="pt in points" :key="`star-${pt.index}`">
              <circle
                :cx="pt.x"
                :cy="pt.y"
                r="1"
                class="star-glyph"
                :class="{ 'is-hovered': hoveredIndex === pt.index }"
              />
              <text
                :x="pt.x"
                :y="pt.y - 2.6"
                class="star-label"
                :class="{ 'is-hovered': hoveredIndex === pt.index }"
              >
                {{ pt.name }}
              </text>
            </g>

            <!-- sidebar: Angular Distance Tips + Legend, single column -->
            <text
              v-for="(item, i) in sidebar.items"
              :key="`sb-${i}`"
              :x="item.x"
              :y="item.y"
              :class="item.cls"
            >
              {{ item.text }}
            </text>

            <circle
              :cx="sidebarX + 2"
              :cy="sidebar.legendStarY - 0.6"
              r="1.4"
              class="star-glyph"
            />
            <text
              :x="sidebarX + 6"
              :y="sidebar.legendStarY"
              class="legend-label"
            >
              Star
            </text>
            <line
              :x1="sidebarX"
              :y1="sidebar.legendPathY"
              :x2="sidebarX + 4"
              :y2="sidebar.legendPathY"
              class="route-line"
            />
            <text
              :x="sidebarX + 6"
              :y="sidebar.legendPathY + 0.7"
              class="legend-label"
            >
              Path
            </text>
          </svg>

          <!-- HTML hover/tooltip overlay for stars, aligned to the SVG viewBox by
               percentage position (the SVG's aspect-ratio is locked to 297:210,
               so percentage-of-container always lines up with viewBox coordinates). -->
          <div class="pointer-events-none absolute inset-0">
            <Tooltip v-for="pt in points" :key="`tip-${pt.index}`">
              <TooltipTrigger as-child>
                <button
                  type="button"
                  class="pointer-events-auto absolute size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-none bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  :style="{
                    left: `${(pt.x / PAGE_W) * 100}%`,
                    top: `${(pt.y / PAGE_H) * 100}%`,
                  }"
                  :aria-label="pt.name"
                  @pointerenter="hoveredIndex = pt.index"
                  @pointerleave="hoveredIndex = null"
                  @focus="hoveredIndex = pt.index"
                  @blur="hoveredIndex = null"
                />
              </TooltipTrigger>
              <TooltipContent
                >{{ pt.name }} — Alt {{ pt.alt.toFixed(1) }}°, Az
                {{ pt.az.toFixed(1) }}°</TooltipContent
              >
            </Tooltip>
          </div>
        </div>
      </div>
    </TooltipProvider>
  </div>
</template>

<style scoped>
.page-frame {
  fill: none;
  stroke: #000;
  stroke-width: 0.9;
}
.page-title {
  font:
    700 7px Georgia,
    "Times New Roman",
    serif;
  letter-spacing: 0.3px;
  fill: #000;
}

.grid-dashed circle,
.grid-dashed line {
  stroke: #999;
  stroke-width: 0.15;
  stroke-dasharray: 0.6, 0.8;
}
.axis-line {
  stroke: #000;
  stroke-width: 0.2;
}
.horizon-circle {
  fill: none;
  stroke: #000;
  stroke-width: 0.5;
}

.alt-label,
.az-label {
  font:
    400 2.2px Arial,
    Helvetica,
    sans-serif;
  fill: #777;
  text-anchor: start;
}
.cardinal-label {
  font:
    700 4px Georgia,
    "Times New Roman",
    serif;
  fill: #000;
  text-anchor: middle;
  dominant-baseline: middle;
}

.route-line {
  stroke: #000;
  stroke-width: 0.3;
  stroke-dasharray: 1.3, 1;
}
.route-line.is-active {
  stroke-width: 0.6;
}
.distance-label {
  font:
    400 2.6px Arial,
    Helvetica,
    sans-serif;
  fill: #000;
  text-anchor: middle;
  paint-order: stroke;
  stroke: #fff;
  stroke-width: 1px;
  stroke-linejoin: round;
}
.distance-label.is-active {
  font-weight: 700;
}

.star-glyph {
  fill: #000;
  transition: fill 0.1s ease;
}
.star-glyph.is-hovered {
  fill: #b8472d;
}
.star-label {
  font:
    400 2.8px Georgia,
    "Times New Roman",
    serif;
  fill: #000;
  text-anchor: middle;
}
.star-label.is-hovered {
  fill: #b8472d;
  font-weight: 700;
}

.sb-heading {
  font:
    700 4px Georgia,
    "Times New Roman",
    serif;
  fill: #000;
}
.sb-body {
  font:
    400 3.4px Arial,
    Helvetica,
    sans-serif;
  fill: #1a1a1a;
}
.sb-note {
  font:
    italic 400 3px Arial,
    Helvetica,
    sans-serif;
  fill: #666;
}
.legend-label {
  font:
    400 3.2px Arial,
    Helvetica,
    sans-serif;
  fill: #1a1a1a;
}
</style>
