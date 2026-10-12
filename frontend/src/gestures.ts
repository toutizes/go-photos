export type Point = { x: number; y: number };
export type Transform = { scale: number; x: number; y: number };
export const identity: Transform = { scale: 1, x: 0, y: 0 };
export function constrain(
  t: Transform,
  width: number,
  height: number,
  imageWidth: number,
  imageHeight: number,
): Transform {
  const scale = Math.max(1, Math.min(5, t.scale));
  const fit = Math.min(width / imageWidth, height / imageHeight);
  const mx = Math.max(0, (imageWidth * fit * scale - width) / 2),
    my = Math.max(0, (imageHeight * fit * scale - height) / 2);
  return {
    scale,
    x: Math.max(-mx, Math.min(mx, t.x)) || 0,
    y: Math.max(-my, Math.min(my, t.y)) || 0,
  };
}
export const distance = (a: Point, b: Point) =>
  Math.hypot(a.x - b.x, a.y - b.y);
export const midpoint = (a: Point, b: Point): Point => ({
  x: (a.x + b.x) / 2,
  y: (a.y + b.y) / 2,
});
export function pinch(
  before: Transform,
  oldCenter: Point,
  newCenter: Point,
  ratio: number,
): Transform {
  const scale = Math.max(1, Math.min(5, before.scale * ratio));
  const actual = scale / before.scale;
  return {
    scale,
    x: newCenter.x - (oldCenter.x - before.x) * actual,
    y: newCenter.y - (oldCenter.y - before.y) * actual,
  };
}
