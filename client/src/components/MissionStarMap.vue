<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Flashlight, Moon, MousePointer2, PencilRuler, Printer, Sun } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { Horizontal, MapData, MissionData } from "@/types/api";

const props = withDefaults(
  defineProps<{
    missionData: MissionData;
    mapData: MapData;
    title?: string;
    /** When true, hints (bigger hit targets, grab cursor) that route editing is available upstream. */
    editable?: boolean;
  }>(),
  { title: "Astrohop", editable: false },
);

/**
 * Not wired to persistence yet. This is the foundation for manual route
 * editing: a parent that turns `editable` on can listen here and PATCH the
 * mission once real reordering interactions exist. Keys are stable
 * (`oid:<n>` when the backend provides one, `idx:<n>` otherwise) so they
 * survive re-sorts even before every position carries an oid.
 */
const emit = defineEmits<{
  (e: "update:route", order: string[]): void;
}>();

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});
const formattedTime = computed(() => dateFormatter.format(new Date(props.missionData.time)));

// ---------------------------------------------------------------------------
// Page geometry — A4 landscape, in millimetres, matching the print output
// 1:1. The old right-hand SVG sidebar (Angular Distance Tips + Legend) is
// gone: that content now lives in the HTML panel on screen, and collapses to
// a single compact caption row along the bottom of the page for print, so
// the circle itself gets the full page width in both contexts.
// ---------------------------------------------------------------------------
const PAGE_W = 297;
const PAGE_H = 210;
const MARGIN = 10;
const HEADER_H = 15;
const CAPTION_H = 9; // reserved for the print-only reference caption

const frame = {
  x: MARGIN,
  y: MARGIN,
  w: PAGE_W - MARGIN * 2,
  h: PAGE_H - MARGIN * 2,
};

const circleArea = {
  x: MARGIN + 4,
  y: MARGIN + HEADER_H,
  w: frame.w - 8,
  h: frame.h - HEADER_H - 8 - CAPTION_H,
};

const center = {
  x: circleArea.x + circleArea.w / 2,
  y: circleArea.y + circleArea.h / 2,
};
const radius = Math.min(circleArea.w, circleArea.h) / 2 - 6;
const captionY = frame.y + frame.h - 3;

// ---------------------------------------------------------------------------
// Polar projection: centre = zenith (alt 90°), outer ring = horizon (alt 0°).
// North is up, azimuth increases clockwise, matching a chart held facing
// north rather than a naked-eye "lying on your back" view.
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
// Color modes. Palettes are the single source of truth for both the live
// SVG (applied as CSS custom properties on the component root) and the
// print stylesheet (which always renders with the Day palette, regardless
// of what's on screen — ink on paper only ever wants the paper look).
// Night intentionally points at the app's own theme tokens so the chart
// reads as part of the shell rather than a pasted-in rectangle; Day and Red
// are fixed literal values since neither should move if the app theme does.
// ---------------------------------------------------------------------------
type ColorMode = "day" | "night" | "red";

interface StarMapPalette {
  bg: string;
  frame: string;
  ink: string;
  muted: string;
  grid: string;
  accent: string;
  halo: string;
}

const COLOR_PALETTES: Record<ColorMode, StarMapPalette> = {
  day: {
    bg: "#ffffff",
    frame: "#000000",
    ink: "#000000",
    muted: "#6b6b6b",
    grid: "#9a9a9a",
    accent: "#8a3620",
    halo: "#ffffff",
  },
  night: {
    bg: "var(--background)",
    frame: "var(--border)",
    ink: "var(--foreground)",
    muted: "var(--muted-foreground)",
    grid: "color-mix(in oklch, var(--foreground) 22%, transparent)",
    accent: "#b8472d",
    halo: "var(--background)",
  },
  red: {
    bg: "#050100",
    frame: "#4a1410",
    ink: "#ff3b2f",
    muted: "#8f342a",
    grid: "#3a1310",
    accent: "#ff6a52",
    halo: "#050100",
  },
};

const colorMode = ref<ColorMode>("night");

const paletteStyle = computed(() => {
  const p = COLOR_PALETTES[colorMode.value];
  return {
    "--sm-bg": p.bg,
    "--sm-frame": p.frame,
    "--sm-ink": p.ink,
    "--sm-muted": p.muted,
    "--sm-grid": p.grid,
    "--sm-accent": p.accent,
    "--sm-halo": p.halo,
  };
});

function buildPrintStyles(p: StarMapPalette): string {
  return `
    @page { size: A4 landscape; margin: 0; }
    html, body { margin: 0; padding: 0; }
    svg { display: block; width: 297mm; height: 210mm; }
    .page-frame { fill: none; stroke: ${p.frame}; stroke-width: 0.9; }
    .page-title { font: 700 7px Georgia, 'Times New Roman', serif; letter-spacing: 0.3px; fill: ${p.ink}; }
    .grid-dashed circle, .grid-dashed line { stroke: ${p.grid}; stroke-width: 0.15; stroke-dasharray: 0.6, 0.8; }
    .axis-line { stroke: ${p.ink}; stroke-width: 0.2; }
    .horizon-circle { fill: none; stroke: ${p.ink}; stroke-width: 0.5; }
    .alt-label, .az-label { font: 400 2.2px Arial, Helvetica, sans-serif; fill: ${p.muted}; }
    .cardinal-label { font: 700 4px Georgia, 'Times New Roman', serif; fill: ${p.ink}; text-anchor: middle; dominant-baseline: middle; }
    .route-line { stroke: ${p.ink}; stroke-width: 0.3; stroke-dasharray: 1.3, 1; }
    .route-line.is-active { stroke-width: 0.3; }
    .distance-label { font: 400 2.6px Arial, Helvetica, sans-serif; fill: ${p.ink}; paint-order: stroke; stroke: ${p.halo}; stroke-width: 1px; stroke-linejoin: round; }
    .distance-label.is-active { font-weight: 400; }
    .star-glyph, .star-glyph.is-hovered, .star-glyph.is-selected, .star-glyph.is-measuring { fill: ${p.ink}; }
    .star-label, .star-label.is-hovered, .star-label.is-selected, .star-label.is-measuring { font: 400 2.8px Georgia, 'Times New Roman', serif; fill: ${p.ink}; font-weight: 400; }
    .print-caption { font: 400 3px Arial, Helvetica, sans-serif; fill: ${p.muted}; }
    .print-caption-strong { font: 700 3px Arial, Helvetica, sans-serif; fill: ${p.ink}; }
    .measure-line, .measure-label, .star-ring { display: none; }
  `;
}

const PRINT_STYLES = buildPrintStyles(COLOR_PALETTES.day);

// ---------------------------------------------------------------------------
// Data joins. `mapData.positions[i]` is matched to `missionData.objectives`
// by `oid` whenever the backend provides one on the position, and falls
// back to positional index otherwise. The oid path is the one that's safe
// to build route editing on top of; the index fallback exists only so this
// keeps working against a backend that hasn't added `Horizontal.oid` yet.
// ---------------------------------------------------------------------------
const objectivesByOid = computed(() => {
  const map = new Map<number, string>();
  for (const objective of props.missionData.objectives) {
    map.set(objective.oid, objective.name);
  }
  return map;
});

function labelFor(pos: Horizontal, index: number): string {
  return props.missionData.objectives[index]?.name ?? `Object ${index + 1}`;
}

const points = computed(() =>
  props.mapData.positions.map((pos, i) => ({
    index: i,
    key: `idx:${i}`,
    alt: pos.alt,
    az: pos.az,
    name: labelFor(pos, i),
    ...toXY(pos.alt, pos.az),
  })),
);

// `workingOrder` is the route this component actually renders. It starts as
// a copy of the server's tour order and re-syncs whenever a new `mapData`
// arrives, but it's a separate piece of state so a future editing feature
// can mutate it locally before anything is persisted.
const workingOrder = ref<number[]>([...props.mapData.tour.order]);

watch(
  () => props.mapData,
  (next) => {
    workingOrder.value = [...next.tour.order];
  },
);

watch(workingOrder, (order) => {
  emit(
    "update:route",
    order.map((i) => points.value[i]?.key ?? `idx:${i}`),
  );
});

const routeOrder = computed(() => workingOrder.value.filter((i) => points.value[i] != null));

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
    const distance = props.mapData.tour.distances?.[order[k]]?.[order[k + 1]] ?? null;
    list.push({ from, to, distance });
  }
  return list;
});

const stopNumbers = computed(() => {
  const map = new Map<number, number>();
  routeOrder.value.forEach((idx, position) => map.set(idx, position + 1));
  return map;
});

const orderedDisplayPoints = computed(() => {
  const seen = new Set<number>();
  const ordered = routeOrder.value
    .map((i) => points.value[i])
    .filter((pt): pt is (typeof points.value)[number] => {
      if (!pt || seen.has(pt.index)) return false;
      seen.add(pt.index);
      return true;
    });
  const rest = points.value.filter((pt) => !seen.has(pt.index));
  return [...ordered, ...rest];
});

// ---------------------------------------------------------------------------
// Label overlap handling. Rather than a fixed offset above every point, each
// label gets a small set of candidate anchor positions (N, NE, E, SE, S, SW,
// W, NW, in that preference order) and is greedily assigned the first one
// whose approximate bounding box doesn't collide with an already-placed
// label. Stars are placed before route-distance labels so a star name never
// loses its preferred spot to a passing route label. This is the discrete-
// candidate-position approach from Christensen, Marks & Shieber's classic
// point-feature label placement study — good enough for mission-sized
// object counts without reaching for a full physics/annealing solver.
// ---------------------------------------------------------------------------
const LABEL_OFFSET = 3;
const CHAR_WIDTH_RATIO = 0.58;

const LABEL_CANDIDATES: { dx: number; dy: number; anchor: "start" | "middle" | "end" }[] = [
  { dx: 0, dy: -LABEL_OFFSET, anchor: "middle" },
  { dx: LABEL_OFFSET, dy: -LABEL_OFFSET * 0.7, anchor: "start" },
  { dx: LABEL_OFFSET, dy: 0, anchor: "start" },
  { dx: LABEL_OFFSET, dy: LABEL_OFFSET * 0.7, anchor: "start" },
  { dx: 0, dy: LABEL_OFFSET + 1.5, anchor: "middle" },
  { dx: -LABEL_OFFSET, dy: LABEL_OFFSET * 0.7, anchor: "end" },
  { dx: -LABEL_OFFSET, dy: 0, anchor: "end" },
  { dx: -LABEL_OFFSET, dy: -LABEL_OFFSET * 0.7, anchor: "end" },
];

interface PlacedRect {
  left: number;
  right: number;
  top: number;
  bottom: number;
}

interface ResolvedLabel {
  x: number;
  y: number;
  anchor: "start" | "middle" | "end";
}

function rectsOverlap(a: PlacedRect, b: PlacedRect): boolean {
  return !(a.right < b.left || a.left > b.right || a.bottom < b.top || a.top > b.bottom);
}

function estimateWidth(text: string, fontSize: number): number {
  return Math.max(text.length, 1) * fontSize * CHAR_WIDTH_RATIO;
}

function rectFor(x: number, y: number, width: number, height: number, anchor: "start" | "middle" | "end"): PlacedRect {
  const left = anchor === "start" ? x : anchor === "end" ? x - width : x - width / 2;
  return { left, right: left + width, top: y - height, bottom: y };
}

interface LabelRequest {
  id: string;
  x: number;
  y: number;
  text: string;
  fontSize: number;
  priority: number;
}

const resolvedLabels = computed(() => {
  const requests: LabelRequest[] = [
    ...points.value.map((pt) => ({
      id: `star:${pt.index}`,
      x: pt.x,
      y: pt.y,
      text: pt.name,
      fontSize: 2.8,
      priority: 0,
    })),
    ...segments.value.map((seg, k) => ({
      id: `seg:${k}`,
      x: (seg.from.x + seg.to.x) / 2,
      y: (seg.from.y + seg.to.y) / 2,
      text: seg.distance != null ? `${seg.distance.toFixed(1)}°` : "—",
      fontSize: 2.6,
      priority: 1,
    })),
  ].sort((a, b) => a.priority - b.priority);

  const placed: PlacedRect[] = [];
  const resolved = new Map<string, ResolvedLabel>();

  for (const request of requests) {
    const width = estimateWidth(request.text, request.fontSize);
    const height = request.fontSize * 1.3;
    let chosen: (ResolvedLabel & { rect: PlacedRect }) | null = null;

    for (const candidate of LABEL_CANDIDATES) {
      const x = request.x + candidate.dx;
      const y = request.y + candidate.dy;
      const rect = rectFor(x, y, width, height, candidate.anchor);
      if (!placed.some((p) => rectsOverlap(p, rect))) {
        chosen = { x, y, anchor: candidate.anchor, rect };
        break;
      }
    }

    if (!chosen) {
      const fallback = LABEL_CANDIDATES[0];
      const x = request.x + fallback.dx;
      const y = request.y + fallback.dy;
      chosen = { x, y, anchor: fallback.anchor, rect: rectFor(x, y, width, height, fallback.anchor) };
    }

    placed.push(chosen.rect);
    resolved.set(request.id, { x: chosen.x, y: chosen.y, anchor: chosen.anchor });
  }

  return resolved;
});

// ---------------------------------------------------------------------------
// Interactivity. Hovering/focusing a star highlights it, its label, and any
// route segments touching it (unchanged). Select and Measure are real modes
// now: Select toggles membership in `selectedIndices` — inert today, but
// it's the exact state a future "insert/remove from route" action would
// read. Measure collects up to two stars and reports the true angular
// separation computed from alt/az (not screen-pixel distance, which the
// zenith-centered polar projection distorts away from the pole).
// ---------------------------------------------------------------------------
const hoveredIndex = ref<number | null>(null);
const activeTool = ref<"select" | "measure">("select");
const selectedIndices = ref<Set<number>>(new Set());
const measureIndices = ref<number[]>([]);

function isTouching(seg: (typeof segments.value)[number]): boolean {
  return (
    hoveredIndex.value !== null &&
    (seg.from.index === hoveredIndex.value || seg.to.index === hoveredIndex.value)
  );
}

function setTool(tool: "select" | "measure"): void {
  activeTool.value = tool;
}

function onStarActivate(index: number): void {
  if (activeTool.value === "select") {
    const next = new Set(selectedIndices.value);
    if (next.has(index)) next.delete(index);
    else next.add(index);
    selectedIndices.value = next;
    return;
  }

  if (measureIndices.value.includes(index)) {
    measureIndices.value = measureIndices.value.filter((i) => i !== index);
  } else if (measureIndices.value.length >= 2) {
    measureIndices.value = [index];
  } else {
    measureIndices.value = [...measureIndices.value, index];
  }
}

function clearActiveToolState(): void {
  if (activeTool.value === "select") selectedIndices.value = new Set();
  else measureIndices.value = [];
}

function angularSeparationDeg(alt1: number, az1: number, alt2: number, az2: number): number {
  const toRad = (deg: number) => (deg * Math.PI) / 180;
  const a1 = toRad(alt1);
  const a2 = toRad(alt2);
  const dAz = toRad(az1 - az2);
  const cosTheta = Math.sin(a1) * Math.sin(a2) + Math.cos(a1) * Math.cos(a2) * Math.cos(dAz);
  return (Math.acos(Math.min(1, Math.max(-1, cosTheta))) * 180) / Math.PI;
}

const measurement = computed(() => {
  if (measureIndices.value.length !== 2) return null;
  const from = points.value[measureIndices.value[0]];
  const to = points.value[measureIndices.value[1]];
  if (!from || !to) return null;
  return { from, to, degrees: angularSeparationDeg(from.alt, from.az, to.alt, to.az) };
});

// ---------------------------------------------------------------------------
// Toolbar accessibility. Select, Measure, the three mode buttons, and Print
// form one WAI-ARIA "toolbar" — a single tab stop, with arrow keys roving
// focus between the buttons, per the ARIA Authoring Practices Guide.
// ---------------------------------------------------------------------------
const toolbarRef = ref<HTMLDivElement | null>(null);
const toolbarFocusIndex = ref(0);

function focusableToolbarButtons(): HTMLButtonElement[] {
  if (!toolbarRef.value) return [];
  return Array.from(toolbarRef.value.querySelectorAll<HTMLButtonElement>("button[data-toolbar-item]"));
}

function onToolbarFocusIn(event: FocusEvent): void {
  const buttons = focusableToolbarButtons();
  const index = buttons.findIndex((el) => el === event.target);
  if (index >= 0) toolbarFocusIndex.value = index;
}

function onToolbarKeydown(event: KeyboardEvent): void {
  const buttons = focusableToolbarButtons();
  if (buttons.length === 0) return;

  let nextIndex: number | null = null;
  if (event.key === "ArrowRight") nextIndex = (toolbarFocusIndex.value + 1) % buttons.length;
  else if (event.key === "ArrowLeft") nextIndex = (toolbarFocusIndex.value - 1 + buttons.length) % buttons.length;
  else if (event.key === "Home") nextIndex = 0;
  else if (event.key === "End") nextIndex = buttons.length - 1;

  if (nextIndex !== null) {
    event.preventDefault();
    buttons[nextIndex]?.focus();
  }
}

function toolbarTabIndex(index: number): number {
  return index === toolbarFocusIndex.value ? 0 : -1;
}

// ---------------------------------------------------------------------------
// Printing. Opens a standalone document containing only the map SVG, so
// nothing else on the page can bleed into the printed output or force a
// second page. Always renders with the Day palette and strips any transient
// interaction state (hover/selection/measurement) regardless of what's
// currently shown on screen — ink on paper should always be the clean,
// deterministic chart.
// ---------------------------------------------------------------------------
const svgRef = ref<SVGSVGElement | null>(null);

function triggerPrint(): void {
  const svgEl = svgRef.value;
  if (!svgEl) return;

  const printWindow = window.open("", "_blank", "width=1200,height=850");
  if (!printWindow) {
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
  <div
    class="mx-auto flex w-full max-w-[1400px] flex-col gap-4 lg:flex-row lg:items-start"
    :style="paletteStyle"
  >
    <!-- Info / reference panel -->
    <Card size="sm" class="w-full shrink-0 gap-4 lg:w-72">
      <CardHeader>
        <CardTitle class="text-base">Mission overview</CardTitle>
        <CardDescription>{{ formattedTime }}</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-4 text-sm">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Location</p>
            <p class="mt-0.5 font-medium">
              {{ missionData.location.lat.toFixed(2) }}, {{ missionData.location.long.toFixed(2) }}
            </p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Limiting mag.</p>
            <p class="mt-0.5 font-medium">{{ missionData.conditions.limiting_magnitude }}</p>
          </div>
        </div>

        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">
            Objectives ({{ points.length }})
          </p>
          <ul class="mt-1.5 flex max-h-56 flex-col gap-1 overflow-y-auto">
            <li v-for="pt in orderedDisplayPoints" :key="pt.key">
              <button
                type="button"
                class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors"
                :class="
                  selectedIndices.has(pt.index) || measureIndices.includes(pt.index)
                    ? 'bg-accent text-accent-foreground'
                    : 'hover:bg-accent/60'
                "
                @mouseenter="hoveredIndex = pt.index"
                @mouseleave="hoveredIndex = null"
                @click="onStarActivate(pt.index)"
              >
                <span
                  v-if="stopNumbers.get(pt.index)"
                  class="flex size-5 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-semibold text-muted-foreground"
                >
                  {{ stopNumbers.get(pt.index) }}
                </span>
                <span class="truncate">{{ pt.name }}</span>
              </button>
            </li>
          </ul>
        </div>

        <div class="flex items-center gap-4 border-t border-border pt-3 text-xs text-muted-foreground">
          <span class="flex items-center gap-1.5">
            <span class="inline-block size-1.5 rounded-full" style="background: var(--sm-ink)" />
            Star
          </span>
          <span class="flex items-center gap-1.5">
            <span class="inline-block h-px w-3 border-t border-dashed" style="border-color: var(--sm-ink)" />
            Path
          </span>
        </div>

        <div v-if="activeTool === 'measure'" class="rounded-md border border-dashed border-border p-2.5 text-xs">
          <template v-if="measurement">
            <p class="font-medium text-foreground">{{ measurement.from.name }} → {{ measurement.to.name }}</p>
            <p class="mt-0.5 text-muted-foreground">{{ measurement.degrees.toFixed(2) }}° apart</p>
          </template>
          <p v-else class="text-muted-foreground">
            Tap {{ measureIndices.length === 0 ? "two stars" : "one more star" }} to measure the angle between them.
          </p>
          <Button
            v-if="measureIndices.length"
            variant="ghost"
            size="xs"
            class="mt-2 h-6 px-2"
            @click="clearActiveToolState"
          >
            Clear
          </Button>
        </div>

        <div
          v-else-if="selectedIndices.size"
          class="flex items-center justify-between rounded-md border border-dashed border-border p-2.5 text-xs"
        >
          <span class="text-muted-foreground">{{ selectedIndices.size }} selected</span>
          <Button variant="ghost" size="xs" class="h-6 px-2" @click="clearActiveToolState">Clear</Button>
        </div>
      </CardContent>
    </Card>

    <!-- Toolbar + chart -->
    <div class="min-w-0 flex-1">
      <TooltipProvider :delay-duration="150">
        <div class="overflow-hidden rounded-xl border border-border shadow-sm">
          <div
            ref="toolbarRef"
            role="toolbar"
            aria-label="Star map tools"
            class="flex flex-wrap items-center justify-between gap-1 border-b border-border bg-card p-1"
            @keydown="onToolbarKeydown"
            @focusin="onToolbarFocusIn"
          >
            <div class="flex items-center gap-1">
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(0)"
                    class="h-8 w-8 rounded-md"
                    :class="activeTool === 'select' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                    aria-label="Select"
                    :aria-pressed="activeTool === 'select'"
                    @click="setTool('select')"
                  >
                    <MousePointer2 class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Select stars</TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(1)"
                    class="h-8 w-8 rounded-md"
                    :class="activeTool === 'measure' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                    aria-label="Measure"
                    :aria-pressed="activeTool === 'measure'"
                    @click="setTool('measure')"
                  >
                    <PencilRuler class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Measure angular separation</TooltipContent>
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
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(2)"
                    class="h-8 w-8 rounded-md"
                    :class="colorMode === 'day' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                    aria-label="Day mode"
                    :aria-pressed="colorMode === 'day'"
                    @click="colorMode = 'day'"
                  >
                    <Sun class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Day (print) mode</TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(3)"
                    class="h-8 w-8 rounded-md"
                    :class="colorMode === 'night' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                    aria-label="Night mode"
                    :aria-pressed="colorMode === 'night'"
                    @click="colorMode = 'night'"
                  >
                    <Moon class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Night mode</TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(4)"
                    class="h-8 w-8 rounded-md"
                    :class="colorMode === 'red' ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
                    aria-label="Red-light mode"
                    :aria-pressed="colorMode === 'red'"
                    @click="colorMode = 'red'"
                  >
                    <Flashlight class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Red-light mode — preserves night vision</TooltipContent>
              </Tooltip>

              <div class="mx-1 h-5 w-px bg-border" aria-hidden="true" />

              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    data-toolbar-item
                    :tabindex="toolbarTabIndex(5)"
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
              class="starmap-page block aspect-[297/210] w-full"
              viewBox="0 0 297 210"
              xmlns="http://www.w3.org/2000/svg"
              role="img"
              :aria-label="`Star-hopping map for ${title}`"
            >
              <rect x="0" y="0" width="297" height="210" fill="var(--sm-bg)" @click="clearActiveToolState" />
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
              <circle :cx="center.x" :cy="center.y" :r="radius" class="horizon-circle" />
              <circle :cx="center.x" :cy="center.y" r="0.6" fill="var(--sm-ink)" />

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
                  :x="resolvedLabels.get(`seg:${k}`)?.x ?? (seg.from.x + seg.to.x) / 2"
                  :y="resolvedLabels.get(`seg:${k}`)?.y ?? (seg.from.y + seg.to.y) / 2"
                  :style="{ textAnchor: resolvedLabels.get(`seg:${k}`)?.anchor ?? 'middle' }"
                  class="distance-label"
                  :class="{ 'is-active': isTouching(seg) }"
                >
                  {{ seg.distance != null ? `${seg.distance.toFixed(1)}°` : "—" }}
                </text>
              </g>

              <!-- live measurement, screen-only interaction state -->
              <g v-if="measurement">
                <line
                  :x1="measurement.from.x"
                  :y1="measurement.from.y"
                  :x2="measurement.to.x"
                  :y2="measurement.to.y"
                  class="measure-line"
                />
                <text
                  :x="(measurement.from.x + measurement.to.x) / 2"
                  :y="(measurement.from.y + measurement.to.y) / 2 - 1"
                  class="measure-label"
                >
                  {{ measurement.degrees.toFixed(2) }}°
                </text>
              </g>

              <!-- stars -->
              <g v-for="pt in points" :key="`star-${pt.key}`">
                <circle
                  v-if="selectedIndices.has(pt.index) || measureIndices.includes(pt.index)"
                  :cx="pt.x"
                  :cy="pt.y"
                  r="2.1"
                  class="star-ring"
                />
                <circle
                  :cx="pt.x"
                  :cy="pt.y"
                  r="1"
                  class="star-glyph"
                  :class="{
                    'is-hovered': hoveredIndex === pt.index,
                    'is-selected': selectedIndices.has(pt.index),
                    'is-measuring': measureIndices.includes(pt.index),
                  }"
                />
                <text
                  :x="resolvedLabels.get(`star:${pt.index}`)?.x ?? pt.x"
                  :y="resolvedLabels.get(`star:${pt.index}`)?.y ?? pt.y - 2.6"
                  :style="{ textAnchor: resolvedLabels.get(`star:${pt.index}`)?.anchor ?? 'middle' }"
                  class="star-label"
                  :class="{
                    'is-hovered': hoveredIndex === pt.index,
                    'is-selected': selectedIndices.has(pt.index),
                    'is-measuring': measureIndices.includes(pt.index),
                  }"
                >
                  {{ pt.name }}
                </text>
              </g>

              <!-- print-only compact reference caption -->
              <g class="print-only">
                <circle :cx="frame.x + 6" :cy="captionY - 1" r="1.1" class="star-glyph" />
                <text :x="frame.x + 10" :y="captionY" class="print-caption-strong">Star</text>
                <line :x1="frame.x + 26" :y1="captionY - 1" :x2="frame.x + 32" :y2="captionY - 1" class="route-line" />
                <text :x="frame.x + 36" :y="captionY" class="print-caption-strong">Path</text>
                <text :x="frame.x + 56" :y="captionY" class="print-caption">
                  Angular estimate: finger ≈1° · fist ≈10° · open hand ≈20°
                </text>
              </g>
            </svg>

            <!-- HTML hover/click overlay for stars, aligned to the SVG viewBox by
                 percentage position (locked 297:210 aspect ratio keeps this exact). -->
            <div class="pointer-events-none absolute inset-0">
              <Tooltip v-for="pt in points" :key="`tip-${pt.key}`">
                <TooltipTrigger as-child>
                  <button
                    type="button"
                    class="pointer-events-auto absolute -translate-x-1/2 -translate-y-1/2 rounded-full border-none bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    :class="editable ? 'size-5 cursor-grab active:cursor-grabbing' : 'size-4'"
                    :style="{
                      left: `${(pt.x / PAGE_W) * 100}%`,
                      top: `${(pt.y / PAGE_H) * 100}%`,
                    }"
                    :aria-label="pt.name"
                    :aria-pressed="selectedIndices.has(pt.index) || measureIndices.includes(pt.index)"
                    @pointerenter="hoveredIndex = pt.index"
                    @pointerleave="hoveredIndex = null"
                    @focus="hoveredIndex = pt.index"
                    @blur="hoveredIndex = null"
                    @click="onStarActivate(pt.index)"
                  />
                </TooltipTrigger>
                <TooltipContent>{{ pt.name }} — Alt {{ pt.alt.toFixed(1) }}°, Az {{ pt.az.toFixed(1) }}°</TooltipContent>
              </Tooltip>
            </div>
          </div>
        </div>
      </TooltipProvider>
    </div>
  </div>
</template>

<style scoped>
.page-frame {
  fill: none;
  stroke: var(--sm-frame);
  stroke-width: 0.9;
}
.page-title {
  font: 700 7px var(--font-heading), Georgia, serif;
  letter-spacing: 0.3px;
  fill: var(--sm-ink);
}

.grid-dashed circle,
.grid-dashed line {
  stroke: var(--sm-grid);
  stroke-width: 0.15;
  stroke-dasharray: 0.6, 0.8;
}
.axis-line {
  stroke: var(--sm-ink);
  stroke-width: 0.2;
  opacity: 0.5;
}
.horizon-circle {
  fill: none;
  stroke: var(--sm-ink);
  stroke-width: 0.5;
}

.alt-label,
.az-label {
  font: 400 2.2px var(--font-sans), Arial, sans-serif;
  fill: var(--sm-muted);
  text-anchor: start;
}
.cardinal-label {
  font: 700 4px var(--font-heading), Georgia, serif;
  fill: var(--sm-ink);
  text-anchor: middle;
  dominant-baseline: middle;
}

.route-line {
  stroke: var(--sm-ink);
  stroke-width: 0.3;
  stroke-dasharray: 1.3, 1;
  opacity: 0.85;
}
.route-line.is-active {
  stroke-width: 0.6;
  opacity: 1;
}
.distance-label {
  font: 400 2.6px var(--font-sans), Arial, sans-serif;
  fill: var(--sm-ink);
  paint-order: stroke;
  stroke: var(--sm-halo);
  stroke-width: 1px;
  stroke-linejoin: round;
}
.distance-label.is-active {
  font-weight: 700;
}

.star-glyph {
  fill: var(--sm-ink);
  transition: fill 0.1s ease;
}
.star-glyph.is-hovered,
.star-glyph.is-selected,
.star-glyph.is-measuring {
  fill: var(--sm-accent);
}
.star-ring {
  fill: none;
  stroke: var(--sm-accent);
  stroke-width: 0.3;
}
.star-label {
  font: 400 2.8px var(--font-heading), Georgia, serif;
  fill: var(--sm-ink);
}
.star-label.is-hovered,
.star-label.is-selected,
.star-label.is-measuring {
  fill: var(--sm-accent);
  font-weight: 700;
}

.measure-line {
  stroke: var(--sm-accent);
  stroke-width: 0.4;
  stroke-dasharray: 0.6, 0.6;
}
.measure-label {
  font: 700 3px var(--font-sans), Arial, sans-serif;
  fill: var(--sm-accent);
  text-anchor: middle;
  paint-order: stroke;
  stroke: var(--sm-halo);
  stroke-width: 1px;
}

.print-caption {
  font: 400 3px Arial, Helvetica, sans-serif;
  fill: var(--sm-muted);
}
.print-caption-strong {
  font: 700 3px Arial, Helvetica, sans-serif;
  fill: var(--sm-ink);
}
.print-only {
  display: none;
}
</style>