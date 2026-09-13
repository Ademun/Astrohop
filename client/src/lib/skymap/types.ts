interface Point {
  x: number;
  y: number;
  oid: number;
  name: string;
  alt: number;
  az: number;
}

interface Edge {
  from: number;
  to: number;
}

interface Segment {
  key: string;
  from: Point;
  to: Point;
  distance: number;
}

export type {Point, Edge, Segment}