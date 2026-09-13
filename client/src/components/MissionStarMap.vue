<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { Mission } from "@/types/api";
import {
  Contrast,
  Flashlight,
  Pencil,
  Printer,
  Ruler,
  Scissors,
} from "@lucide/vue";
import Button from "./ui/button/Button.vue";
import * as Skymap from "@/lib/skymap/";
import StarMapLayer from "./StarMapLayer.vue";

const props = defineProps<{
  mission: Mission;
}>();

const objectives = computed(() => props.mission.data?.objectives ?? []);
const mapData = computed(() => props.mission.map_data ?? null);

const points = computed<Skymap.Point[]>(() => {
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
        ...Skymap.toXY(pos),
      },
    ];
  });
});

const pointsByOid = computed(
  () => new Map(points.value.map((p) => [p.oid, p])),
);

function edgesFromTour(tour: number[]): Skymap.Edge[] {
  const list: Skymap.Edge[] = [];
  for (let i = 0; i < tour.length - 1; i++) {
    list.push({ from: tour[i], to: tour[i + 1] });
  }
  if (tour.length > 1) {
    list.push({ from: tour[tour.length - 1], to: tour[0] });
  }
  return list;
}

const manualEdges = ref<Skymap.Edge[]>([]);

watch(
  () => mapData.value?.tour,
  (tour) => {
    manualEdges.value = tour ? edgesFromTour(tour) : [];
  },
  { immediate: true },
);

const segments = computed<Skymap.Segment[]>(() => {
  return manualEdges.value.flatMap((edge) => {
    const from = pointsByOid.value.get(edge.from);
    const to = pointsByOid.value.get(edge.to);
    if (!from || !to) return [];
    return [
      {
        key: Skymap.edgeKey(edge),
        from,
        to,
        distance: Skymap.angularSeparation(from, to),
      },
    ];
  });
});

const moon = computed<Skymap.Point | null>(() => {
  const pos = mapData.value?.moon_position;
  if (!pos) return null;
  return {
    oid: -1,
    name: "moon",
    alt: pos.alt,
    az: pos.az,
    ...Skymap.toXY(pos),
  };
});

const hoveredOid = ref<number | null>(null);

function isTouching(seg: (typeof segments.value)[number]): boolean {
  return (
    hoveredOid.value !== null &&
    (seg.from.oid === hoveredOid.value || seg.to.oid === hoveredOid.value)
  );
}

type ToolId = "ruler" | "scissors" | "pencil";

const activeTool = ref<ToolId | null>(null);

const MAP_STYLES = [
  { id: "red", label: "Red", css: Skymap.STYLE_RED, icon: Flashlight },
  {
    id: "grayscale",
    label: "Grayscale",
    css: Skymap.STYLE_GRAYSCALE,
    icon: Contrast,
  },
] as const;

type MapStyleId = (typeof MAP_STYLES)[number]["id"];

const activeStyleId = ref<MapStyleId>("red");
const activeStyle = computed(
  () => MAP_STYLES.find((s) => s.id === activeStyleId.value) ?? MAP_STYLES[0],
);
const inactiveStyles = computed(() =>
  MAP_STYLES.filter((s) => s.id !== activeStyleId.value),
);
const mapStyles = computed(
  () => Skymap.STYLE_BASE + activeStyle.value.css + Skymap.STYLE_INTERACTIVE,
);

const styleMenuOpen = ref(false);
const styleMenuRef = ref<HTMLElement | null>(null);

function onDocPointerDown(e: PointerEvent) {
  if (!styleMenuOpen.value) return;
  if (styleMenuRef.value && !styleMenuRef.value.contains(e.target as Node)) {
    styleMenuOpen.value = false;
  }
}
function onDocKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") styleMenuOpen.value = false;
}

const pendingStars = ref<number[]>([]);
const measurement = ref<Skymap.Segment | null>(null);

function resetToolState() {
  pendingStars.value = [];
  measurement.value = null;
}

function selectTool(tool: ToolId) {
  activeTool.value = activeTool.value === tool ? null : tool;
  resetToolState();
}

function onStarClick(oid: number) {
  switch (activeTool.value) {
    case "ruler":
      pendingStars.value.push(oid);
      if (pendingStars.value.length === 2) {
        const from = pointsByOid.value.get(pendingStars.value[0]);
        const to = pointsByOid.value.get(pendingStars.value[1]);
        if (from && to) {
          measurement.value = {
            key: "measurement",
            from,
            to,
            distance: Skymap.angularSeparation(from, to),
          };
        }
        pendingStars.value = [];
      }
      break;
    case "pencil":
      if (pendingStars.value[0] === oid) return;
      pendingStars.value.push(oid);
      if (pendingStars.value.length === 2) {
        const [from, to] = pendingStars.value;
        const exists = manualEdges.value.some(
          (e) => Skymap.edgeKey(e) === Skymap.edgeKey({ from, to }),
        );
        if (!exists) {
          manualEdges.value.push({ from, to });
        }
        pendingStars.value = [];
      }
      break;
  }
}

function onSegmentClick(seg: (typeof segments.value)[number]) {
  switch (activeTool.value) {
    case "scissors":
      manualEdges.value = manualEdges.value.filter(
        (e) => Skymap.edgeKey(e) !== seg.key,
      );
      break;
  }
}

const canPrint = computed(() => !!mapData.value);

const pageTitle = computed(
  () =>
    `Mission ${Skymap.dateFormatter.format(new Date(props.mission.created_at))}`,
);

const mapEl = ref<HTMLElement | null>(null);
const toolbarLeft = ref(0);

let rafId = 0;
let resizeObserver: ResizeObserver | null = null;

function updateToolbarPosition() {
  rafId = 0;
  const el = mapEl.value;
  if (!el) return;
  const rect = el.getBoundingClientRect();
  toolbarLeft.value = rect.left + rect.width / 2;
}

function scheduleToolbarUpdate() {
  if (rafId) return;
  rafId = requestAnimationFrame(updateToolbarPosition);
}

onMounted(() => {
  updateToolbarPosition();
  window.addEventListener("resize", scheduleToolbarUpdate);
  window.addEventListener("scroll", scheduleToolbarUpdate, true);
  if (mapEl.value) {
    resizeObserver = new ResizeObserver(scheduleToolbarUpdate);
    resizeObserver.observe(mapEl.value);
  }
  document.addEventListener("pointerdown", onDocPointerDown);
  document.addEventListener("keydown", onDocKeydown);
});

onBeforeUnmount(() => {
  if (rafId) cancelAnimationFrame(rafId);
  window.removeEventListener("resize", scheduleToolbarUpdate);
  window.removeEventListener("scroll", scheduleToolbarUpdate, true);
  resizeObserver?.disconnect();
  resizeObserver = null;
  document.removeEventListener("pointerdown", onDocPointerDown);
  document.removeEventListener("keydown", onDocKeydown);
});

function collectParentStyles(): string {
  const links = Array.from(
    document.querySelectorAll<HTMLLinkElement>('link[rel="stylesheet"]'),
  )
    .map((el) => `<link rel="stylesheet" href="${el.href}">`)
    .join("");
  const styles = Array.from(document.querySelectorAll("style"))
    .map((el) => el.outerHTML)
    .join("");
  return links + styles;
}

async function triggerPrint() {
  if (!canPrint.value) return;

  const iframe = document.createElement("iframe");
  iframe.setAttribute("aria-hidden", "true");
  iframe.style.cssText =
    "position:fixed;right:0;bottom:0;width:0;height:0;border:0;visibility:hidden;";
  document.body.appendChild(iframe);

  const win = iframe.contentWindow;
  const doc = iframe.contentDocument;
  if (!win || !doc) {
    iframe.remove();
    window.print();
    return;
  }

  doc.open();
  doc.write(
    `<!doctype html><html><head>` +
      `<meta charset="utf-8" />` +
      `<base href="${location.href}" />` +
      `<title>${Skymap.escapeXml(pageTitle.value)} — star map</title>` +
      collectParentStyles() +
      `<style>${activeStyle.value.css}</style></head>` +
      `<body>${await Skymap.buildPrintSvg(points.value, manualEdges.value, moon.value)}</body></html>`,
  );
  doc.close();

  const fire = () => {
    win.focus();
    win.print();
    setTimeout(() => iframe.remove(), 1000);
  };

  if (doc.fonts) {
    doc.fonts.ready.then(fire).catch(fire);
  } else {
    fire();
  }
}
</script>

<template>
  <div class="w-full overflow-x-hidden flex justify-center">
    <TooltipProvider :delay-duration="150">
      <div ref="mapEl" class="relative overflow-hidden h-dvh aspect-297/210">
        <StarMapLayer
          class="absolute inset-0 w-full h-full"
          interactive
          :styles="mapStyles"
          :points="points"
          :edges="manualEdges"
          :moon="moon"
          :hovered-oid="hoveredOid"
          :selected-stars="pendingStars"
          :active-tool="activeTool"
          :measurement="measurement"
          @segment-click="onSegmentClick"
        />

        <div class="pointer-events-none absolute inset-0">
          <Tooltip v-if="moon">
            <TooltipTrigger as-child>
              <button
                type="button"
                class="pointer-events-auto absolute size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-none bg-transparent p-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :style="Skymap.mapToPercent(moon)"
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
                :style="Skymap.mapToPercent(pt)"
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

        <div
          class="fixed bottom-2 z-50 flex -translate-x-1/2 items-center gap-0.5 rounded-lg border border-white/10 bg-[#0d1420]/95 p-1 shadow-lg backdrop-blur w-fit"
          :style="{ left: `${toolbarLeft}px` }"
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
                <Ruler class="size-5 stroke-1" nonScalingStroke />
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
                <Scissors class="size-5 stroke-1" nonScalingStroke />
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
                <Pencil class="size-5 stroke-1" nonScalingStroke />
              </Button>
            </TooltipTrigger>
            <TooltipContent
              >Connect two stars with a new segment</TooltipContent
            >
          </Tooltip>

          <div class="mx-1 h-5 w-px shrink-0 bg-white/20" aria-hidden="true" />

          <div ref="styleMenuRef" class="relative">
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-8 rounded-full"
                  aria-label="Change color mode"
                  :aria-expanded="styleMenuOpen"
                  @click="styleMenuOpen = !styleMenuOpen"
                >
                  <component
                    :is="activeStyle.icon"
                    class="size-5 stroke-1"
                    nonScalingStroke
                  />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Change color mode</TooltipContent>
            </Tooltip>

            <div
              v-if="styleMenuOpen"
              class="absolute bottom-full left-1/2 z-[60] mb-2 flex -translate-x-1/2 gap-0.5 rounded-lg border border-white/10 bg-[#0d1420]/95 p-1 shadow-lg backdrop-blur"
              role="menu"
            >
              <Tooltip v-for="style in inactiveStyles" :key="style.id">
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="size-8 rounded-full"
                    role="menuitem"
                    :aria-label="style.label"
                    @click="
                      activeStyleId = style.id;
                      styleMenuOpen = false;
                    "
                  >
                    <component
                      :is="style.icon"
                      class="size-5 stroke-1"
                      nonScalingStroke
                    />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{{ style.label }}</TooltipContent>
              </Tooltip>
            </div>
          </div>

          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-8 rounded-full"
                :disabled="!canPrint"
                aria-label="Print star map"
                @click="triggerPrint"
              >
                <Printer class="size-5 stroke-1" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {{ canPrint ? "Print star map" : "Route isn't ready yet" }}
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  </div>
</template>
