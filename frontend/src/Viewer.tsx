import { useEffect, useRef, useState, type PointerEvent } from "react";
import {
  ChevronLeft,
  ChevronRight,
  X,
  Maximize,
  Download,
  RotateCcw,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import { useImage } from "./images";
import {
  dateLabel,
  photoPath,
  stereoOf,
  type Photo,
  type Stereo,
} from "./model";
import { download } from "./api";
import { StereoImage, stereoModes, type StereoMode } from "./Stereo";
import {
  constrain,
  distance,
  identity,
  midpoint,
  pinch,
  type Point,
  type Transform,
} from "./gestures";
export function Viewer({
  photos,
  id,
  full,
  onStep,
  onFull,
  onClose,
  onSearch,
}: {
  photos: Photo[];
  id: number;
  full: boolean;
  onStep: (id: number) => void;
  onFull: () => void;
  onClose: () => void;
  onSearch: (q: string) => void;
}) {
  const [mode, setMode] = useState<StereoMode>("parallel");
  const offsets = useRef(new Map<number, Stereo>());
  const index = photos.findIndex((p) => p.Id === id),
    photo = photos[index];
  if (!photo)
    return (
      <section className="empty">
        <h2>Photo introuvable</h2>
        <p>Cette photo ne fait plus partie des résultats.</p>
        <button onClick={onClose}>Revenir aux photos</button>
      </section>
    );
  return (
    <ViewerPhoto
      key={photo.Id}
      {...{ photo, full, onFull, onClose, onSearch }}
      mode={mode}
      setMode={setMode}
      initialOffset={offsets.current.get(photo.Id) ?? stereoOf(photo)}
      saveOffset={(s) => offsets.current.set(photo.Id, s)}
      count={photos.length}
      index={index}
      previous={photos[index - 1]}
      next={photos[index + 1]}
      onStep={onStep}
    />
  );
}
function ViewerPhoto({
  photo,
  full,
  index,
  count,
  previous,
  next,
  onStep,
  onFull,
  onClose,
  onSearch,
  mode,
  setMode,
  initialOffset,
  saveOffset,
}: {
  photo: Photo;
  mode: StereoMode;
  setMode: (m: StereoMode) => void;
  initialOffset?: Stereo;
  saveOffset: (s: Stereo) => void;
  full: boolean;
  index: number;
  count: number;
  previous?: Photo;
  next?: Photo;
  onStep: (id: number) => void;
  onFull: () => void;
  onClose: () => void;
  onSearch: (q: string) => void;
}) {
  const date = new Date(photo.Its * 1000);
  const monthDay = `${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  const image = useImage(photoPath(photo));
  const previousImage = useImage(previous ? photoPath(previous) : undefined);
  const nextImage = useImage(next ? photoPath(next) : undefined);
  const stage = useRef<HTMLDivElement>(null);
  const [transform, setTransform] = useState(identity);
  const [slide, setSlide] = useState(0);
  const [settling, setSettling] = useState(false);
  const slideLocked = useRef(false);
  const slideTimer = useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined,
  );
  useEffect(() => () => clearTimeout(slideTimer.current), []);
  function settleSlide(direction = 0) {
    if (slideLocked.current) return;
    const target = direction < 0 ? previous : direction > 0 ? next : undefined;
    const reducedMotion = window.matchMedia?.(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    if (reducedMotion) {
      setSlide(0);
      if (target) onStep(target.Id);
      return;
    }
    slideLocked.current = true;
    setSettling(true);
    setSlide(
      target
        ? -direction * (stage.current?.getBoundingClientRect().width || 0)
        : 0,
    );
    slideTimer.current = setTimeout(() => {
      slideLocked.current = false;
      setSettling(false);
      setSlide(0);
      if (target) onStep(target.Id);
    }, 220);
  }
  const [offset, setOffset] = useState(initialOffset);
  const [natural, setNatural] = useState({
    width: Math.max(1, photo.W),
    height: Math.max(1, photo.H),
  });
  const stereoActive = !!offset && mode !== "original";
  const original = useImage(
    full && transform.scale > 1.15 && !stereoActive
      ? photoPath(photo, "maxi")
      : undefined,
  );
  const pointerType = useRef("mouse");
  const current = useRef(transform);
  current.current = transform;
  const points = useRef(new Map<number, Point>()),
    start = useRef<Point | undefined>(undefined),
    moved = useRef(false),
    pinched = useRef(false),
    lastTap = useRef(0);
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const controls = useRef<HTMLButtonElement>(null);
  const navigate = (direction: number) => {
    const p = direction < 0 ? previous : next;
    if (p && !slideLocked.current) onStep(p.Id);
  };
  const bounds = (t: Transform) => {
    const r = stage.current!.getBoundingClientRect();
    return constrain(
      t,
      r.width,
      r.height,
      stereoActive && (mode === "animated" || mode === "align")
        ? photo.W / 2
        : stereoActive
          ? photo.W
          : natural.width,
      stereoActive ? photo.H : natural.height,
    );
  };
  useEffect(() => {
    setTransform(identity);
    setSlide(0);
    setSettling(false);
    slideLocked.current = false;
    clearTimeout(slideTimer.current);
    points.current.clear();
  }, [full]);
  useEffect(() => {
    const obs = new ResizeObserver(() => setTransform((t) => bounds(t)));
    if (stage.current) obs.observe(stage.current);
    return () => obs.disconnect();
  }, [mode, stereoActive, natural.width, natural.height]);
  useEffect(() => {
    controls.current?.focus({ preventScroll: true });
  }, [full]);
  useEffect(() => {
    const handle = (e: KeyboardEvent) => {
      if ((e.target as HTMLElement).matches("input,textarea,select")) return;
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
      }
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        navigate(-1);
      }
      if (e.key === "ArrowRight") {
        e.preventDefault();
        navigate(1);
      }
    };
    window.addEventListener("keydown", handle);
    return () => window.removeEventListener("keydown", handle);
  });
  function position(e: PointerEvent): Point {
    const r = stage.current!.getBoundingClientRect();
    return {
      x: e.clientX - r.left - r.width / 2,
      y: e.clientY - r.top - r.height / 2,
    };
  }
  function down(e: PointerEvent) {
    pointerType.current = e.pointerType;
    if (e.button !== 0 || slideLocked.current) return;
    e.currentTarget.setPointerCapture(e.pointerId);
    const p = position(e);
    points.current.set(e.pointerId, p);
    if (points.current.size === 1) {
      start.current = p;
      moved.current = false;
      pinched.current = false;
    } else {
      pinched.current = true;
      setSlide(0);
    }
  }
  function move(e: PointerEvent) {
    const old = points.current.get(e.pointerId);
    if (!old) return;
    const p = position(e);
    if (start.current && distance(start.current, p) > 8) moved.current = true;
    const before = [...points.current.values()];
    points.current.set(e.pointerId, p);
    const after = [...points.current.values()];
    if (full && before.length === 2) {
      const ratio =
        distance(after[0], after[1]) /
        Math.max(1, distance(before[0], before[1]));
      setTransform((t) =>
        bounds(
          pinch(
            t,
            midpoint(before[0], before[1]),
            midpoint(after[0], after[1]),
            ratio,
          ),
        ),
      );
    } else if (full && current.current.scale > 1)
      setTransform((t) =>
        bounds({ ...t, x: t.x + p.x - old.x, y: t.y + p.y - old.y }),
      );
    else if (!pinched.current && start.current && before.length === 1) {
      const dx = p.x - start.current.x;
      const dy = p.y - start.current.y;
      if (Math.abs(dx) > Math.abs(dy) * 1.4) {
        const hasNeighbor = dx < 0 ? next : previous;
        setSlide(hasNeighbor ? dx : dx * 0.2);
      }
    }
  }
  function up(e: PointerEvent) {
    if (!points.current.has(e.pointerId)) return;
    const p = position(e);
    points.current.delete(e.pointerId);
    if (points.current.size || pinched.current) return;
    const origin = start.current;
    if (!origin) return;
    if (
      current.current.scale <= 1.01 &&
      Math.abs(p.x - origin.x) > 55 &&
      Math.abs(p.x - origin.x) > Math.abs(p.y - origin.y) * 1.4
    ) {
      settleSlide(p.x < origin.x ? 1 : -1);
      return;
    }
    if (slide !== 0) settleSlide();
    if (!moved.current && e.pointerType === "touch") {
      const now = Date.now();
      if (now - lastTap.current < 300) {
        lastTap.current = 0;
        if (full)
          setTransform((t) =>
            t.scale > 1 ? identity : bounds({ ...t, scale: 2 }),
          );
        else onFull();
      } else lastTap.current = now;
    }
  }
  async function save() {
    setBusy(true);
    setError("");
    try {
      await download(photoPath(photo, "maxi"), photo.In);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  return (
    <section
      className={`viewer ${full ? "fullscreen" : ""}`}
      aria-label="Visionneuse de photos"
    >
      <header className="viewer-toolbar">
        <button
          ref={controls}
          className="icon-button"
          onClick={onClose}
          aria-label={full ? "Quitter le plein écran" : "Revenir à la grille"}
        >
          <X />
        </button>
        <span className="counter" aria-live="polite">
          {index + 1} <span>/ {count}</span>
        </span>
        <div className="toolbar-spacer" />
        {offset && (
          <select
            className="stereo-select"
            aria-label="Mode stéréo"
            value={mode}
            onChange={(e) => {
              setMode(e.target.value as StereoMode);
              setTransform(identity);
            }}
          >
            {stereoModes.map((m) => (
              <option value={m.value} key={m.value}>
                {m.label}
              </option>
            ))}
          </select>
        )}
        {full ? (
          <>
            <button
              className="icon-button"
              aria-label="Réduire"
              disabled={stereoActive && mode === "align"}
              onClick={() =>
                setTransform((t) => bounds({ ...t, scale: t.scale / 1.5 }))
              }
            >
              <ZoomOut />
            </button>
            <button
              className="icon-button"
              aria-label="Agrandir"
              disabled={stereoActive && mode === "align"}
              onClick={() =>
                setTransform((t) => bounds({ ...t, scale: t.scale * 1.5 }))
              }
            >
              <ZoomIn />
            </button>
            <button
              className="icon-button"
              aria-label="Réinitialiser le zoom"
              onClick={() => setTransform(identity)}
            >
              <RotateCcw />
            </button>
          </>
        ) : (
          <button
            className="icon-button"
            onClick={onFull}
            aria-label="Afficher en plein écran"
          >
            <Maximize />
          </button>
        )}
      </header>
      <div className="viewer-body">
        <div
          className="image-stage"
          ref={stage}
          onPointerDown={down}
          onPointerMove={move}
          onPointerUp={up}
          onPointerCancel={() => {
            points.current.clear();
            pinched.current = true;
            settleSlide();
          }}
          onDoubleClick={(e) => {
            if (pointerType.current !== "touch" && !full) onFull();
          }}
        >
          <div
            className={`photo-slider ${settling ? "is-settling" : ""}`}
            style={{ transform: `translateX(${slide}px)` }}
          >
            <div className="photo-slide">
              {image.url && stereoActive ? (
                <div
                  className="stereo-transform"
                  style={{
                    transform: `translate(${transform.x}px,${transform.y}px) scale(${transform.scale})`,
                  }}
                >
                  <StereoImage
                    photo={photo}
                    url={image.url}
                    mode={mode}
                    offset={offset!}
                    onOffset={(s) => {
                      setOffset(s);
                      saveOffset(s);
                    }}
                  />
                </div>
              ) : image.url ? (
                <img
                  className="viewer-image"
                  onLoad={(e) => {
                    const img = e.currentTarget;
                    if (img.naturalWidth && img.naturalHeight)
                      setNatural({
                        width: img.naturalWidth,
                        height: img.naturalHeight,
                      });
                  }}
                  src={original.url ?? image.url}
                  alt={photo.In}
                  draggable={false}
                  style={{
                    transform: `translate(${transform.x}px,${transform.y}px) scale(${transform.scale})`,
                  }}
                />
              ) : image.error ? (
                <button
                  onPointerDown={(e) => e.stopPropagation()}
                  onClick={image.retry}
                >
                  Réessayer le chargement
                </button>
              ) : (
                <span className="loading">Chargement de la photo…</span>
              )}
            </div>
            {slide !== 0 &&
              [
                { neighbor: previous, resource: previousImage, side: -1 },
                { neighbor: next, resource: nextImage, side: 1 },
              ].map(
                ({ neighbor, resource, side }) =>
                  neighbor && (
                    <div
                      key={side}
                      className="photo-slide neighbor-slide"
                      aria-hidden="true"
                      style={{ transform: `translateX(${side * 100}%)` }}
                    >
                      {resource.url ? (
                        stereoOf(neighbor) && mode !== "original" ? (
                          <StereoImage
                            photo={neighbor}
                            url={resource.url}
                            mode={mode}
                            offset={stereoOf(neighbor)!}
                            onOffset={() => {}}
                          />
                        ) : (
                          <img
                            className="viewer-image"
                            src={resource.url}
                            alt=""
                            draggable={false}
                          />
                        )
                      ) : (
                        <span className="loading">Chargement de la photo…</span>
                      )}
                    </div>
                  ),
              )}
          </div>
          <button
            className="step previous"
            disabled={!previous}
            aria-label="Photo précédente"
            onPointerDown={(e) => e.stopPropagation()}
            onDoubleClick={(e) => e.stopPropagation()}
            onClick={() => navigate(-1)}
          >
            <ChevronLeft />
          </button>
          <button
            className="step next"
            disabled={!next}
            aria-label="Photo suivante"
            onPointerDown={(e) => e.stopPropagation()}
            onDoubleClick={(e) => e.stopPropagation()}
            onClick={() => navigate(1)}
          >
            <ChevronRight />
          </button>
        </div>
        {!full && (
          <aside className="metadata scroll">
            <button
              className="eyebrow date-link"
              onClick={() => onSearch(`${date.getFullYear()}-${monthDay}`)}
            >
              LE {dateLabel(photo.Its).toLocaleUpperCase("fr")}
            </button>
            <h2>{photo.In}</h2>
            <button
              className="album-link"
              onClick={() => onSearch(`album:${photo.Ad}`)}
            >
              {photo.Ad}
            </button>
            <div className="chips">
              {(photo.Kwd ?? []).map((k) => (
                <button key={k} onClick={() => onSearch(k)}>
                  {k}
                </button>
              ))}
            </div>
            <button
              className="text-button"
              onClick={() =>
                onSearch(new Date(photo.Its * 1000).getFullYear().toString())
              }
            >
              Voir les photos de {new Date(photo.Its * 1000).getFullYear()}
            </button>
            <button className="text-button" onClick={() => onSearch(monthDay)}>
              Ce jour-là, au fil des années
            </button>
            {offset && mode === "align" && (
              <div className="alignment-controls">
                <p className="hint">
                  Glissez la vue gauche pour aligner les images. Ajustements
                  temporaires.
                </p>
                <label>
                  Horizontal
                  <input
                    type="range"
                    aria-label="Alignement horizontal"
                    min={-Math.floor(photo.W / 4)}
                    max={Math.floor(photo.W / 4)}
                    step={1}
                    value={offset.Dx}
                    onChange={(e) => {
                      const s = { ...offset, Dx: Number(e.target.value) };
                      setOffset(s);
                      saveOffset(s);
                    }}
                  />
                </label>
                <label>
                  Vertical
                  <input
                    type="range"
                    aria-label="Alignement vertical"
                    min={-Math.floor(photo.H / 4)}
                    max={Math.floor(photo.H / 4)}
                    step={1}
                    value={offset.Dy}
                    onChange={(e) => {
                      const s = { ...offset, Dy: Number(e.target.value) };
                      setOffset(s);
                      saveOffset(s);
                    }}
                  />
                </label>
                <button
                  onClick={() => {
                    const s = stereoOf(photo)!;
                    setOffset(s);
                    saveOffset(s);
                  }}
                >
                  Réinitialiser l’alignement
                </button>
              </div>
            )}
            <div className="metadata-footer">
              <button onClick={save} disabled={busy}>
                <Download size={17} />
                {busy ? "Téléchargement…" : "Télécharger l’original"}
              </button>
              <p className="hint">
                Double-touchez la photo pour le plein écran.
              </p>
              {error && <p role="alert">{error}</p>}
            </div>
          </aside>
        )}
      </div>
    </section>
  );
}
