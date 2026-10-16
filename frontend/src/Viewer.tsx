import { useEffect, useRef, useState, type PointerEvent } from "react";
import {
  ChevronLeft,
  ChevronRight,
  X,
  Download,
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
  onStep,
  onClose,
  onSearch,
}: {
  photos: Photo[];
  id: number;
  onStep: (id: number) => void;
  onClose: () => void;
  onSearch: (q: string) => void;
}) {
  const [controlsVisible, setControlsVisible] = useState(true);
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
      {...{ photo, onClose, onSearch }}
      mode={mode}
      setMode={setMode}
      controlsVisible={controlsVisible}
      onToggleControls={() => setControlsVisible((visible) => !visible)}
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
  index,
  count,
  previous,
  next,
  onStep,
  onClose,
  onSearch,
  mode,
  setMode,
  initialOffset,
  saveOffset,
  controlsVisible,
  onToggleControls,
}: {
  photo: Photo;
  controlsVisible: boolean;
  onToggleControls: () => void;
  mode: StereoMode;
  setMode: (m: StereoMode) => void;
  initialOffset?: Stereo;
  saveOffset: (s: Stereo) => void;
  index: number;
  count: number;
  previous?: Photo;
  next?: Photo;
  onStep: (id: number) => void;
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
  const edgePull = useRef(0);
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
    transform.scale > 1.15 && !stereoActive
      ? photoPath(photo, "maxi")
      : undefined,
  );
  const keywords = useRef<HTMLDivElement>(null);
  const [keywordsExpanded, setKeywordsExpanded] = useState(false);
  const [keywordHeight, setKeywordHeight] = useState<number | null>(null);
  const keywordDrag = useRef<{ y: number; height: number } | null>(null);
  const keywordDragged = useRef(false);
  function keywordLimit() {
    return Math.max(
      70,
      Math.min(keywords.current?.scrollHeight ?? 70, window.innerHeight * 0.55),
    );
  }
  function startKeywords(e: PointerEvent) {
    if (e.button !== 0 || e.isPrimary === false) return;
    keywordDragged.current = false;
    keywordDrag.current = {
      y: e.clientY,
      height: keywords.current?.getBoundingClientRect().height ?? 70,
    };
  }
  function moveKeywords(e: PointerEvent) {
    const start = keywordDrag.current;
    if (!start) return;
    const delta = start.y - e.clientY;
    if (Math.abs(delta) < 6 && !keywordDragged.current) return;
    keywordDragged.current = true;
    e.currentTarget.setPointerCapture(e.pointerId);
    setKeywordHeight(
      Math.max(70, Math.min(keywordLimit(), start.height + delta)),
    );
  }
  function finishKeywords(e: PointerEvent) {
    const start = keywordDrag.current;
    if (!start) return;
    if (keywordDragged.current) setKeywordsExpanded(start.y - e.clientY > 0);
    keywordDrag.current = null;
    setKeywordHeight(null);
  }

  const current = useRef(transform);
  current.current = transform;
  const points = useRef(new Map<number, Point>()),
    start = useRef<Point | undefined>(undefined),
    moved = useRef(false),
    pinched = useRef(false);
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const viewerFocus = useRef<HTMLElement>(null);
  const navigate = (direction: number) => {
    const p = direction < 0 ? previous : next;
    if (p) settleSlide(direction);
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
    const obs = new ResizeObserver(() => setTransform((t) => bounds(t)));
    if (stage.current) obs.observe(stage.current);
    return () => obs.disconnect();
  }, [mode, stereoActive, natural.width, natural.height]);
  useEffect(() => {
    viewerFocus.current?.focus({ preventScroll: true });
  }, []);
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
    if (e.button !== 0 || slideLocked.current) return;
    e.currentTarget.setPointerCapture(e.pointerId);
    const p = position(e);
    points.current.set(e.pointerId, p);
    if (points.current.size === 1) {
      edgePull.current = 0;
      start.current = p;
      moved.current = false;
      pinched.current = false;
    } else {
      pinched.current = true;
      edgePull.current = 0;
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
    if (before.length === 2) {
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
    } else if (current.current.scale > 1) {
      const t = current.current;
      const proposed = {
        ...t,
        x: t.x + p.x - old.x + edgePull.current,
        y: t.y + p.y - old.y,
      };
      const bounded = bounds(proposed);
      current.current = bounded;
      setTransform(bounded);
      const horizontal =
        start.current &&
        Math.abs(p.x - start.current.x) > Math.abs(p.y - start.current.y) * 1.4;
      edgePull.current =
        !pinched.current && horizontal ? proposed.x - bounded.x : 0;
      const hasNeighbor = edgePull.current < 0 ? next : previous;
      setSlide(hasNeighbor ? edgePull.current : edgePull.current * 0.2);
    } else if (!pinched.current && start.current && before.length === 1) {
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
      (current.current.scale <= 1.01
        ? Math.abs(p.x - origin.x) > 55
        : Math.abs(edgePull.current) > 55) &&
      Math.abs(p.x - origin.x) > Math.abs(p.y - origin.y) * 1.4
    ) {
      const drag =
        current.current.scale <= 1.01 ? p.x - origin.x : edgePull.current;
      settleSlide(drag < 0 ? 1 : -1);
      return;
    }
    if (slide !== 0) settleSlide();
    if (!moved.current) onToggleControls();
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
      ref={viewerFocus}
      tabIndex={-1}
      className="viewer fullscreen"
      aria-label="Visionneuse de photos"
    >
      <header
        className={`viewer-toolbar ${controlsVisible ? "" : "controls-hidden"}`}
      >
        <button
          className="icon-button"
          onClick={onClose}
          aria-label="Revenir à la grille"
        >
          <X />
        </button>
        {controlsVisible && (
          <>
            <span className="counter" aria-live="polite">
              {index + 1} <span>/ {count}</span>
            </span>
            <div className="toolbar-spacer" />
            <button
              className="icon-button"
              onClick={save}
              disabled={busy}
              aria-label={busy ? "Téléchargement…" : "Télécharger l’original"}
              title="Télécharger l’original"
              aria-busy={busy}
            >
              <Download size={20} />
            </button>
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
          </>
        )}
      </header>
      {error && controlsVisible && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <div className="viewer-body">
        <div
          className="image-stage"
          ref={stage}
          onPointerDown={down}
          onPointerMove={move}
          onPointerUp={up}
          onPointerCancel={() => {
            points.current.clear();
            edgePull.current = 0;
            pinched.current = true;
            settleSlide();
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
            className={`step previous ${controlsVisible ? "" : "is-hidden"}`}
            aria-hidden={!controlsVisible}
            tabIndex={controlsVisible ? 0 : -1}
            disabled={!previous}
            aria-label="Photo précédente"
            onPointerDown={(e) => e.stopPropagation()}
            onDoubleClick={(e) => e.stopPropagation()}
            onClick={() => navigate(-1)}
          >
            <ChevronLeft />
          </button>
          <button
            className={`step next ${controlsVisible ? "" : "is-hidden"}`}
            aria-hidden={!controlsVisible}
            tabIndex={controlsVisible ? 0 : -1}
            disabled={!next}
            aria-label="Photo suivante"
            onPointerDown={(e) => e.stopPropagation()}
            onDoubleClick={(e) => e.stopPropagation()}
            onClick={() => navigate(1)}
          >
            <ChevronRight />
          </button>
        </div>
        <aside
          className={`photo-metadata ${controlsVisible ? "" : "is-hidden"} ${keywordsExpanded ? "keywords-expanded" : ""}`}
          aria-hidden={!controlsVisible}
          inert={!controlsVisible}
          aria-label="Informations de la photo"
        >
          <button
            className="keyword-handle"
            aria-label={
              keywordsExpanded
                ? "Réduire les mots-clés"
                : "Développer les mots-clés"
            }
            aria-expanded={keywordsExpanded}
            onPointerDown={startKeywords}
            onPointerMove={moveKeywords}
            onPointerUp={finishKeywords}
            onPointerCancel={() => {
              keywordDrag.current = null;
              setKeywordHeight(null);
            }}
            onClick={() => {
              if (!keywordDragged.current)
                setKeywordsExpanded((expanded) => !expanded);
              keywordDragged.current = false;
            }}
          >
            <span />
          </button>
          <div className="photo-caption">
            <button
              className="eyebrow date-link"
              onClick={() => onSearch(`${date.getFullYear()}-${monthDay}`)}
            >
              {dateLabel(photo.Its).toLocaleUpperCase("fr")}
            </button>
            <span className="photo-filename" title={photo.In}>
              {photo.In}
            </span>
          </div>
          <div
            className="chips photo-keywords"
            ref={keywords}
            style={{
              maxHeight: keywordHeight ?? (keywordsExpanded ? "55dvh" : 70),
              transition: keywordHeight === null ? undefined : "none",
            }}
            onPointerDown={keywordsExpanded ? undefined : startKeywords}
            onPointerMove={moveKeywords}
            onPointerUp={finishKeywords}
            onPointerCancel={() => {
              keywordDrag.current = null;
              setKeywordHeight(null);
            }}
            onClickCapture={(e) => {
              if (keywordDragged.current) {
                e.preventDefault();
                e.stopPropagation();
                keywordDragged.current = false;
              }
            }}
            tabIndex={0}
            role="region"
            aria-label="Mots-clés"
          >
            {(photo.Kwd ?? []).map((k) => (
              <button key={k} onClick={() => onSearch(k)}>
                {k}
              </button>
            ))}
          </div>
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
        </aside>
      </div>
    </section>
  );
}
