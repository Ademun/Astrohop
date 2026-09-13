import { Horizontal } from "@/types/api";
import * as Markup from "./markup";
import * as Types from "./types";

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});

function ringRadius(alt: number): number {
  if (alt < 0) {
    alt = 0 - alt;
  }
  return Markup.radius * (1 - Math.max(0, Math.min(90, alt)) / 90);
}

function toXY(h: Horizontal) {
  const r = ringRadius(h.alt);
  const theta = (h.az * Math.PI) / 180;
  return {
    x: Markup.center.x + r * Math.sin(theta),
    y: Markup.center.y - r * Math.cos(theta),
  };
}

function edgeXY(az: number, extra: number) {
  const r = Markup.radius + extra;
  const theta = (az * Math.PI) / 180;
  return {
    x: Markup.center.x + r * Math.sin(theta),
    y: Markup.center.y - r * Math.cos(theta),
  };
}

function edgeKey(e: Types.Edge): string {
  return e.from < e.to ? `${e.from}:${e.to}` : `${e.to}:${e.from}`;
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

function getTextWidth(name: string): number {
  return name.length * 2.2;
}

function mapToPercent(p: { x: number; y: number }) {
  const px = Markup.MAP_TX + p.x * Markup.MAP_SCALE;
  const py = Markup.MAP_TY + p.y * Markup.MAP_SCALE;
  return {
    left: `${(px / Markup.PAGE_W) * 100}%`,
    top: `${(py / Markup.PAGE_H) * 100}%`,
  };
}

function zenithToAirmass(zenithDeg: number): number {
  const z = (zenithDeg * Math.PI) / 180;
  return 1 / (Math.cos(z) + 0.025 * Math.exp(-11 * Math.cos(z)));
}

function extinctionMagnitude(
  zenithDeg: number,
  kMagPerAirmass: number,
): number {
  return zenithToAirmass(zenithDeg) * kMagPerAirmass;
}

export {
  dateFormatter,
  ringRadius,
  toXY,
  edgeXY,
  edgeKey,
  angularSeparation,
  getTextWidth,
  mapToPercent,
  zenithToAirmass,
  extinctionMagnitude
};
