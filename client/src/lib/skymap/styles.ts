const STYLE_BASE = `
  @page { size: A4 landscape; margin: 0; }
  html, body { margin: 0; padding: 0; }
  svg { display: block; }

  .page-frame { fill: none; stroke: var(--map-ink); stroke-width: 1.6; }
  .divider { stroke: var(--map-ink); stroke-width: 0.3; }

  .grid-dashed circle,
  .grid-dashed line {
    fill: none;
    stroke: var(--map-grid);
    stroke-width: calc(0.15px / var(--map-scale, 1));
    stroke-dasharray:
      calc(0.6px / var(--map-scale, 1)),
      calc(0.8px / var(--map-scale, 1));
  }

  .axis-line {
    stroke: var(--map-ink);
    stroke-width: calc(0.2px / var(--map-scale, 1));
  }
  .horizon-circle {
    fill: none;
    stroke: var(--map-ink);
    stroke-width: calc(0.5px / var(--map-scale, 1));
  }
  .horizon-tick {
    stroke: var(--map-ink);
    stroke-width: calc(0.5px / var(--map-scale, 1));
  }
  .center-dot { fill: var(--map-ink); }

  .cardinal-label {
    font: 700 calc(4px / var(--map-scale, 1)) 'MuseoModerno', monospace;
    fill: var(--map-ink);
    text-anchor: middle;
    dominant-baseline: middle;
  }

  .route-line {
    fill: none;
    stroke: var(--map-ink);
    stroke-width: calc(0.3px / var(--map-scale, 1));
    stroke-dasharray:
      calc(1.3px / var(--map-scale, 1)),
      calc(1px / var(--map-scale, 1));
  }

  .star-core { fill: var(--map-ink); stroke: none; }
  .star-core.is-invisible {
    fill: none;
    stroke: var(--map-ink);
    stroke-width: calc(0.4px / var(--map-scale, 1));
  }

  .star-label {
    font: 400 calc(2.8px / var(--map-scale, 1)) 'LINE Seed JP', 'Inter', monospace;
    fill: var(--map-ink);
    text-anchor: middle;
    paint-order: stroke;
    stroke-linejoin: round;
  }

  .moon-outline {
    fill: none;
    stroke: var(--map-ink);
    stroke-width: calc(0.4px / var(--map-scale, 1));
  }
  .moon-fill { fill: var(--map-paper); }

  .sb-heading { font: 700 4px 'MuseoModerno', monospace; fill: var(--map-ink); }
  .sb-body    { font: 400 3.4px 'LINE Seed JP', 'Inter', monospace; fill: var(--map-ink-soft); }
  .sb-note    { font: italic 400 3px 'LINE Seed JP', 'Inter', monospace; fill: var(--map-muted); }

  .logo-text {
    font: 700 6px 'MuseoModerno', monospace;
    fill: var(--map-ink);
    letter-spacing: 0.5px;
  }

  .alt-scale-line {
    stroke: var(--map-ink);
    stroke-width: calc(0.4px / var(--map-scale, 1));
  }
  .alt-scale-tick {
    stroke: var(--map-ink);
    stroke-width: calc(0.4px / var(--map-scale, 1));
  }
  .alt-scale-label {
    font: 400 calc(3px / var(--map-scale, 1)) 'LINE Seed JP', 'Inter', monospace;
    fill: var(--map-ink);
    text-anchor: middle;
    dominant-baseline: hanging;
  }
    .extinction-chart-frame {
  fill: none;
  stroke: var(--map-ink);
  stroke-width: 0.3;
}
.extinction-chart-line {
  fill: none;
  stroke: var(--map-ink);
  stroke-width: 0.5;
}
  .extinction-chart-tick {
  stroke: var(--map-ink);
  stroke-width: 0.3;
}
.extinction-axis-label {
  font: 400 2.6px 'LINE Seed JP', 'Inter', monospace;
  fill: var(--map-ink);
  text-anchor: end;
  dominant-baseline: middle;
}
`;

const STYLE_GRAYSCALE = `
  .theme {
    --map-ink: #000;
    --map-ink-soft: #1a1a1a;
    --map-muted: #666;
    --map-grid: #999;
    --map-paper: #fff;
  }
`;

const STYLE_RED = `
  .theme {
    --map-ink: #c00;
    --map-ink-soft: #7a1010;
    --map-muted: #a33;
    --map-grid: #e0a0a0;
    --map-paper: #fff;
  }
`;

const STYLE_INTERACTIVE = `
  .route-hit {
    fill: none;
    stroke: transparent;
    stroke-width: calc(4px / var(--map-scale, 1));
    pointer-events: stroke;
  }
  .route-hit.is-cuttable { cursor: pointer; }

  .route-line.is-active {
    stroke: #b8472d;
    stroke-width: calc(1px / var(--map-scale, 1));
  }

  .measure-line {
    stroke: #b8472d;
    stroke-width: calc(0.6px / var(--map-scale, 1));
    stroke-dasharray:
      calc(0.8px / var(--map-scale, 1)),
      calc(0.8px / var(--map-scale, 1));
  }
  .measure-label {
    font: 700 calc(4px / var(--map-scale, 1)) 'LINE Seed JP', 'Inter', monospace;
    fill: #b8472d;
    text-anchor: middle;
    paint-order: stroke;
    stroke: var(--map-paper);
    stroke-width: calc(1px / var(--map-scale, 1));
  }

  .star-core { transition: fill 0.1s ease; }
  .star-core.is-hovered  { fill: #b8472d; }
  .star-core.is-pending  { fill: #b8472d; }
  .star-core.is-invisible.is-hovered { stroke: #b8472d; }

  .star-halo {
    fill: url(#starGlow);
    opacity: 0.55;
    transition: opacity 0.1s ease;
  }
  .star-halo.is-hovered { opacity: 1; }

  .label-backplate {
    fill: var(--map-paper);
  }
  .label-backplate.is-hovered { fill: #b8472d; }

  .star-label.is-hovered { fill: #b8472d; font-weight: 700; }
`;

export { STYLE_BASE, STYLE_GRAYSCALE, STYLE_RED, STYLE_INTERACTIVE };
