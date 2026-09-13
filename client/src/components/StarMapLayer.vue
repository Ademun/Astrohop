<script setup lang="ts">
import * as Skymap from "@/lib/skymap";
import { computed } from "vue";

type ToolId = "ruler" | "scissors" | "pencil";

interface Props {
  points: Skymap.Point[];
  edges: Skymap.Edge[];
  moon?: Skymap.Point | null;
  interactive?: boolean;
  styles?: string;
  hoveredOid?: number | null;
  selectedStars?: number[];
  activeTool?: ToolId | null;
  measurement?: Skymap.Segment | null;
}

const props = withDefaults(defineProps<Props>(), {
  moon: null,
  interactive: false,
  styles: "",
  hoveredOid: null,
  selectedStars: () => [],
  activeTool: null,
  measurement: null,
});

const emit = defineEmits<{
  (e: "segmentClick", seg: Skymap.Segment): void;
}>();

const byOid = computed(() => new Map(props.points.map((p) => [p.oid, p])));

const segments = computed<Skymap.Segment[]>(() =>
  props.edges.flatMap((e) => {
    const from = byOid.value.get(e.from);
    const to = byOid.value.get(e.to);
    if (!from || !to) return [];
    return [
      {
        key: Skymap.edgeKey(e),
        from,
        to,
        distance: Skymap.angularSeparation(from, to),
      },
    ];
  }),
);

function isTouching(seg: Skymap.Segment) {
  return (
    props.hoveredOid !== null &&
    (seg.from.oid === props.hoveredOid || seg.to.oid === props.hoveredOid)
  );
}

const CARDINAL_AZ = Object.keys(Skymap.CARDINALS).map(Number);
const COMPASS_TICKS = [0, 90, 180, 270];

const textArea = {
  x: Skymap.contentX + Skymap.colMapW + Skymap.COL_GAP,
  y: Skymap.contentY,
  w: Skymap.colTextW,
  h: Skymap.contentH,
};

const dividerX = Skymap.contentX + Skymap.colMapW + Skymap.COL_GAP / 2;

const ANGULAR_TIPS = [
  "Finger width ~ 1°",
  "Three middle fingers ~ 5°",
  "Fist width ~ 10°",
  "Index to pinky ~ 15°",
  "Open hand width ~ 20°",
];

interface TextLine {
  x: number;
  y: number;
  cls: string;
  text: string;
}

const textLines = computed<TextLine[]>(() => {
  const lines: TextLine[] = [];
  let y = textArea.y + 4;

  const addHeading = (text: string) => {
    lines.push({ x: textArea.x, y, cls: "sb-heading", text });
    y += 5.4;
  };
  const addBody = (text: string, cls = "sb-body") => {
    lines.push({ x: textArea.x, y, cls, text });
    y += 4.6;
  };

  addHeading("ANGULAR SIZE — HAND REFERENCE");
  ANGULAR_TIPS.forEach((tip) => addBody(`\u2022 ${tip}`));
  addBody("Extend your arm as fully as possible.", "sb-note");

  y += 4;
  addHeading("MOON");
  if (props.moon) {
    addBody(
      `Alt ${props.moon.alt.toFixed(1)}°, Az ${props.moon.az.toFixed(1)}°`,
    );
    addBody(
      props.moon.alt > 0
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

  return lines;
});

const barW = textArea.w;
const barH = 10;
const barX = textArea.x;
const barY = textArea.y + textArea.h - barH;
const BAR_FRAME_STROKE = 0.9;
const BAR_INNER_PAD = 1.1;
const BAR_GAP = 0.35;
const cellX0 = barX + BAR_INNER_PAD;
const cellY0 = barY + BAR_INNER_PAD;
const cellW = (barW - BAR_INNER_PAD * 2 - BAR_GAP * 3) / 4;
const cellH = barH - BAR_INNER_PAD * 2;
const BAR_SHADES = ["#000000", "#3f3f3f", "#808080", "#d9d9d9"];

const CHART_H = 20;
const CHART_GAP = 10;
const CHART_HEADER_H = 8;
const chartX = barX;
const chartW = barW;
const chartY = barY - CHART_HEADER_H - CHART_GAP - CHART_H;

const extinctionZs = Array.from(
  {
    length:
      (Skymap.EXTINCTION_Z_MAX - Skymap.EXTINCTION_Z_MIN) /
        Skymap.EXTINCTION_Z_STEP +
      1,
  },
  (_, i) => Skymap.EXTINCTION_Z_MIN + i * Skymap.EXTINCTION_Z_STEP,
);
const extinctionValues = extinctionZs.map((z) =>
  Skymap.extinctionMagnitude(z, Skymap.EXTINCTION_K),
);
const plotX = chartX + Skymap.EXTINCTION_AXIS_LABEL_W;
const plotW = chartW - Skymap.EXTINCTION_AXIS_LABEL_W;

const extinctionMagMin = Math.min(...extinctionValues);
const extinctionMagMax = Math.max(...extinctionValues);
const logMin = Math.log10(extinctionMagMin);
const logMax = Math.log10(extinctionMagMax);

function magToY(v: number): number {
  return (
    chartY + CHART_H - ((Math.log10(v) - logMin) / (logMax - logMin)) * CHART_H
  );
}

const extinctionPolyline = computed(() =>
  extinctionZs
    .map((z, i) => {
      const px =
        plotX +
        ((z - Skymap.EXTINCTION_Z_MIN) /
          (Skymap.EXTINCTION_Z_MAX - Skymap.EXTINCTION_Z_MIN)) *
          plotW;
      return `${px.toFixed(2)},${magToY(extinctionValues[i]).toFixed(2)}`;
    })
    .join(" "),
);

const extinctionYTicks = Skymap.EXTINCTION_MAG_TICKS.filter(
  (v) => v >= extinctionMagMin && v <= extinctionMagMax,
).map((v) => ({ value: v, y: magToY(v) }));

const ALT_SCALE_Y = Skymap.center.y + Skymap.radius + 22;
const ALT_SCALE_W = Skymap.radius;
const ALT_SCALE_X0 = Skymap.center.x - ALT_SCALE_W;
const ALT_SCALE_STEP = 15;
const ALT_SCALE_TICK_HALF = 1.5;
const ALT_SCALE_LABEL_OFFSET = 3.5;

const altScaleTicks = Array.from(
  { length: 90 / ALT_SCALE_STEP + 1 },
  (_, i) => i * ALT_SCALE_STEP,
).map((alt) => {
  const x = ALT_SCALE_X0 + (alt / 90) * ALT_SCALE_W;
  return {
    alt,
    x1: x,
    y1: ALT_SCALE_Y - ALT_SCALE_TICK_HALF,
    x2: x,
    y2: ALT_SCALE_Y + ALT_SCALE_TICK_HALF,
    labelX: x,
    labelY: ALT_SCALE_Y + ALT_SCALE_LABEL_OFFSET,
  };
});
</script>

<template>
  <svg
    class="theme block"
    :viewBox="`0 0 ${Skymap.PAGE_W} ${Skymap.PAGE_H}`"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Star-hopping route map"
  >
    <component v-if="styles" :is="'style'" v-html="styles" />

    <defs>
      <radialGradient id="starGlow">
        <stop offset="0%" stop-color="#ffffff" stop-opacity="1" />
        <stop offset="40%" stop-color="#d0e8ff" stop-opacity="0.8" />
        <stop offset="100%" stop-color="#a0c8f0" stop-opacity="0" />
      </radialGradient>
    </defs>

    <rect
      x="0"
      y="0"
      :width="Skymap.PAGE_W"
      :height="Skymap.PAGE_H"
      class="fill-white"
    />
    <rect
      :x="Skymap.frame.x"
      :y="Skymap.frame.y"
      :width="Skymap.frame.w"
      :height="Skymap.frame.h"
      class="page-frame"
    />
    <line
      :x1="dividerX"
      :y1="Skymap.contentY"
      :x2="dividerX"
      :y2="Skymap.contentY + Skymap.contentH"
      class="divider"
    />

    <g
      :transform="Skymap.MAP_TRANSFORM"
      :style="{ '--map-scale': String(Skymap.MAP_SCALE) }"
    >
      <g class="grid-dashed">
        <circle
          v-for="alt in Skymap.ALT_RINGS"
          :key="`ring-${alt}`"
          :cx="Skymap.center.x"
          :cy="Skymap.center.y"
          :r="Skymap.ringRadius(alt)"
          fill="none"
        />
        <line
          v-for="az in Skymap.AZ_TICKS"
          :key="`az-${az}`"
          :x1="Skymap.center.x"
          :y1="Skymap.center.y"
          :x2="Skymap.edgeXY(az, 0).x"
          :y2="Skymap.edgeXY(az, 0).y"
        />
      </g>

      <line
        :x1="Skymap.edgeXY(0, 0).x"
        :y1="Skymap.edgeXY(0, 0).y"
        :x2="Skymap.edgeXY(180, 0).x"
        :y2="Skymap.edgeXY(180, 0).y"
        class="axis-line"
      />
      <line
        :x1="Skymap.edgeXY(90, 0).x"
        :y1="Skymap.edgeXY(90, 0).y"
        :x2="Skymap.edgeXY(270, 0).x"
        :y2="Skymap.edgeXY(270, 0).y"
        class="axis-line"
      />

      <circle
        :cx="Skymap.center.x"
        :cy="Skymap.center.y"
        :r="Skymap.radius"
        class="horizon-circle"
      />
      <circle
        :cx="Skymap.center.x"
        :cy="Skymap.center.y"
        r="0.6"
        class="center-dot"
      />

      <line
        v-for="az in COMPASS_TICKS"
        :key="`tick-${az}`"
        :x1="Skymap.edgeXY(az, -2).x"
        :y1="Skymap.edgeXY(az, -2).y"
        :x2="Skymap.edgeXY(az, 2).x"
        :y2="Skymap.edgeXY(az, 2).y"
        class="horizon-tick"
      />

      <text
        v-for="az in CARDINAL_AZ"
        :key="`card-${az}`"
        :x="Skymap.edgeXY(az, 6.5).x"
        :y="Skymap.edgeXY(az, 6.5).y"
        class="cardinal-label"
      >
        {{ Skymap.CARDINALS[az] }}
      </text>

      <!-- Линейная шкала высоты 0–90° -->
      <g class="alt-scale" aria-label="Altitude scale from 0 to 90 degrees">
        <line
          :x1="ALT_SCALE_X0"
          :y1="ALT_SCALE_Y"
          :x2="ALT_SCALE_X0 + ALT_SCALE_W"
          :y2="ALT_SCALE_Y"
          class="alt-scale-line"
        />
        <g v-for="tick in altScaleTicks" :key="`alt-${tick.alt}`">
          <line
            :x1="tick.x1"
            :y1="tick.y1"
            :x2="tick.x2"
            :y2="tick.y2"
            class="alt-scale-tick"
          />
          <text :x="tick.labelX" :y="tick.labelY" class="alt-scale-label">
            {{ tick.alt }}°
          </text>
        </g>
      </g>

      <g v-if="moon && moon.alt > 0">
        <circle :cx="moon.x" :cy="moon.y" r="2.6" class="moon-outline" />
        <circle :cx="moon.x + 1" :cy="moon.y - 0.4" r="2.3" class="moon-fill" />
      </g>

      <g v-for="seg in segments" :key="`seg-${seg.key}`">
        <line
          :x1="seg.from.x"
          :y1="seg.from.y"
          :x2="seg.to.x"
          :y2="seg.to.y"
          class="route-line"
          :class="{ 'is-active': interactive && isTouching(seg) }"
        />

        <line
          v-if="interactive"
          :x1="seg.from.x"
          :y1="seg.from.y"
          :x2="seg.to.x"
          :y2="seg.to.y"
          class="route-hit"
          :class="{ 'is-cuttable': activeTool === 'scissors' }"
          @click="emit('segmentClick', seg)"
        />
      </g>

      <g v-if="interactive && measurement">
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

      <g v-for="pt in points" :key="`star-${pt.oid}`">
        <circle
          v-if="interactive && pt.alt >= 0"
          :cx="pt.x"
          :cy="pt.y"
          r="2.5"
          class="star-halo"
          :class="{ 'is-hovered': hoveredOid === pt.oid }"
        />

        <circle
          :cx="pt.x"
          :cy="pt.y"
          r="1.2"
          class="star-core"
          :class="[
            pt.alt < 0 && 'is-invisible',
            interactive && hoveredOid === pt.oid && 'is-hovered',
            interactive && selectedStars.includes(pt.oid) && 'is-pending',
          ]"
        />

        <rect
          v-if="interactive"
          :x="pt.x - Skymap.getTextWidth(pt.name) / 2 - 1"
          :y="pt.y - 6.2"
          :width="Skymap.getTextWidth(pt.name) + 2"
          height="4.2"
          rx="0.6"
          class="label-backplate"
          :class="{ 'is-hovered': hoveredOid === pt.oid }"
        />

        <text
          :x="pt.x"
          :y="pt.y - 3.2"
          class="star-label"
          :class="{ 'is-hovered': interactive && hoveredOid === pt.oid }"
        >
          {{ pt.name }}
        </text>
      </g>
    </g>

    <text
      v-for="(line, i) in textLines"
      :key="`tl-${i}`"
      :x="line.x"
      :y="line.y"
      :class="line.cls"
    >
      {{ line.text }}
    </text>

    <g
      class="extinction-chart"
      aria-label="Magnitude loss from atmospheric extinction, 60 to 90 degree zenith angle"
    >
      <text :x="chartX" :y="chartY - 1.6" class="sb-note">
        MAG LOSS NEAR HORIZON 60° to 90°
      </text>

      <rect
        :x="plotX"
        :y="chartY"
        :width="plotW"
        :height="CHART_H"
        class="extinction-chart-frame"
      />
      <polyline :points="extinctionPolyline" class="extinction-chart-line" />

      <g v-for="tick in extinctionYTicks" :key="`ext-tick-${tick.value}`">
        <line
          :x1="plotX - 1"
          :y1="tick.y"
          :x2="plotX"
          :y2="tick.y"
          class="extinction-chart-tick"
        />
        <text :x="plotX - 1.5" :y="tick.y" class="extinction-axis-label">
          {{ tick.value }}
        </text>
      </g>

      <text :x="plotX" :y="chartY + CHART_H + 5" class="sb-note">60°</text>
      <text
        :x="plotX + plotW"
        :y="chartY + CHART_H + 3"
        text-anchor="end"
        class="sb-note"
      >
        90°
      </text>
    </g>

    <text :x="barX" :y="barY - 6.2" class="sb-heading">
      FLASHLIGHT CALIBRATION BAR
    </text>
    <text :x="barX" :y="barY - 1.6" class="sb-note">
      The rightmost section should be barely seen.
    </text>
    <rect
      :x="barX"
      :y="barY"
      :width="barW"
      :height="barH"
      fill="none"
      stroke="#000"
      :stroke-width="BAR_FRAME_STROKE"
    />
    <rect
      v-for="(fill, i) in BAR_SHADES"
      :key="`bar-${i}`"
      :x="cellX0 + i * (cellW + BAR_GAP)"
      :y="cellY0"
      :width="cellW"
      :height="cellH"
      :fill="fill"
    />

    <text
      :x="Skymap.frame.x"
      :y="Skymap.frame.y - Skymap.TOP_GAP / 2"
      text-anchor="start"
      dominant-baseline="central"
      class="logo-text"
    >
      ASTROHOP
    </text>
    <text
      :x="Skymap.frame.x + Skymap.frame.w"
      :y="Skymap.frame.y + Skymap.frame.h + Skymap.BOTTOM_GAP / 2"
      text-anchor="end"
      dominant-baseline="central"
      class="logo-text"
    >
      ASTROHOP
    </text>
  </svg>
</template>
