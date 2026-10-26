import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { PhotoImage } from "./images";
import { photoPath, type Photo } from "./model";
export function Grid({
  photos,
  open,
  hidden,
  activeId,
  header,
}: {
  photos: Photo[];
  open: (p: Photo) => void;
  hidden: boolean;
  activeId?: number;
  header?: ReactNode;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const [headerHeight, setHeaderHeight] = useState(0);
  useLayoutEffect(() => {
    const el = headerRef.current;
    if (!el) return;
    if (el.offsetHeight) setHeaderHeight(el.offsetHeight);
    const observer = new ResizeObserver(([entry]) => {
      const next = entry.contentRect.height;
      setHeaderHeight((prev) => (prev === next ? prev : next));
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [Boolean(header)]);
  const [size, setSize] = useState({ width: 1000, height: 700, minimum: 170 });
  const [top, setTop] = useState(0);
  useLayoutEffect(() => {
    const el = ref.current!;
    const readMinimum = () =>
      parseFloat(getComputedStyle(el).getPropertyValue("--photo-cell-min")) ||
      170;
    const updateSize = (width: number, height: number) => {
      if (!width) return;
      const minimum = readMinimum();
      setSize((prev) =>
        prev.width === width &&
        prev.height === height &&
        prev.minimum === minimum
          ? prev
          : { width, height, minimum },
      );
    };
    updateSize(el.clientWidth, el.clientHeight);
    const observer = new ResizeObserver(([e]) =>
      updateSize(e.contentRect.width, e.contentRect.height),
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);
  const columns = Math.max(2, Math.floor(size.width / size.minimum));
  const cell = (size.width - 8 * (columns - 1)) / columns;
  const step = cell + 8,
    rows = Math.ceil(photos.length / columns);
  const gridTop = Math.max(0, top - headerHeight);
  const start = Math.max(0, Math.floor(gridTop / step) - 2),
    end = Math.min(rows, Math.ceil((gridTop + size.height) / step) + 2);
  useLayoutEffect(() => {
    if (hidden || activeId === undefined || !ref.current) return;
    const i = photos.findIndex((p) => p.Id === activeId);
    if (i < 0) return;
    const rowTop = headerHeight + Math.floor(i / columns) * step,
      el = ref.current;
    if (rowTop < el.scrollTop || rowTop + cell > el.scrollTop + size.height)
      el.scrollTop = Math.max(0, rowTop - (size.height - cell) / 2);
    setTop(el.scrollTop);
  }, [hidden, activeId, columns, step, size.height, photos, headerHeight]);
  useEffect(() => {
    if (hidden || activeId === undefined) return;
    const frame = requestAnimationFrame(() =>
      ref.current
        ?.querySelector<HTMLButtonElement>(`[data-photo-id="${activeId}"]`)
        ?.focus({ preventScroll: true }),
    );
    return () => cancelAnimationFrame(frame);
  }, [hidden]);
  return (
    <div
      ref={ref}
      tabIndex={0}
      className="photo-grid scroll"
      style={{ display: hidden ? "none" : undefined }}
      onScroll={(e) => setTop(e.currentTarget.scrollTop)}
      aria-label="Grille de photos"
    >
      {header && <div ref={headerRef}>{header}</div>}
      <div style={{ height: rows * step, position: "relative" }}>
        {photos.slice(start * columns, end * columns).map((p, i) => {
          const index = start * columns + i;
          return (
            <button
              key={p.Id}
              data-photo-id={p.Id}
              className="photo-tile"
              style={{
                width: cell,
                height: cell,
                left: (index % columns) * (cell + 8),
                top: Math.floor(index / columns) * step,
              }}
              onClick={() => open(p)}
              aria-label={`Ouvrir ${p.In}`}
            >
              <PhotoImage path={photoPath(p, "mini")} alt={p.In} />
              <span className="tile-caption">{p.In}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
