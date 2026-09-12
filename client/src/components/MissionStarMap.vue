<script setup lang="ts">
import { computed, ref, watch } from "vue";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { Horizontal, Mission } from "@/types/api";
import { Pencil, Printer, Ruler, Scissors } from "@lucide/vue";
import Button from "./ui/button/Button.vue";

const props = defineProps<{
  mission: Mission;
}>();

const emit = defineEmits<{
  (e: "update:tour", edges: { from: number; to: number }[]): void;
}>();

const SIZE = 200;
const MARGIN = 8;
const center = { x: SIZE / 2, y: SIZE / 2 };
const radius = SIZE / 2 - MARGIN - 14;

const ALT_RINGS = [20, 40, 60, 80];
const AZ_TICKS = [0, 45, 90, 135, 180, 225, 270, 315];
const CARDINALS: Record<number, string> = {
  0: "N",
  90: "E",
  180: "S",
  270: "W",
};

const ANGULAR_TIPS = [
  "Finger width ~ 1°",
  "Three middle fingers ~ 5°",
  "Fist width ~ 10°",
  "Index to pinky ~ 15°",
  "Open hand width ~ 20°",
];

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});

function ringRadius(alt: number): number {
  if (alt < 0) {
    alt = 0 - alt;
  }
  return radius * (1 - Math.max(0, Math.min(90, alt)) / 90);
}

function toXY(h: Horizontal) {
  const r = ringRadius(h.alt);
  const theta = (h.az * Math.PI) / 180;
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

function angularSeparation(a: Horizontal, b: Horizontal): number {
  const toRad = (d: number) => (d * Math.PI) / 180;
  const alt1 = toRad(a.alt);
  const alt2 = toRad(b.alt);
  const dAz = toRad(a.az - b.az);
  const cosD =
    Math.sin(alt1) * Math.sin(alt2) +
    Math.cos(alt1) * Math.cos(alt2) * Math.cos(dAz);

  return (Math.acos(Math.min(1, Math.max(-1, cosD))) * 180) / Math.PI;
}

const objectives = computed(() => props.mission.data?.objectives ?? []);
const mapData = computed(() => props.mission.map_data ?? null);

const points = computed(() => {
  const positions = mapData.value?.positions;
  if (!positions) return [];
  return objectives.value.flatMap((objective) => {
    const pos = positions[objective.oid];
    if (!pos) return [];
    return [
      {
        oid: objective.oid,
        name: objective.name,
        alt: pos.alt,
        az: pos.az,
        ...toXY(pos),
      },
    ];
  });
});

const pointsByOid = computed(
  () => new Map(points.value.map((p) => [p.oid, p])),
);

/* ---------------------------------------------------------------------- */
/* Editable edge list (source of truth for the drawn route)                */
/* ---------------------------------------------------------------------- */

type Edge = { from: number; to: number };

function edgesFromTour(tour: number[]): Edge[] {
  const list: Edge[] = [];
  for (let i = 0; i < tour.length - 1; i++) {
    list.push({ from: tour[i], to: tour[i + 1] });
  }
  if (tour.length > 1) {
    list.push({ from: tour[tour.length - 1], to: tour[0] });
  }
  return list;
}

function edgeKey(e: Edge): string {
  return e.from < e.to ? `${e.from}:${e.to}` : `${e.to}:${e.from}`;
}

const manualEdges = ref<Edge[]>([]);

// Re-seed local edges whenever the server sends a new tour (e.g. after our
// own edits round-trip through the backend and come back persisted).
watch(
  () => mapData.value?.tour,
  (tour) => {
    manualEdges.value = tour ? edgesFromTour(tour) : [];
  },
  { immediate: true },
);

function commitEdges() {
  emit(
    "update:tour",
    manualEdges.value.map((e) => ({ ...e })),
  );
}

const segments = computed(() => {
  return manualEdges.value.flatMap((edge) => {
    const from = pointsByOid.value.get(edge.from);
    const to = pointsByOid.value.get(edge.to);
    if (!from || !to) return [];
    return [
      {
        key: edgeKey(edge),
        from,
        to,
        distance: angularSeparation(from, to),
      },
    ];
  });
});

const moon = computed(() => {
  const pos = mapData.value?.moon_position;
  if (!pos) return null;
  return { alt: pos.alt, az: pos.az, ...toXY(pos) };
});

const hoveredOid = ref<number | null>(null);

function isTouching(seg: (typeof segments.value)[number]): boolean {
  return (
    hoveredOid.value !== null &&
    (seg.from.oid === hoveredOid.value || seg.to.oid === hoveredOid.value)
  );
}

function getTextWidth(name: string): number {
  return name.length * 2.2;
}

/* ---------------------------------------------------------------------- */
/* Toolbar                                                                 */
/* ---------------------------------------------------------------------- */

type ToolId = "ruler" | "scissors" | "pencil";

const activeTool = ref<ToolId | null>(null);
const pendingStars = ref<number[]>([]);
const measurement = ref<{
  from: (typeof points.value)[number];
  to: (typeof points.value)[number];
  distance: number;
} | null>(null);

function resetToolState() {
  pendingStars.value = [];
  measurement.value = null;
}

function selectTool(tool: ToolId) {
  activeTool.value = activeTool.value === tool ? null : tool;
  resetToolState();
}

function onStarClick(oid: number) {
  if (activeTool.value === "ruler") {
    pendingStars.value.push(oid);
    if (pendingStars.value.length === 2) {
      const from = pointsByOid.value.get(pendingStars.value[0]);
      const to = pointsByOid.value.get(pendingStars.value[1]);
      if (from && to) {
        measurement.value = { from, to, distance: angularSeparation(from, to) };
      }
      pendingStars.value = [];
    }
    return;
  }

  if (activeTool.value === "pencil") {
    if (pendingStars.value[0] === oid) return; // ignore re-clicking same star
    pendingStars.value.push(oid);
    if (pendingStars.value.length === 2) {
      const [from, to] = pendingStars.value;
      const exists = manualEdges.value.some(
        (e) => edgeKey(e) === edgeKey({ from, to }),
      );
      if (!exists) {
        manualEdges.value.push({ from, to });
        commitEdges();
      }
      pendingStars.value = [];
    }
  }
}

function onSegmentClick(seg: (typeof segments.value)[number]) {
  if (activeTool.value !== "scissors") return;
  manualEdges.value = manualEdges.value.filter((e) => edgeKey(e) !== seg.key);
  commitEdges();
}

/* ---------------------------------------------------------------------- */
/* Print                                                                   */
/*                                                                         */
/* The on-screen preview is a small square SVG styled for a dark UI — it   */
/* can't just be cloned for print: white glyphs disappear on paper, the    */
/* aspect ratio doesn't fit A4, and edit-mode state (hover, ruler, pending */
/* clicks) has no business on a physical page. Print instead builds a      */
/* dedicated black-on-white A4 page from the same alt/az data and opens it */
/* in a standalone popup — the same print-isolation approach already used  */
/* by MissionStarMap.vue, needed because `position: fixed` print hacks     */
/* repeat on every page and hidden surrounding UI can still force a second */
/* page.                                                                   */
/* ---------------------------------------------------------------------- */

const canPrint = computed(() => !!mapData.value);

const pageTitle = computed(
  () => `Mission ${dateFormatter.format(new Date(props.mission.created_at))}`,
);

const PRINT_STYLES = `
  @page { size: A4 landscape; margin: 0; }
  html, body { margin: 0; padding: 0; }
  svg { display: block; width: 297mm; height: 210mm; }
  .page-frame { fill: none; stroke: #000; stroke-width: 1.6; }
  .divider { stroke: #000; stroke-width: 0.3; }
  .grid-dashed circle, .grid-dashed line { fill: none; stroke: #999; stroke-width: 0.15; stroke-dasharray: 0.6, 0.8; }
  .axis-line { stroke: #000; stroke-width: 0.2; }
  .horizon-circle { fill: none; stroke: #000; stroke-width: 0.5; }
  .horizon-tick { stroke: #000; stroke-width: 0.5; }
  .cardinal-label { font: 700 4px 'MuseoModerno', monospace; fill: #000; text-anchor: middle; dominant-baseline: middle; }
  .route-line { fill: none; stroke: #000; stroke-width: 0.3; stroke-dasharray: 1.3, 1; }
  .star-glyph { fill: #000; }
  .star-glyph--below { fill: none; stroke: #000; stroke-width: 0.4; }
  .star-label { font: 400 2.8px 'LINE Seed JP', 'Inter', monospace; fill: #000; text-anchor: middle; paint-order: stroke; stroke: #fff; stroke-width: 1px; stroke-linejoin: round; }
  .sb-heading { font: 700 4px 'MuseoModerno', monospace; fill: #000; }
  .sb-body { font: 400 3.4px 'LINE Seed JP', 'Inter', monospace; fill: #1a1a1a; }
  .sb-note { font: italic 400 3px 'LINE Seed JP', 'Inter', monospace; fill: #666; }
  .logo-text { font: 700 6px 'MuseoModerno', monospace; fill: #000; letter-spacing: 0.5px; }
`;

function escapeXml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

const LOGO_WIDTH = 46; // ~ визуальная ширина значка ASTROHOP (мм)

function logoMarkup(
  x: number,
  y: number,
  anchor: "start" | "end" = "start",
): string {
  return (
    `<text x="${x.toFixed(2)}" y="${y.toFixed(2)}" ` +
    `text-anchor="${anchor}" dominant-baseline="central" ` +
    `class="logo-text">ASTROHOP</text>`
  );
}

function buildPrintSvg(): string {
  const pageW = 297;
  const pageH = 210;
  const margin = 12;
  const topGap = 18;
  const bottomGap = 18;
  const pad = 6; // внутренний отступ от рамки
  const gap = 8; // зазор между колонками

  // Острая рамка — защита от пальцев
  const frame = {
    x: margin,
    y: topGap,
    w: pageW - margin * 2,
    h: pageH - topGap - bottomGap,
  };

  const contentX = frame.x + pad;
  const contentY = frame.y + pad;
  const contentW = frame.w - pad * 2;
  const contentH = frame.h - pad * 2;

  // Левая колонка шире — карта крупнее
  const colMapW = contentW * 0.56;
  const colTextW = contentW - colMapW - gap;

  const mapArea = { x: contentX, y: contentY, w: colMapW, h: contentH };
  const textArea = {
    x: contentX + colMapW + gap,
    y: contentY,
    w: colTextW,
    h: contentH,
  };

  const mapCenter = {
    x: mapArea.x + mapArea.w / 2,
    y: mapArea.y + mapArea.h / 2,
  };
  const mapRadius = Math.min(mapArea.w, mapArea.h) / 2 - 8;

  const num = (n: number) => n.toFixed(2);

  const ring = (alt: number) =>
    mapRadius * (1 - Math.max(0, Math.min(90, Math.abs(alt))) / 90);

  const project = (alt: number, az: number) => {
    const r = ring(alt);
    const theta = (az * Math.PI) / 180;
    return {
      x: mapCenter.x + r * Math.sin(theta),
      y: mapCenter.y - r * Math.cos(theta),
    };
  };

  const edge = (az: number, extra: number) => {
    const r = mapRadius + extra;
    const theta = (az * Math.PI) / 180;
    return {
      x: mapCenter.x + r * Math.sin(theta),
      y: mapCenter.y - r * Math.cos(theta),
    };
  };

  const gridCircles = ALT_RINGS.map(
    (alt) =>
      `<circle cx="${num(mapCenter.x)}" cy="${num(mapCenter.y)}" r="${num(ring(alt))}" />`,
  ).join("");

  const gridLines = AZ_TICKS.map((az) => {
    const p = edge(az, 0);
    return `<line x1="${num(mapCenter.x)}" y1="${num(mapCenter.y)}" x2="${num(p.x)}" y2="${num(p.y)}" />`;
  }).join("");

  const axisLines = [
    [edge(0, 0), edge(180, 0)],
    [edge(90, 0), edge(270, 0)],
  ]
    .map(
      ([a, b]) =>
        `<line x1="${num(a.x)}" y1="${num(a.y)}" x2="${num(b.x)}" y2="${num(b.y)}" class="axis-line" />`,
    )
    .join("");

  const horizonTicks = [0, 90, 180, 270]
    .map((az) => {
      const a = edge(az, -2);
      const b = edge(az, 2);
      return `<line x1="${num(a.x)}" y1="${num(a.y)}" x2="${num(b.x)}" y2="${num(b.y)}" class="horizon-tick" />`;
    })
    .join("");

  const cardinalLabels = AZ_TICKS.filter((az) => CARDINALS[az])
    .map((az) => {
      const p = edge(az, 6.5);
      return `<text x="${num(p.x)}" y="${num(p.y)}" class="cardinal-label">${CARDINALS[az]}</text>`;
    })
    .join("");

  const printPoints = points.value.map((p) => ({
    ...p,
    ...project(p.alt, p.az),
  }));
  const printPointsByOid = new Map(printPoints.map((p) => [p.oid, p]));

  const routeMarkup = manualEdges.value
    .flatMap((e) => {
      const from = printPointsByOid.get(e.from);
      const to = printPointsByOid.get(e.to);
      if (!from || !to) return [];
      return [
        `<line x1="${num(from.x)}" y1="${num(from.y)}" x2="${num(to.x)}" y2="${num(to.y)}" class="route-line" />`,
      ];
    })
    .join("");

  const starMarkup = printPoints
    .map((p) => {
      const glyph =
        p.alt >= 0
          ? `<circle cx="${num(p.x)}" cy="${num(p.y)}" r="1.1" class="star-glyph" />`
          : `<circle cx="${num(p.x)}" cy="${num(p.y)}" r="1.1" class="star-glyph star-glyph--below" />`;
      return (
        glyph +
        `<text x="${num(p.x)}" y="${num(p.y - 3)}" class="star-label">${escapeXml(p.name)}</text>`
      );
    })
    .join("");

  const textLines: string[] = [];
  let y = textArea.y + 4;

  const addHeading = (text: string) => {
    textLines.push(
      `<text x="${num(textArea.x)}" y="${num(y)}" class="sb-heading">${escapeXml(text)}</text>`,
    );
    y += 5.4;
  };
  const addBody = (text: string, cls = "sb-body") => {
    textLines.push(
      `<text x="${num(textArea.x)}" y="${num(y)}" class="${cls}">${escapeXml(text)}</text>`,
    );
    y += 4.6;
  };

  addHeading("ANGULAR SIZE — HAND REFERENCE");
  ANGULAR_TIPS.forEach((tip) => addBody(`\u2022 ${tip}`));
  addBody("Extend your arm as fully as possible.", "sb-note");

  y += 4;
  addHeading("MOON");
  if (moon.value) {
    addBody(
      `Alt ${moon.value.alt.toFixed(1)}°, Az ${moon.value.az.toFixed(1)}°`,
    );
    addBody(
      moon.value.alt > 0
        ? "Above the horizon"
        : "Below the horizon (not visible)",
      "sb-note",
    );
  } else {
    addBody("No moon data for this mission.", "sb-note");
  }

  y += 4;
  addHeading("LEGEND");
  addBody("\u25CF observable now      \u25CB below horizon");

    // Надпись по центру верхней полосы между краем листа и рамкой
  const logoTop = logoMarkup(frame.x, frame.y - topGap / 2, "start");

  // Надпись по центру нижней полосы между рамкой и краем листа
  const logoBottom = logoMarkup(
    frame.x + frame.w,
    frame.y + frame.h + bottomGap / 2,
    "end",
  );

  return `
<svg viewBox="0 0 ${pageW} ${pageH}" xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="${pageW}" height="${pageH}" fill="#fff" />
  <rect x="${num(frame.x)}" y="${num(frame.y)}" width="${num(frame.w)}" height="${num(frame.h)}" class="page-frame" />
  <line x1="${num(contentX + colMapW + gap / 2)}" y1="${num(contentY)}" x2="${num(contentX + colMapW + gap / 2)}" y2="${num(contentY + contentH)}" class="divider" />
  <g class="grid-dashed">${gridCircles}${gridLines}</g>
  ${axisLines}
  <circle cx="${num(mapCenter.x)}" cy="${num(mapCenter.y)}" r="${num(mapRadius)}" class="horizon-circle" />
  <circle cx="${num(mapCenter.x)}" cy="${num(mapCenter.y)}" r="0.6" fill="#000" />
  ${horizonTicks}
  ${cardinalLabels}
  ${routeMarkup}
  ${starMarkup}
  ${textLines.join("")}
  ${logoTop}
  ${logoBottom}
</svg>`.trim();
}

function collectParentStyles(): string {
  const links = Array.from(
    document.querySelectorAll<HTMLLinkElement>('link[rel="stylesheet"]'),
  )
    .map((el) => `<link rel="stylesheet" href="${el.href}">`) // el.href — уже абсолютный
    .join("");
  const styles = Array.from(document.querySelectorAll("style"))
    .map((el) => el.outerHTML)
    .join("");
  return links + styles;
}

function triggerPrint(): void {
  if (!canPrint.value) return;

  const printWindow = window.open("", "_blank", "width=1200,height=850");
  if (!printWindow) {
    window.print();
    return;
  }

  printWindow.document.open();
  printWindow.document.write(
    `<!doctype html><html><head>` +
      `<meta charset="utf-8" />` +
      // <base> заставляет относительные пути в скопированных <link> резолвиться
      // относительно исходной страницы, а не about:blank — иначе шрифты
      // молча падают в fallback и метрики уезжают.
      `<base href="${location.href}" />` +
      `<title>${escapeXml(pageTitle.value)} — star map</title>` +
      collectParentStyles() +
      `<style>${PRINT_STYLES}</style></head><body>${buildPrintSvg()}</body></html>`,
  );
  printWindow.document.close();

  printWindow.onload = () => {
    const fire = () => {
      printWindow.focus();
      printWindow.print();
    };
    if (printWindow.document.fonts) {
      printWindow.document.fonts.ready.then(fire).catch(fire);
    } else {
      fire();
    }
  };
}
</script>

<template>
  <div class="w-full">
    <div
      v-if="!mission.data"
      class="flex aspect-square w-full items-center justify-center rounded-lg border border-dashed border-border bg-card text-sm text-muted-foreground"
    >
      No data yet
    </div>

    <TooltipProvider v-else :delay-duration="150">
      <div class="relative overflow-hidden">
        <svg
          class="block aspect-square w-full"
          :viewBox="`0 0 ${SIZE} ${SIZE}`"
          xmlns="http://www.w3.org/2000/svg"
          role="img"
          aria-label="Star-hopping route preview"
        >
          <defs>
            <radialGradient id="starGlow">
              <stop offset="0%" stop-color="#ffffff" stop-opacity="1" />
              <stop offset="40%" stop-color="#d0e8ff" stop-opacity="0.8" />
              <stop offset="100%" stop-color="#a0c8f0" stop-opacity="0" />
            </radialGradient>
          </defs>

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
              v-for="az in AZ_TICKS"
              :key="`az-${az}`"
              :x1="center.x"
              :y1="center.y"
              :x2="edgeXY(az, 0).x"
              :y2="edgeXY(az, 0).y"
            />
          </g>

          <circle
            :cx="center.x"
            :cy="center.y"
            :r="radius"
            class="horizon-circle"
          />
          <circle :cx="center.x" :cy="center.y" r="0.7" fill="black" />

          <!-- Засечки на горизонте для сторон света -->
          <g v-for="az in [0, 90, 180, 270]" :key="`tick-${az}`">
            <line
              :x1="edgeXY(az, -2).x"
              :y1="edgeXY(az, -2).y"
              :x2="edgeXY(az, 2).x"
              :y2="edgeXY(az, 2).y"
              class="horizon-tick"
            />
          </g>

          <text
            v-for="az in [0, 45, 90, 135, 180, 270]"
            :key="`card-${az}`"
            :x="edgeXY(az, 9).x"
            :y="edgeXY(az, 9).y"
            class="cardinal-label"
          >
            {{ CARDINALS[az] }}
          </text>

          <!-- moon -->
          <g v-if="moon && moon.alt > 0">
            <circle
              :cx="moon.x"
              :cy="moon.y"
              r="2.6"
              fill="none"
              stroke="#000"
              stroke-width="0.4"
            />
            <circle :cx="moon.x + 1" :cy="moon.y - 0.4" r="2.3" fill="white" />
          </g>

          <!-- star-hop route -->
          <g v-for="seg in segments" :key="`seg-${seg.key}`">
            <line
              :x1="seg.from.x"
              :y1="seg.from.y"
              :x2="seg.to.x"
              :y2="seg.to.y"
              class="route-line"
              :class="{ 'is-active': isTouching(seg) }"
            />
            <!-- wider transparent hit area so the segment is easy to click -->
            <line
              :x1="seg.from.x"
              :y1="seg.from.y"
              :x2="seg.to.x"
              :y2="seg.to.y"
              class="route-hit"
              :class="{ 'is-cuttable': activeTool === 'scissors' }"
              @click="onSegmentClick(seg)"
            />
          </g>

          <!-- ruler measurement -->
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
              :y="(measurement.from.y + measurement.to.y) / 2 - 2"
              class="measure-label"
            >
              {{ measurement.distance.toFixed(1) }}°
            </text>
          </g>

          <!-- stars -->
          <g v-for="pt in points" :key="`star-${pt.oid}`">
            <g v-if="pt.alt >= 0">
              <circle
                :cx="pt.x"
                :cy="pt.y"
                r="2.5"
                class="star-halo"
                :class="{ 'is-hovered': hoveredOid === pt.oid }"
              />
            </g>

            <!-- Основной круг звезды (или контур, если ниже горизонта) -->
            <circle
              :cx="pt.x"
              :cy="pt.y"
              r="1.2"
              class="star-core"
              :class="[
                hoveredOid === pt.oid && 'is-hovered',
                pt.alt < 0 && 'is-invisible',
                pendingStars.includes(pt.oid) && 'is-pending',
              ]"
            />

            <!-- Label backplate -->
            <rect
              :x="pt.x - getTextWidth(pt.name) / 2 - 1"
              :y="pt.y - 6.2"
              :width="getTextWidth(pt.name) + 2"
              height="4.2"
              rx="0.6"
              class="label-backplate"
              :class="{ 'is-hovered': hoveredOid === pt.oid }"
            />

            <text
              :x="pt.x"
              :y="pt.y - 3.2"
              class="star-label"
              :class="{ 'is-hovered': hoveredOid === pt.oid }"
            >
              {{ pt.name }}
            </text>
          </g>
        </svg>

        <div class="pointer-events-none absolute inset-0">
          <Tooltip v-if="moon">
            <TooltipTrigger as-child>
              <button
                type="button"
                class="pointer-events-auto absolute size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-none bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :style="{
                  left: `${(moon.x / SIZE) * 100}%`,
                  top: `${(moon.y / SIZE) * 100}%`,
                }"
                aria-label="Moon"
              />
            </TooltipTrigger>
            <TooltipContent>
              Moon — Alt {{ moon.alt.toFixed(1) }}°, Az
              {{ moon.az.toFixed(1) }}°
            </TooltipContent>
          </Tooltip>

          <Tooltip v-for="pt in points" :key="`tip-${pt.oid}`">
            <TooltipTrigger as-child>
              <button
                type="button"
                class="pointer-events-auto absolute size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-none bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :class="{
                  'cursor-crosshair':
                    activeTool === 'ruler' || activeTool === 'pencil',
                }"
                :style="{
                  left: `${(pt.x / SIZE) * 100}%`,
                  top: `${(pt.y / SIZE) * 100}%`,
                }"
                :aria-label="pt.name"
                @pointerenter="hoveredOid = pt.oid"
                @pointerleave="hoveredOid = null"
                @focus="hoveredOid = pt.oid"
                @blur="hoveredOid = null"
                @click="onStarClick(pt.oid)"
              />
            </TooltipTrigger>
            <TooltipContent>
              {{ pt.name }} — Alt {{ pt.alt.toFixed(1) }}°, Az
              {{ pt.az.toFixed(1) }}°
            </TooltipContent>
          </Tooltip>
        </div>

        <!-- Print -->
        <div class="pointer-events-auto absolute right-2 top-2 z-10">
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 rounded-full border border-white/10 bg-[#0d1420]/90 text-muted-foreground shadow-md backdrop-blur hover:text-foreground"
                :disabled="!canPrint"
                aria-label="Print star map"
                @click="triggerPrint"
              >
                <Printer class="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {{ canPrint ? "Print star map" : "Route isn't ready yet" }}
            </TooltipContent>
          </Tooltip>
        </div>

        <!-- Figma-style floating toolbar -->
        <div
          class="pointer-events-auto flex justify-center items-center gap-0.5 rounded-full border border-white/10 bg-[#0d1420]/95 p-1 shadow-lg backdrop-blur w-fit mx-auto"
        >
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 rounded-full"
                :class="{
                  'bg-white/10 text-[#b8472d]': activeTool === 'ruler',
                }"
                aria-label="Measure distance"
                @click="selectTool('ruler')"
              >
                <Ruler class="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Measure distance between two stars</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 rounded-full"
                :class="{
                  'bg-white/10 text-[#b8472d]': activeTool === 'scissors',
                }"
                aria-label="Cut route segment"
                @click="selectTool('scissors')"
              >
                <Scissors class="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Click a route segment to delete it</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 rounded-full"
                :class="{
                  'bg-white/10 text-[#b8472d]': activeTool === 'pencil',
                }"
                aria-label="Draw new connection"
                @click="selectTool('pencil')"
              >
                <Pencil class="size-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent
              >Connect two stars with a new segment</TooltipContent
            >
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  </div>
</template>

<style scoped>
.grid-dashed circle,
.grid-dashed line {
  stroke: #54617c;
  stroke-width: 0.3;
  stroke-dasharray: 1, 1.2;
}
.horizon-circle {
  fill: none;
  stroke: #f2f2fa;
  stroke-width: 0.5;
}
.horizon-tick {
  stroke: #f2f2fa;
  stroke-width: 0.5;
}
.cardinal-label {
  font:
    700 8px "MuseoModerno",
    monospace;
  fill: #f2f2fa;
  text-anchor: middle;
  dominant-baseline: middle;
}
.route-line {
  stroke: #f2f2fa;
  stroke-width: 0.5;
  stroke-dasharray: 1.6, 1.2;
  pointer-events: none;
}
.route-line.is-active {
  stroke-width: 1;
  stroke: #b8472d;
}
.route-hit {
  stroke: transparent;
  stroke-width: 4;
  fill: none;
}
.route-hit.is-cuttable {
  cursor: pointer;
}
.measure-line {
  stroke: #ffd966;
  stroke-width: 0.6;
  stroke-dasharray: 0.8, 0.8;
}
.measure-label {
  font:
    700 4px "LINE Seed JP",
    "Inter",
    monospace;
  fill: #ffd966;
  text-anchor: middle;
}

/* Звезда класса B */
.star-core {
  fill: #ffffff;
  stroke: none;
  transition:
    fill 0.1s ease,
    filter 0.1s ease;
}
.star-core.is-hovered {
  fill: #ffd966; /* тёплый золотистый при наведении */
}
.star-core.is-pending {
  fill: #b8472d;
}
.star-core.is-invisible {
  fill: transparent;
  stroke: #a0c8f0;
  stroke-width: 1;
}
.star-core.is-invisible.is-hovered {
  stroke: #ffd966;
}

.star-halo {
  fill: url(#starGlow);
  opacity: 0.6;
  transition: opacity 0.1s ease;
}
.star-halo.is-hovered {
  opacity: 1;
}

.label-backplate {
  fill: #020914;
}
.star-label {
  font:
    400 4px "LINE Seed JP",
    "Inter",
    monospace;
  fill: #f2f2fa;
  text-anchor: middle;
}
.star-label.is-hovered {
  fill: #b8472d;
  font-weight: 700;
}
</style>
