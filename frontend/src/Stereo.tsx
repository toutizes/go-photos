import { useEffect, useRef, useState, type PointerEvent } from "react";
import type { Photo, Stereo } from "./model";
export type StereoMode =
  "original" | "parallel" | "crossed" | "animated" | "align";
export const stereoModes: { value: StereoMode; label: string }[] = [
  { value: "original", label: "Original" },
  { value: "parallel", label: "Parallèle" },
  { value: "crossed", label: "Vision croisée" },
  { value: "animated", label: "Alternance" },
  { value: "align", label: "Alignement" },
];
export function StereoImage({
  photo,
  url,
  mode,
  offset,
  onOffset,
}: {
  photo: Photo;
  url: string;
  mode: StereoMode;
  offset: Stereo;
  onOffset: (s: Stereo) => void;
}) {
  const ref = useRef<HTMLDivElement>(null),
    [size, setSize] = useState({ w: 1, h: 1 }),
    [left, setLeft] = useState(true);
  const drag = useRef<
    { x: number; y: number; dx: number; dy: number } | undefined
  >(undefined);
  useEffect(() => {
    const observer = new ResizeObserver(([e]) =>
      setSize({ w: e.contentRect.width, h: e.contentRect.height }),
    );
    observer.observe(ref.current!);
    return () => observer.disconnect();
  }, []);
  useEffect(() => {
    if (mode !== "animated") return;
    const timer = setInterval(() => setLeft((v) => !v), 300);
    return () => clearInterval(timer);
  }, [mode]);
  const columns = mode === "parallel" || mode === "crossed" ? 2 : 1;
  const scale = Math.min(size.w / ((photo.W / 2) * columns), size.h / photo.H),
    w = (photo.W / 2) * scale,
    h = photo.H * scale;
  function eye(isLeft: boolean, overlay = false) {
    return (
      <div
        className={`stereo-eye ${overlay ? "stereo-overlay" : ""}`}
        style={{ width: w, height: h }}
      >
        <img
          src={url}
          alt={isLeft ? "Vue gauche" : "Vue droite"}
          draggable={false}
          style={{
            width: w * 2,
            height: h,
            left: isLeft ? offset.Dx * scale : -w,
            top: isLeft ? offset.Dy * scale : 0,
          }}
        />
      </div>
    );
  }
  function down(e: PointerEvent) {
    if (mode !== "align") return;
    e.stopPropagation();
    e.currentTarget.setPointerCapture(e.pointerId);
    drag.current = { x: e.clientX, y: e.clientY, dx: offset.Dx, dy: offset.Dy };
  }
  function move(e: PointerEvent) {
    if (!drag.current) return;
    e.stopPropagation();
    onOffset({
      ...offset,
      Dx: drag.current.dx + (e.clientX - drag.current.x) / scale,
      Dy: drag.current.dy + (e.clientY - drag.current.y) / scale,
    });
  }
  return (
    <div
      className={`stereo-surface ${mode === "align" ? "aligning" : ""}`}
      ref={ref}
      onPointerDown={down}
      onPointerMove={move}
      onPointerUp={(e) => {
        if (drag.current) {
          e.stopPropagation();
          drag.current = undefined;
        }
      }}
      onPointerCancel={() => {
        drag.current = undefined;
      }}
    >
      <div className="stereo-pair" style={{ width: w * columns, height: h }}>
        {mode === "parallel" ? (
          <>
            {eye(true)}
            {eye(false)}
          </>
        ) : mode === "crossed" ? (
          <>
            {eye(false)}
            {eye(true)}
          </>
        ) : mode === "animated" ? (
          eye(left)
        ) : (
          <>
            {eye(false)}
            {eye(true, true)}
          </>
        )}
      </div>
    </div>
  );
}
