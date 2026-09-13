import StarMapLayer from "@/components/StarMapLayer.vue";
import * as Types from "./types";
import { h } from "vue";
import { renderToString } from "vue/server-renderer";
import { STYLE_BASE, STYLE_RED } from "./styles";

const PAGE_W = 297;
const PAGE_H = 210;
const PAGE_MARGIN = 12;
const TOP_GAP = 18;
const BOTTOM_GAP = 18;
const PAGE_PAD = 6;
const COL_GAP = 8;

const SIZE = 300;
const MARGIN = 0;
const center = { x: SIZE / 2, y: SIZE / 2 };
const radius = SIZE / 2 - MARGIN - 14;

const frame = {
  x: PAGE_MARGIN,
  y: TOP_GAP,
  w: PAGE_W - PAGE_MARGIN * 2,
  h: PAGE_H - TOP_GAP - BOTTOM_GAP,
};

const contentX = frame.x + PAGE_PAD;
const contentY = frame.y + PAGE_PAD;
const contentW = frame.w - PAGE_PAD * 2;
const contentH = frame.h - PAGE_PAD * 2;

const colMapW = contentW * 0.56;
const colTextW = contentW - colMapW - COL_GAP;
const mapArea = { x: contentX, y: contentY, w: colMapW, h: contentH };

const MAP_SCALE = Math.min(mapArea.w, mapArea.h) / SIZE;
const MAP_TX = mapArea.x + (mapArea.w - SIZE * MAP_SCALE) / 2;
const MAP_TY = mapArea.y + (mapArea.h - SIZE * MAP_SCALE) / 2;
const MAP_TRANSFORM = `translate(${MAP_TX} ${MAP_TY}) scale(${MAP_SCALE})`;

const ALT_RINGS = [15, 30, 45, 60, 75];
const AZ_TICKS = Array.from(
  { length: Math.floor(360 / 30) + 1 },
  (_, i) => i * 30,
);
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

const LOGO_WIDTH = 46;

const EXTINCTION_K = 0.28;
const EXTINCTION_Z_MIN = 60;
const EXTINCTION_Z_MAX = 90;
const EXTINCTION_Z_STEP = 2;
const EXTINCTION_MAG_TICKS = [1, 2, 5, 10];
const EXTINCTION_AXIS_LABEL_W = 6;

function escapeXml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

async function buildPrintSvg(
  points: Types.Point[],
  edges: Types.Edge[],
  moon: Types.Point | null,
): Promise<string> {
  return await renderToString(
    h(StarMapLayer, {
      points,
      edges,
      moon,
      interactive: false,
    }),
  );
}

export {
  PAGE_W,
  PAGE_H,
  PAGE_MARGIN,
  TOP_GAP,
  BOTTOM_GAP,
  COL_GAP,
  PAGE_PAD,
  SIZE,
  MARGIN,
  center,
  radius,
  ALT_RINGS,
  AZ_TICKS,
  CARDINALS,
  ANGULAR_TIPS,
  LOGO_WIDTH,
  MAP_SCALE,
  MAP_TX,
  MAP_TY,
  MAP_TRANSFORM,
  EXTINCTION_K,
  EXTINCTION_Z_MAX,
  EXTINCTION_Z_MIN,
  EXTINCTION_Z_STEP,
  EXTINCTION_MAG_TICKS,
  EXTINCTION_AXIS_LABEL_W,
  frame,
  contentX,
  contentY,
  colMapW,
  contentH,
  mapArea,
  colTextW,
};
export { escapeXml, buildPrintSvg };
